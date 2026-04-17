package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/dualwrite/product-api/models"
	"github.com/dualwrite/product-api/services"
)

// ProductHandler exposes HTTP endpoints for product CRUD functionality.
type ProductHandler struct {
	service *services.ProductService
}

// NewProductHandler creates a new handler instance.
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// HandleProduct is responsible for POST /product.
func (h *ProductHandler) HandleProduct(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createProduct(w, r)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleProducts is responsible for GET /products.
func (h *ProductHandler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := models.ProductFilter{
		Category: r.URL.Query().Get("category"),
	}

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		filter.MinPrice = parseFloat(minPrice)
	}
	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		filter.MaxPrice = parseFloat(maxPrice)
	}
	if tagQuery := r.URL.Query().Get("tags"); tagQuery != "" {
		filter.Tags = strings.Split(tagQuery, ",")
	}

	products, err := h.service.SearchProducts(r.Context(), filter)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, products)
}

// HandleProductByID is responsible for GET, PUT, DELETE /product/{id}.
func (h *ProductHandler) HandleProductByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/product/")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "missing product id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getProductByID(w, r, id)
	case http.MethodPut:
		h.updateProduct(w, r, id)
	case http.MethodDelete:
		h.deleteProduct(w, r, id)
	default:
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := h.decodeJSON(r.Body, &product); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if err := h.service.CreateProduct(r.Context(), &product); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) getProductByID(w http.ResponseWriter, r *http.Request, id string) {
	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		h.respondError(w, http.StatusNotFound, "product not found")
		return
	}

	h.respondJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) updateProduct(w http.ResponseWriter, r *http.Request, id string) {
	var product models.Product
	if err := h.decodeJSON(r.Body, &product); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if err := h.service.UpdateProduct(r.Context(), id, &product); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *ProductHandler) deleteProduct(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) decodeJSON(body io.ReadCloser, dest interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(dest)
}

func (h *ProductHandler) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Slice && v.IsNil() {
		payload = v.Slice(0, 0).Interface()
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *ProductHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}
