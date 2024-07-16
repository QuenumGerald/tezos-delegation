package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// initTestDB initialise une base de données SQLite en mémoire pour les tests
func initTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS delegations (
        timestamp TEXT,
        amount INT64,
        delegator TEXT,
        level INT,
        PRIMARY KEY (timestamp, delegator))`)
	if err != nil {
		return nil, err
	}
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TestInitDB function
func TestInitDB(t *testing.T) {
	log.Println("Starting TestInitDB")

	db, err := initTestDB()
	assert.NoError(t, err, "Error initializing database")
	defer db.Close()

	assert.NotNil(t, db, "Database should be initialized")

	log.Println("Database initialized successfully")
}

// TestGetLastTimestamp function
func TestGetLastTimestamp(t *testing.T) {
	log.Println("Starting TestGetLastTimestamp")

	db, err := initTestDB()
	assert.NoError(t, err, "Error initializing database")
	defer db.Close()

	// Insert a test record
	_, err = db.Exec("INSERT INTO delegations (timestamp, amount, delegator, level) VALUES (?, ?, ?, ?)",
		"2023-01-01T00:00:00Z", 1000, "tz1TestAddress", 123456)
	assert.NoError(t, err, "Error inserting test record")

	log.Println("Test record inserted successfully")

	timestamp := getLastTimestamp(db)
	assert.Equal(t, "2023-01-01T00:00:00Z", timestamp, "Timestamp should match the latest record")

	log.Printf("Timestamp retrieved: %s", timestamp)
}

// TestGetDelegations function
func TestGetDelegations(t *testing.T) {
	log.Println("Starting TestGetDelegations")

	db, err := initTestDB()
	assert.NoError(t, err, "Error initializing database")
	defer db.Close()

	// Insert test records
	_, err = db.Exec("INSERT INTO delegations (timestamp, amount, delegator, level) VALUES (?, ?, ?, ?)",
		"2023-01-01T00:00:00Z", 1000, "tz1TestAddress", 123456)
	assert.NoError(t, err, "Error inserting test record")

	log.Println("Test record inserted successfully")

	req, err := http.NewRequest("GET", "/xtz/delegations", nil)
	assert.NoError(t, err, "Error creating HTTP request")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getDelegations(w, r, db)
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be OK")

	expected := `{"data":[{"timestamp":"2023-01-01T00:00:00Z","amount":1000,"delegator":"tz1TestAddress","level":123456}]}`
	assert.JSONEq(t, expected, rr.Body.String(), "Response body should match")

	log.Printf("Response body: %s", rr.Body.String())
}

// TestFetchDelegations function with real API
func TestFetchDelegationsRealAPI(t *testing.T) {
	log.Println("Starting TestFetchDelegationsRealAPI")

	db, err := initTestDB()
	assert.NoError(t, err, "Error initializing database")
	defer db.Close()

	realAPIURL := "https://api.tzkt.io/v1/operations/delegations?limit=1"

	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fetchDelegationsFromURL(db, realAPIURL, stopChan)
	}()

	// Let the goroutine run for a bit
	time.Sleep(30 * time.Second)
	close(stopChan)
	wg.Wait()

	// Check if data is inserted (assuming at least one delegation is returned by the real API)
	var count int
	row := db.QueryRow("SELECT COUNT(*), timestamp, amount, delegator, level FROM delegations")
	var d Delegation
	err = row.Scan(&count, &d.Timestamp, &d.Amount, &d.Delegator, &d.Level)
	assert.NoError(t, err, "Error scanning database")
	assert.Greater(t, count, 0, "At least one delegation should be inserted")

	log.Printf("TestFetchDelegationsRealAPI completed successfully, count: %d, first delegation: %v", count, d)
}

// fetchDelegationsFromURL récupère les délégations depuis une URL et les stocke dans la base de données
func fetchDelegationsFromURL(db *sql.DB, url string, stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			fmt.Println("FetchDelegations stopping...")
			return
		default:
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
				Level     int    `json:"level"`
			}
			if err := json.Unmarshal(body, &fetched); err != nil {
				log.Println("Error unmarshaling:", err)
				time.Sleep(30 * time.Second)
				continue
			}

			for _, f := range fetched {
				fmt.Printf("Fetched delegation - Timestamp: %s, Amount: %d, Delegator: %s, Level: %d\n", f.Timestamp, f.Amount, f.Sender.Address, f.Level)
				_, err := db.Exec("INSERT OR IGNORE INTO delegations (timestamp, amount, delegator, level) VALUES (?, ?, ?, ?)",
					f.Timestamp, f.Amount, f.Sender.Address, f.Level)
				if err != nil {
					log.Println("Database insert error:", err)
				}
			}

			time.Sleep(30 * time.Second)
		}
	}
}
