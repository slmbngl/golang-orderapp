# golang-orderapp
This project is a REST API with user authentication, product management, and order processing features. Written in Go and uses PostgreSQL database.

## Features

### 🔐 Authentication
- User registration and login
- JWT token-based authentication
- Secure password hashing

### 📦 Product Management
- List all products
- View product details
- Add new products (Admin)
- Update products (Admin)
- Delete products (Admin)

### 🛒 Order Management
- Create orders
- List user orders
- View order details
- Update order status (pending, confirmed, cancelled)
- Delete orders
- Stock control and automatic updates

## Technologies

- **Backend**: Go (Golang)
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL
- **ORM**: pgx/v5
- **Authentication**: JWT
- **Documentation**: Swagger/OpenAPI
- **Environment Variables**: godotenv

## Requirements

- Go 1.21+
- PostgreSQL 12+
- Git

## Installation

### 1. Clone the Repository
```bash
git clone https://github.com/slmbngl/OrderAplication.git
cd OrderAplication
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Set Environment Variables
Create a `.env` file:
```env
DATABASE_URL=postgres://username:password@localhost:5432/order_app
```

### 4. Create Database
Create the `order_app` database in PostgreSQL and add the following tables:

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order Items table
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order Summary View
CREATE VIEW order_summary_view AS
SELECT 
    o.id as order_id,
    o.user_id,
    o.total_amount,
    o.status,
    o.created_at,
    u.username
FROM orders o
JOIN users u ON o.user_id = u.id;
```

### 5. Run the Application
```bash
go run cmd/main.go
```

The application will run at `http://localhost:4504`.

## API Documentation

You can access the Swagger UI documentation at:
```
http://localhost:4504/swagger/
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

- **Developer**: [slmbngl](https://github.com/slmbngl)
- **Email**: slmbingol@gmail.com

