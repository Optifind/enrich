package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lattots/enrich/pkg/product"
)

// HandleGetProduct handles GET request to single product endpoint. See API documentation for more details.
func (h *Handler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	// Product ID is retrieved from URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "No product ID provided", http.StatusBadRequest)
		return
	}

	// Product is fetched from database.
	p, err := product.GetById(
		h.MilvusClient,
		h.Config,
		id,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting product", http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, "no product found", http.StatusNotFound)
		return
	}

	// Product's info is written to response writer.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintln(w, *p)
	if err != nil {
		log.Println(err)
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}
