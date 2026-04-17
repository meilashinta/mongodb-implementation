package models

import "time"

// Product represents a product record in both MySQL and MongoDB.
type Product struct {
	ID          string    `json:"id" bson:"_id" db:"id"`
	Name        string    `json:"name" bson:"name" db:"name"`
	Description string    `json:"description" bson:"description" db:"description"`
	Category    string    `json:"category" bson:"category" db:"category"`
	Price       float64   `json:"price" bson:"price" db:"price"`
	Stock       int       `json:"stock" bson:"stock" db:"stock"`
	Tags        []string  `json:"tags" bson:"tags" db:"tags"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" db:"updated_at"`
}

// ProductFilter defines the search criteria used for MongoDB read queries.
type ProductFilter struct {
	Category string   `json:"category" bson:"category"`
	MinPrice float64  `json:"min_price" bson:"min_price"`
	MaxPrice float64  `json:"max_price" bson:"max_price"`
	Tags     []string `json:"tags" bson:"tags"`
}
