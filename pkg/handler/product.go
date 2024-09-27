package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lattots/enrich/pkg/catalog"
	"github.com/lattots/enrich/pkg/product"
)

// ProductResponse is a struct that represents a product without certain fields for the API response.
type ProductResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	Link        string  `json:"product-url"`
	Image       string  `json:"product-image"`
}

// HandleGetProduct handles GET request to single product endpoint. See API documentation for more details.
func (h *Handler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	// Product ID is retrieved from URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "no id was provided", http.StatusBadRequest)
		return
	}

	// Product is fetched from database.
	p, err := product.GetById(
		h.DB,
		h.Config,
		id,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting product", http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, "no product found for the given id", http.StatusNotFound)
		return
	}

	productResponse := convertProductToResponse(*p)

	// Product's info is written to response writer.
	writeJSONResponse(w, http.StatusOK, productResponse)
}

// HandleGetProducts handles GET request to multiple product endpoint. Method will route request to
// HandleGetRandomProducts if no `similar_to` IDs are provided in the request.
// See API documentation for more details.
func (h *Handler) HandleGetProducts(w http.ResponseWriter, r *http.Request) {
	// Similar to IDs are retrieved from request.
	idsStr := r.URL.Query().Get("similar_to")
	if idsStr == "" {
		// If no similar to IDs are provided, server returns random products.
		h.HandleGetRandomProducts(w, r)
		return
	}

	ids := strings.Split(idsStr, ",")

	// Product count is retrieved from request.
	countStr := r.URL.Query().Get("count")
	count, err := parseCount(countStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Catalog object is created for holding product data.
	cat := catalog.Catalog{
		DB:       h.DB,
		DBConfig: h.Config.DB,
	}

	// Specified IDs are loaded to catalog.
	err = cat.Load(count, ids...)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting products", http.StatusInternalServerError)
		return
	}

	if len(cat.Products) == 0 {
		http.Error(w, "no products found for the given ids", http.StatusNotFound)
		return
	}

	// Vector database is searched for products that best match the Catalogs products.
	searchResults, err := cat.SearchSimilarStyle(count)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting products", http.StatusInternalServerError)
		return
	}

	// Result products are converted to a slice of ProductResponse.
	productResponses := convertProductsToResponses(searchResults)

	// Product info is written to response writer.
	writeJSONResponse(w, http.StatusOK, productResponses)
}

func (h *Handler) HandleGetRandomProducts(w http.ResponseWriter, r *http.Request) {
	// Product count is retrieved from request.
	countStr := r.URL.Query().Get("count")
	count, err := parseCount(countStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Catalog object is created for holding product data.
	cat := catalog.Catalog{
		DB:       h.DB,
		DBConfig: h.Config.DB,
	}

	// Product info is loaded from Milvus
	err = cat.Load(count)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading product info", http.StatusInternalServerError)
		return
	}

	if len(cat.Products) == 0 {
		http.Error(w, "no products to return", http.StatusNotFound)
		return
	}

	// Result products are converted to a slice of ProductResponse.
	productResponses := convertProductsToResponses(cat.Products)

	// Product info is written to response writer.
	writeJSONResponse(w, http.StatusOK, productResponses)
}

// writeJSONResponse writes JSON encodable data to response writer with the provided status code.
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("error writing JSON response: %v\n", err)
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}

const maxCount = 1000

// parseCount tries to extract count from string. If no count is found, function defaults to predetermined number.
// Returns an integer and an error.
func parseCount(countStr string) (int, error) {
	const defaultCount = 50
	if countStr == "" {
		return defaultCount, nil
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || !isValidCount(count) {
		return 0, fmt.Errorf("count is not an integer between [1-%d] or null", maxCount)
	}
	return count, nil
}

// isValidCount checks if count is within certain parameters.
func isValidCount(count int) bool {
	return count > 0 && count <= maxCount
}

// convertProductsToResponses converts a list of Product to a list of ProductResponse.
func convertProductsToResponses(products []*product.Product) []ProductResponse {
	var responses []ProductResponse
	for _, p := range products {
		resp := convertProductToResponse(*p)
		responses = append(responses, resp)
	}
	return responses
}

// convertProductToResponse converts a Product to a ProductResponse.
func convertProductToResponse(product product.Product) ProductResponse {
	resp := ProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Link:        product.Link,
		Image:       product.Image,
	}

	return resp
}
