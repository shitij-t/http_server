package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// In-memory database
var (
	products   = make(map[int]Product)
	nextID     = 1
	productsMu sync.Mutex // mutex to protect access to 'products' map
)

func main() {
	// initialize dummy data
	productsMu.Lock()
	products[nextID] = Product{ID: nextID, Name: "Laptop", Price: 1200.00}
	nextID++
	products[nextID] = Product{ID: nextID, Name: "Mouse", Price: 25.00}
	nextID++
	productsMu.Unlock()

	// Register handlers for different API endpoints
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/products/", productByIDHandler)

	// start the http server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// productsHandler handles /products requests (GET for all, POST for new)
func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getProducts(w, r)
	case http.MethodPost:
		createProduct(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Product Id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getProductByID(w, r, id)
	case http.MethodDelete:
		deleteProduct(w, r, id)
	case http.MethodPut:
		updateProduct(w, r, id)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// -------- Handle rest operations ---------------

func getProducts(w http.ResponseWriter, r *http.Request) {
	productsMu.Lock()
	defer productsMu.Unlock()

	// convert map values to slice for json encoding
	var productList []Product
	for _, p := range products {
		productList = append(productList, p)
	}
	respondWithJSON(w, http.StatusOK, productList)
}

func getProductByID(w http.ResponseWriter, _ *http.Request, id int) {
	productsMu.Lock()
	defer productsMu.Unlock()

	product, ok := products[id]
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	respondWithJSON(w, http.StatusOK, product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var newProduct Product
	if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	productsMu.Lock()
	defer productsMu.Unlock()

	newProduct.ID = nextID
	products[newProduct.ID] = newProduct
	nextID++

	respondWithJSON(w, http.StatusCreated, newProduct)
}

func updateProduct(w http.ResponseWriter, r *http.Request, id int) {
	var updatedProduct Product
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	productsMu.Lock()
	defer productsMu.Unlock()

	_, ok := products[id]
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if updatedProduct.ID != 0 && updatedProduct.ID != id {
		http.Error(w, "ID in the URL and the Body do not match", http.StatusBadRequest)
		return
	}

	updatedProduct.ID = id

	products[id] = updatedProduct
	respondWithJSON(w, http.StatusOK, updatedProduct)
}

func deleteProduct(w http.ResponseWriter, r *http.Request, id int) {
	productsMu.Lock()
	defer productsMu.Unlock()

	_, ok := products[id]
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	delete(products, id)
	w.WriteHeader(http.StatusNoContent)
}
