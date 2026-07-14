# POS System

A modern Point of Sale system built with **Go + Fiber** (backend) and **SvelteKit + Tailwind CSS** (frontend).

## Stack

- **Backend:** Go, Fiber, PostgreSQL, JWT
- **Frontend:** SvelteKit 5 (runes), Tailwind CSS v4, Lucide icons

## Setup

### 1. PostgreSQL

Create a database:
```sql
CREATE DATABASE pos_db;
```

### 2. Backend

```sh
cd backend
cp .env.example .env
# Edit .env with your database credentials and JWT secret

go run ./cmd/seed   # creates default admin: admin@pos.local / admin123
go run ./cmd/main.go
```

The API runs on `http://localhost:8080`.

### 3. Frontend

```sh
# from the project root
npm install
npm run dev
```

The app runs on `http://localhost:5173`.

## Default Login

| Email             | Password  | Role  |
|-------------------|-----------|-------|
| admin@pos.local   | admin123  | Admin |

Change the password after first login via **Users → Edit**.

## Features

- **Sales** — barcode scanner, hold/resume, discounts, tax, multiple payment methods (F2 focus search, F4 checkout, F5 hold)
- **Products** — categories, barcode, SKU, stock tracking, reorder alerts
- **Inventory** — stock adjustments with audit trail
- **Customers & Suppliers** — full CRUD with search
- **Purchases** — receive stock from suppliers
- **Reports** — daily/monthly sales, top products, inventory value
- **Users** — role-based access (admin / manager / cashier)
- **Settings** — category management

## Project Structure

```
pos/
├── backend/
│   ├── cmd/
│   │   ├── main.go        # entry point
│   │   └── seed/main.go   # create first admin user
│   └── internal/
│       ├── database/      # connection + schema migrations
│       ├── models/        # domain types + input structs
│       ├── repositories/  # SQL queries
│       ├── services/      # business logic
│       ├── handlers/      # HTTP handlers
│       ├── middleware/     # JWT auth
│       ├── routes/        # route registration
│       └── utils/         # response helpers
└── src/
    ├── lib/
    │   ├── components/    # Sidebar, Modal, Pagination, Notification
    │   ├── services/      # API client per domain
    │   ├── stores/        # auth, cart, notification (Svelte 5 runes)
    │   └── types/         # TypeScript interfaces
    └── routes/
        ├── login/
        └── (app)/         # protected layout with sidebar
            ├── dashboard/
            ├── sales/
            ├── products/
            ├── inventory/
            ├── customers/
            ├── suppliers/
            ├── purchases/
            ├── reports/
            ├── users/
            └── settings/
```
