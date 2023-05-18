CREATE DATABASE product_order_db;
USE product_order_db;
CREATE TABLE products (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(20) NOT NULL,
    quantity INT NOT NULL
);
CREATE TABLE orders (
    id VARCHAR(20) PRIMARY KEY,
    value DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    dispatchDate DATE,
    prodQuantity INT NOT NULL
);
CREATE TABLE order_products (
    order_id VARCHAR(20),
    product_id VARCHAR(10),
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
-- Insert values into the products table
INSERT INTO products (id, name, price, category, quantity) VALUES
    ('P1', 'Product 1', 19.99, 'Regular', 100),
    ('P2', 'Product 2', 29.99, 'Regular', 50),
    ('P3', 'Product 3', 49.99, 'Premium', 20),
    ('P4', 'Product 4', 9.99, 'Budget', 200),
    ('P5', 'Product 5', 19.99, 'Premium', 100),
    ('P6', 'Product 6', 29.99, 'Premium', 50),
    ('P7', 'Product 7', 49.99, 'Premium', 20);

-- Insert values into the orders table
INSERT INTO orders (id, value, status, prodQuantity) VALUES
    ('ORD1', 39.98, 'Placed', 2),
    ('ORD2', 59.97, 'Dispatched', 3),
    ('ORD3', 99.99, 'Completed', 1);