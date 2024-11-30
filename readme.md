# **Numeris - Invoice Management System**

Numeris is a backend system designed for managing invoices. It provides APIs for user authentication, invoice creation, retrieval, and management. The system is built for performance, scalability, and ease of maintenance.

## Table of Contents

1. [Introduction](#introduction)
2. [System Overview](#system-overview)
3. [Architecture](#architecture)
4. [Features](#features)
5. [Endpoints](#endpoints)
6. [Technologies and Tools](#technologies-and-tools)
7. [System Design](#system-design)
8. [Scalability and Maintainability](#scalability-and-maintainability)
9. [Getting Started](#getting-started)

## **Introduction**

Numeris is a backend system that manages invoices. It offers functionalities for user registration, invoice management, and related operations. The system is optimized for handling multiple users and invoices efficiently.

## **System Overview**

Numeris is organized into distinct layers to separate responsibilities:

1. **Application Layer (`app/`)**: Manages HTTP requests, routing, and business logic.
2. **Domain Layer (`domain/`)**: Defines core business entities and logic (e.g., `User`, `Invoice`).
3. **Infrastructure Layer (`db/`)**: Handles data storage and interactions with external services.
4. **Entry Point (`cmd/`)**: Initializes the application and routes.

## **Architecture**

The system uses a layered approach:

1. **Application Layer**: Handles user requests, performs business logic, and interacts with other layers.
2. **Domain Layer**: Contains the main business rules and models for users and invoices.
3. **Infrastructure Layer**: Manages data storage and integrates with external services.
4. **Entry Point**: The application's starting point and route configuration.

## **Features**

- **User Authentication**: Users can sign up and log in to the system.
- **Invoice Management**: Create, retrieve, update, and delete invoices.
- **Document Download**: Retrieve invoice documents (e.g., PDF).
- **Invoice Statistics**: View summary statistics for a user's invoices.
- **Activity Logging**: Track user actions related to invoice management.

## **Endpoints**

1. `GET /`: Welcome message for the Numeris API.
2. `POST /api/register`: User registration.
3. `POST /api/login`: User authentication (returns JWT).
4. `POST /api/invoice/:userID/create`: Create a new invoice.
5. `GET /api/invoice/:userID/get/:invoiceID`: Retrieve an invoice by ID.
6. `GET /api/invoice/:userID/all/:invoiceID`: List all invoices for a user.
7. `PUT /api/invoice/:userID/update/:invoiceID`: Update an unissued invoice.
8. `DELETE /api/invoice/:userID/delete/:invoiceID`: Delete an invoice.
9. `GET /api/invoice/:userID/stats/:invoiceID`: Get invoice statistics for a user.
10. `POST /api/invoice/:userID/send/:invoiceID`: Send an issued invoice to the customer.
11. `GET /api/invoice/:userID/download/:invoiceID`: Download an invoice PDF.
12. `GET /api/invoice/:userID/activities/:invoiceID`: Get activities related to a specific invoice.

## **Technologies and Tools**

- **Language**: Go
- **Web Framework**: Fiber
- **Database**: MongoDB
- **Authentication**: JWT
- **Validation**: go-playground/validator
- **Logging**: `slog`

## **System Design**

The system is built with a focus on clear separation of responsibilities:

1. **Dependency Injection**: Used for managing dependencies between components.
2. **Repository Pattern**: Data access is abstracted for flexibility.
3. **Service Layer**: Contains the core business logic, separate from HTTP handlers.
4. **Middleware**: Handles tasks like authentication and error handling.

### **Core Components:**

- **Application Struct**: Manages application configuration and routing.
- **Handlers**: Implement the logic for each API endpoint.
- **Repositories**: Responsible for data storage and retrieval.
- **Services**: Encapsulate business logic and interactions with external services.
- **Models**: Represent business entities and request data.

### **Middleware**:
- **Authentication Middleware**: Protects routes by verifying JWTs.
- **Error Handling Middleware**: Catches and processes errors.

### **Utilities**:
- **Validation**: Ensures data integrity using `go-playground/validator`.
- **Error Definitions**: Standardized error responses.
- **Helper Functions**: Perform tasks like password hashing and JWT generation.

## **Scalability and Maintainability**

The system is designed for efficient scaling and ease of maintenance:

1. **Database Scalability**: MongoDB supports horizontal scaling via sharding.
2. **Stateless Design**: The system is stateless, making it easier to scale by adding more instances.
3. **Modular Code**: The codebase is organized for clarity and easy extension. Adding new features or modifying existing ones is straightforward.
4. **Logging**: Structured logging helps with monitoring and troubleshooting.

## **Getting Started**

To get started with the Numeris backend:

1. **Clone the Repository**:
    ```bash
    git clone https://github.com/TheBraveByte/numeris.git
    cd numeris
    ```

2. **Set Up MongoDB**:
    - Ensure MongoDB is installed and running.
    - Set the connection string in the environment variables.

3. **Install Dependencies**:
    ```bash
    go mod download
    ```

4. **Build the Application**:
    ```bash
    go build ./cmd/main.go
    ```

5. **Run the Application**:
    ```bash
    ./main
    ```
---