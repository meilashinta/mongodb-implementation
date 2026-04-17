package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dualwrite/product-api/config"
	"github.com/dualwrite/product-api/handlers"
	"github.com/dualwrite/product-api/repositories"
	"github.com/dualwrite/product-api/services"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	database, err := config.InitDatabase(ctx)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.MongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("warning: failed to disconnect MongoDB client: %v", err)
		}
		_ = database.SQL.Close()
	}()

	repo := repositories.NewProductRepository(database.SQL, database.ProductCollection)
	service := services.NewProductService(repo)
	handler := handlers.NewProductHandler(service)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/product", handler.HandleProduct)
	http.HandleFunc("/products", handler.HandleProducts)
	http.HandleFunc("/product/", handler.HandleProductByID)

	port := getEnv("PORT", "8082")
	serverAddr := ":" + port
	log.Printf("Dual-Write Product API running on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message":"Product API is running","endpoints":["/products","/product"]}`)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
