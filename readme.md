# Product Order System

This is a simple product order system implemented in Go.

## Usage

1. Set up a MySQL database and update the database connection details in `internal/utils/db.go`.
2. Install dependencies: `go mod tidy`.
3. Build and run the application: `go run cmd/main.go`.

The server will start on `http://localhost:8080`.

## API Endpoints

### Product Catalog Endpoint

- `GET /products`: Get the product catalog.
  curl --location --request GET 'http://localhost:8080/products' \
  --header 'Content-Type: application/json' \
  --data '

### Place Order Endpoint

- `POST /orders`: Place an order.
  curl --location 'http://localhost:8080/orders' \
  --header 'Content-Type: application/json' \
  --data '{
  "orderProducts": [
  {
  "id": "P6",
  "name": "Product 6",
  "price": 29.99,
  "category": "Premium",
  "quantity": 5
  },
  {
  "id": "P7",
  "name": "Product 7",
  "price": 49.99,
  "category": "Premium",
  "quantity": 2
  },
  {
  "id": "P3",
  "name": "Product 3",
  "price": 49.99,
  "category": "Premium",
  "quantity": 11
  }
  ]
  }'

### Update Order Status Endpoint

- `PATCH /orders/{orderID}`: Update the status of an order.

curl --location --request PATCH 'http://localhost:8080/orders/ORD1684398944' \
--header 'Content-Type: application/json' \
--data '{
"status": "Dispatched"
}'
