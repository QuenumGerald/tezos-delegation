
# Tezos Delegation Management Application in Go

## Table of Contents

1. [Introduction](#introduction)
2. [Dependencies](#dependencies)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Usage](#usage)
6. [API](#api)
7. [Code Structure](#code-structure)
8. [Error Handling](#error-handling)
9. [Future Improvements](#future-improvements)

## Introduction

This application fetches and stores delegation data from the Tezos blockchain. It is built using the Go programming language, SQLite3 for database storage, and Gorilla Mux for HTTP routing.

## Dependencies

- **SQLite3**: The database for storing delegation data.
- **Gorilla Mux**: A HTTP router for Go.

To install Gorilla Mux, use the following command:
```sh
go get -u github.com/gorilla/mux
```

## Installation

1. Ensure you have Go and SQLite3 installed on your system.
2. Clone this repository to your local machine.
3. Navigate to the directory where the `main.go` file is located.
4. Install the necessary dependencies using `go mod tidy`.

## Configuration

The SQLite database (`delegations.db`) is automatically created in the same directory as the `main.go` file. To change the database path, modify the `initDB()` function in the source code.

## Usage

### Running the Application Locally

1. Open your terminal and navigate to the directory containing `main.go`.
2. Run the application using the following command:
    ```sh
    go run main.go
    ```
3. The server will start and listen on port 8000. You should see the output:
    ```
    Server is running on port 8000
    ```

### Accessing the Delegations Data

1. Open your web browser or use a tool like `curl` or Postman.
2. Access the endpoint to retrieve delegations:
    ```sh
    curl http://localhost:8000/xtz/delegations
    ```

3. To filter delegations by year, use the `year` query parameter:
    ```sh
    curl http://localhost:8000/xtz/delegations?year=2023
    ```

## API

- **Endpoint**: `GET /xtz/delegations`
    - **Description**: Fetches the delegations from the database and returns them as a JSON array.
    - **Optional Query Parameter**: `year` - Filters data for the specified year in YYYY format.

## Code Structure

### Types

- `Sender`: Struct to hold the sender's address.
    ```go
    type Sender struct {
        Address string `json:"address"`
    }
    ```

- `Delegation`: Struct to hold the delegation data.
    ```go
    type Delegation struct {
        Timestamp string `json:"timestamp"`
        Amount    int64  `json:"amount"`
        Delegator string `json:"delegator"`
        Block     string `json:"block"`
    }
    ```

### Functions

- `initDB()`: Initializes the SQLite database and returns the `*sql.DB` instance. It creates the `delegations` table if it doesn't exist.
    

- `fetchDelegations(db *sql.DB)`: Periodically fetches data from the Tezos API and inserts new delegations into the database.
    
- `getLastTimestamp(db *sql.DB)`: Retrieves the most recent timestamp stored in the database.
    

- `getDelegations(w http.ResponseWriter, r *http.Request, db *sql.DB)`: Retrieves delegations from the SQLite database and sends them as a JSON-encoded HTTP response.
    

## Error Handling

- API fetch errors and database errors are logged but do not stop the application.
- JSON marshaling and unmarshaling errors are also logged for easier debugging.

