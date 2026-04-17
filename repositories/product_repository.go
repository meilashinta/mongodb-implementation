package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dualwrite/product-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductRepository manages dual-write operations across MySQL and MongoDB.
type ProductRepository struct {
	sqlDB           *sql.DB
	mongoCollection *mongo.Collection
}

// NewProductRepository returns a repository instance with pooled connections.
func NewProductRepository(sqlDB *sql.DB, mongoCollection *mongo.Collection) *ProductRepository {
	return &ProductRepository{sqlDB: sqlDB, mongoCollection: mongoCollection}
}

// CreateProduct writes a new product to MySQL first, then replicates to MongoDB.
func (r *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	if product.ID == "" {
		product.ID = generateProductID()
	}
	currentTime := time.Now().UTC()
	product.CreatedAt = currentTime
	product.UpdatedAt = currentTime

	// Use a SQL transaction to make the MySQL write reliable.
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sql transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO products (id, name, description, category, price, stock, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		product.ID,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
		product.Stock,
		strings.Join(product.Tags, ","),
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert mysql product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit mysql insert: %w", err)
	}

	// Replicate the same product to MongoDB after MySQL succeeds.
	if _, err = r.mongoCollection.InsertOne(ctx, product); err != nil {
		log.Printf("dual-write inconsistency: product %s saved to MySQL but failed to insert into MongoDB: %v", product.ID, err)
		return fmt.Errorf("insert mongodb product: %w", err)
	}

	return nil
}

// GetAllProducts reads all products from MongoDB for read-optimized queries.
func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	cursor, err := r.mongoCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find all products mongodb: %w", err)
	}
	defer cursor.Close(ctx)

	var products []models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, fmt.Errorf("decode mongodb product: %w", err)
		}
		products = append(products, product)
	}

	if products == nil {
		products = []models.Product{}
	}

	return products, nil
}

// GetProductByID reads a single product from MongoDB using the provided ID.
func (r *ProductRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	var product models.Product
	if err := r.mongoCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("find product by id mongodb: %w", err)
	}
	return &product, nil
}

// SearchProducts filters products in MongoDB by category, price range, and tags.
func (r *ProductRepository) SearchProducts(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	query := bson.M{}
	if filter.Category != "" {
		query["category"] = filter.Category
	}

	priceFilter := bson.M{}
	if filter.MinPrice > 0 {
		priceFilter["$gte"] = filter.MinPrice
	}
	if filter.MaxPrice > 0 {
		priceFilter["$lte"] = filter.MaxPrice
	}
	if len(priceFilter) > 0 {
		query["price"] = priceFilter
	}

	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$all": filter.Tags}
	}

	cursor, err := r.mongoCollection.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search products mongodb: %w", err)
	}
	defer cursor.Close(ctx)

	var products []models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, fmt.Errorf("decode search result: %w", err)
		}
		products = append(products, product)
	}

	if products == nil {
		products = []models.Product{}
	}

	return products, nil
}

// UpdateProduct updates MySQL first, then applies the same change to MongoDB.
func (r *ProductRepository) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	product.UpdatedAt = time.Now().UTC()

	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sql transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE products SET name = ?, description = ?, category = ?, price = ?, stock = ?, tags = ?, updated_at = ? WHERE id = ?`,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
		product.Stock,
		strings.Join(product.Tags, ","),
		product.UpdatedAt,
		id,
	)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("update mysql product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit mysql update: %w", err)
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"name":        product.Name,
		"description": product.Description,
		"category":    product.Category,
		"price":       product.Price,
		"stock":       product.Stock,
		"tags":        product.Tags,
		"updated_at":  product.UpdatedAt,
	}}
	opts := options.Update().SetUpsert(true)
	if _, err := r.mongoCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		log.Printf("dual-write inconsistency: product %s saved to MySQL but failed to update MongoDB: %v", id, err)
		return fmt.Errorf("update mongodb product: %w", err)
	}

	return nil
}

// DeleteProduct removes a product from MySQL and then from MongoDB.
func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin sql transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM products WHERE id = ?`, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete mysql product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit mysql delete: %w", err)
	}

	if _, err := r.mongoCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		log.Printf("dual-write inconsistency: product %s deleted from MySQL but failed to delete from MongoDB: %v", id, err)
		return fmt.Errorf("delete mongodb product: %w", err)
	}

	return nil
}

func generateProductID() string {
	return fmt.Sprintf("prod-%d", time.Now().UTC().UnixNano())
}
