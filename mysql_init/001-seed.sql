CREATE DATABASE IF NOT EXISTS sample_app;
USE sample_app;

CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    order_total DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
);

INSERT INTO users (first_name, last_name, email) VALUES
('Anand', 'Siva', 'anand@example.com'),
('Jane', 'Doe', 'jane@example.com'),
('John', 'Smith', 'john@example.com');

INSERT INTO orders (user_id, order_total, status) VALUES
(1, 120.50, 'paid'),
(1, 42.00, 'pending'),
(2, 89.99, 'paid'),
(3, 15.75, 'cancelled');
