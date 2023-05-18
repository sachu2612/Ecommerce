package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	Placed     OrderStatus = "Placed"
	Dispatched OrderStatus = "Dispatched"
	Completed  OrderStatus = "Completed"
	Canceled   OrderStatus = "Canceled"
)

// Order represents an order
type Order struct {
	ID           string      `json:"id"`
	Products     []Product   `json:"products"`
	OrderValue   float64     `json:"orderValue"`
	Status       OrderStatus `json:"status"`
	DispatchDate string      `json:"dispatchDate"`
	ProdQuantity int         `json:"prodQuantity"`
}

// ProductCategory represents the category of a product
type ProductCategory string

const (
	Premium ProductCategory = "Premium"
	Regular ProductCategory = "Regular"
	Budget  ProductCategory = "Budget"
)

// Product represents a product
type Product struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Price    float64         `json:"price"`
	Category ProductCategory `json:"category"`
	Quantity int             `json:"quantity"`
}

// ProductService represents a service to manage products
type ProductService struct {
	DB *sql.DB
}

// NewProductService creates a new ProductService
func NewProductService(db *sql.DB) *ProductService {
	return &ProductService{
		DB: db,
	}
}

// getProductCatalog retrieves the product catalog
func (ps *ProductService) getProductCatalog(w http.ResponseWriter, r *http.Request) {
	rows, err := ps.DB.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var productCatalog []Product

	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Category,
			&product.Quantity,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		productCatalog = append(productCatalog, product)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(productCatalog)
}

// updateProductCatalog updates the product catalog
func (ps *ProductService) updateProductCatalog(w http.ResponseWriter, r *http.Request) {
	var updatedProducts []Product
	err := json.NewDecoder(r.Body).Decode(&updatedProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := ps.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updateStmt, err := tx.Prepare("UPDATE products SET name = ?, price = ?, category = ?, quantity = ? WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer updateStmt.Close()

	for _, updatedProduct := range updatedProducts {
		_, err := updateStmt.Exec(
			updatedProduct.Name,
			updatedProduct.Price,
			updatedProduct.Category,
			updatedProduct.Quantity,
			updatedProduct.ID,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// OrderService represents a service to manage orders
type OrderService struct {
	DB *sql.DB
}

// NewOrderService creates a new OrderService
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{
		DB: db,
	}
}

// placeOrder places a new order
func (os *OrderService) placeOrder(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrderProducts []*Product `json:"orderProducts"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderProducts := payload.OrderProducts
	tx, err := os.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, product := range orderProducts {
		if product.Quantity < 1 || product.Quantity > 10 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("invalid quantity for product: %s", product.Name), http.StatusBadRequest)
			return
		}

		_, err := tx.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?", product.Quantity, product.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	orderID := generateOrderID()
	orderValue := calculateOrderValue(orderProducts)

	_, err = tx.Exec("INSERT INTO orders (id, value, status, prodQuantity) VALUES (?, ?, ?, ?)", orderID, orderValue, OrderStatus(Placed), len(orderProducts))
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, product := range orderProducts {
		_, err := tx.Exec("INSERT INTO order_products (order_id, product_id) VALUES (?, ?)", orderID, product.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string `json:"message"`
		OrderID string `json:"orderID"`
	}{
		Message: "Order placed successfully",
		OrderID: orderID,
	}

	json.NewEncoder(w).Encode(response)
}

// updateOrderStatus updates the order status
func (os *OrderService) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["orderID"]
	var status struct {
		Status OrderStatus `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var dispatchDate string
	if status.Status == Dispatched {
		dispatchDate = time.Now().Format("2006-01-02")
	}

	result, err := os.DB.Exec("UPDATE orders SET status = ?, dispatchDate = ? WHERE id = ?", status.Status, dispatchDate, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		http.NotFound(w, r)
	}
}

// generateOrderID generates a unique order ID
func generateOrderID() string {
	// Generate a unique order ID (e.g., using UUID or other mechanisms)
	// This is just a simple implementation for demonstration purposes
	return fmt.Sprintf("ORD%d", time.Now().Unix())
}

// calculateOrderValue calculates the total value of an order
func calculateOrderValue(products []*Product) float64 {
	premiumProductCount := 0
	totalOrderValue := 0.0

	for _, product := range products {
		totalOrderValue += product.Price * float64(product.Quantity)

		if product.Category == Premium {
			premiumProductCount++
		}
	}

	if premiumProductCount >= 3 {
		totalOrderValue *= 0.9 // 10% discount
	}

	return totalOrderValue
}

func main() {
	// Database connection configuration
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/product_order_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productService := NewProductService(db)
	orderService := NewOrderService(db)

	router := mux.NewRouter()

	// Product Catalog Endpoint
	router.HandleFunc("/products", productService.getProductCatalog).Methods("GET")

	// Update Product Catalog Endpoint
	router.HandleFunc("/products", productService.updateProductCatalog).Methods("PUT")

	// Place Order Endpoint
	router.HandleFunc("/orders", orderService.placeOrder).Methods("POST")

	// Update Order Status Endpoint
	router.HandleFunc("/orders/{orderID}", orderService.updateOrderStatus).Methods("PATCH")

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
