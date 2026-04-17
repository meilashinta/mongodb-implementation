package services

import (
	"context"

	"github.com/dualwrite/product-api/models"
	"github.com/dualwrite/product-api/repositories"
)

// ProductService orchestrates business logic for product CRUD operations.
type ProductService struct {
	repository *repositories.ProductRepository
}

// NewProductService returns a new service instance.
func NewProductService(repository *repositories.ProductRepository) *ProductService {
	return &ProductService{repository: repository}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	return s.repository.CreateProduct(ctx, product)
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	return s.repository.GetAllProducts(ctx)
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	return s.repository.GetProductByID(ctx, id)
}

func (s *ProductService) SearchProducts(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	return s.repository.SearchProducts(ctx, filter)
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	return s.repository.UpdateProduct(ctx, id, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repository.DeleteProduct(ctx, id)
}
