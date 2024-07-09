
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

- **Go**: The programming language used.
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
    ```go
    func initDB() *sql.DB {
        database, err := sql.Open("sqlite3", "./delegations.db")
        if err != nil {
            log.Fatal(err)
        }
        statement, err := database.Prepare(\`
            CREATE TABLE IF NOT EXISTS delegations (
                timestamp TEXT,
                amount INT64,
                delegator TEXT,
                block TEXT,
                PRIMARY KEY (timestamp, delegator))
        \`)
        if err != nil {
            log.Fatal(err)
        }
        statement.Exec()
        return database
    }
    ```

- `fetchDelegations(db *sql.DB)`: Periodically fetches data from the Tezos API and inserts new delegations into the database.
    ```go
    func fetchDelegations(db *sql.DB, wg *sync.WaitGroup, stopChan chan struct{}) {
        defer wg.Done()
        for {
            select {
            case <-stopChan:
                fmt.Println("FetchDelegations stopping...")
                return
            default:
                lastTimestamp := getLastTimestamp(db)
                url := fmt.Sprintf("https://api.tzkt.io/v1/operations/delegations?timestamp.gt=%s&limit=10000", lastTimestamp)
                resp, err := http.Get(url)
                if err != nil {
                    log.Println("Error fetching data:", err)
                    time.Sleep(30 * time.Second)
                    continue
                }

                body, err := io.ReadAll(resp.Body)
                if err != nil {
                    log.Println("Error reading body:", err)
                    resp.Body.Close()
                    time.Sleep(30 * time.Second)
                    continue
                }
                resp.Body.Close()

                var fetched []struct {
                    Timestamp string `json:"timestamp"`
                    Amount    int64  `json:"amount"`
                    Sender    Sender `json:"sender"`
                    Block     string `json:"block"`
                }
                if err := json.Unmarshal(body, &fetched); err != nil {
                    log.Println("Error unmarshaling:", err)
                    time.Sleep(30 * time.Second)
                    continue
                }

                for _, f := range fetched {
                    fmt.Printf("Fetched delegation - Timestamp: %s, Amount: %d, Delegator: %s, Block: %s
", f.Timestamp, f.Amount, f.Sender.Address, f.Block)
                    _, err := db.Exec("INSERT OR IGNORE INTO delegations (timestamp, amount, delegator, block) VALUES (?, ?, ?, ?)",
                        f.Timestamp, f.Amount, f.Sender.Address, f.Block)
                    if err != nil {
                        log.Println("Database insert error:", err)
                    }
                }
                time.Sleep(30 * time.Second)
            }
        }
    }
    ```

- `getLastTimestamp(db *sql.DB)`: Retrieves the most recent timestamp stored in the database.
    ```go
    func getLastTimestamp(db *sql.DB) string {
        var lastTimestamp string
        row := db.QueryRow("SELECT timestamp FROM delegations ORDER BY timestamp DESC LIMIT 1")
        switch err := row.Scan(&lastTimestamp); err {
        case sql.ErrNoRows:
            return "2022-01-23T00:00:00Z"
        case nil:
            return lastTimestamp
        default:
            log.Println("Error getting last timestamp:", err)
            return "2022-01-23T00:00:00Z"
        }
    }
    ```

- `getDelegations(w http.ResponseWriter, r *http.Request, db *sql.DB)`: Retrieves delegations from the SQLite database and sends them as a JSON-encoded HTTP response.
    ```go
    func getDelegations(w http.ResponseWriter, r *http.Request, db *sql.DB) {
        year := r.URL.Query().Get("year")
        var rows *sql.Rows
        var err error

        if year != "" {
            rows, err = db.Query("SELECT * FROM delegations WHERE strftime('%Y', timestamp) = ? ORDER BY timestamp DESC", year)
        } else {
            rows, err = db.Query("SELECT * FROM delegations ORDER BY timestamp DESC")
        }

        if err != nil {
            log.Fatal(err)
        }
        defer rows.Close()

        var delegations []Delegation
        for rows.Next() {
            var d Delegation
            if err := rows.Scan(&d.Timestamp, &d.Amount, &d.Delegator, &d.Block); err != nil {
                log.Fatal(err)
            }
            delegations = append(delegations, d)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string][]Delegation{"data": delegations})
    }
    ```

## Error Handling

- API fetch errors and database errors are logged but do not stop the application.
- JSON marshaling and unmarshaling errors are also logged for easier debugging.

