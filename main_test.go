package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// Test initDB function
func TestInitDB(t *testing.T) {
	log.Println("Starting TestInitDB")

	db := initDB()
	defer db.Close()

	assert.NotNil(t, db, "Database should be initialized")

	log.Println("Database initialized successfully")
}

// Test getLastTimestamp function
func TestGetLastTimestamp(t *testing.T) {
	log.Println("Starting TestGetLastTimestamp")

	db := initDB()
	defer db.Close()

	// Clean the database before testing
	_, err := db.Exec("DELETE FROM delegations")
	assert.NoError(t, err, "Error cleaning the database")

	log.Println("Database cleaned successfully")

	// Insert a test record
	_, err = db.Exec("INSERT INTO delegations (timestamp, amount, delegator, level) VALUES (?, ?, ?, ?)",
		"2023-01-01T00:00:00Z", 1000, "tz1TestAddress", 123456)
	assert.NoError(t, err, "Error inserting test record")

	log.Println("Test record inserted successfully")

	timestamp := getLastTimestamp(db)
	assert.Equal(t, "2023-01-01T00:00:00Z", timestamp, "Timestamp should match the latest record")

	log.Printf("Timestamp retrieved: %s", timestamp)
}

// Test getDelegations function
func TestGetDelegations(t *testing.T) {
	log.Println("Starting TestGetDelegations")

	db := initDB()
	defer db.Close()

	// Clean the database before testing
	_, err := db.Exec("DELETE FROM delegations")
	assert.NoError(t, err, "Error cleaning the database")

	log.Println("Database cleaned successfully")

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

// Test the entire application flow
func TestIntegration(t *testing.T) {
	log.Println("Starting TestIntegration")

	db := initDB()
	defer db.Close()

	// Clean the database before testing
	_, err := db.Exec("DELETE FROM delegations")
	assert.NoError(t, err, "Error cleaning the database")

	log.Println("Database cleaned successfully")

	// Start the HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/xtz/delegations", func(w http.ResponseWriter, r *http.Request) {
		getDelegations(w, r, db)
	}).Methods("GET")
	server := httptest.NewServer(r)
	defer server.Close()

	log.Println("HTTP server started")

	// Insert a test record
	_, err = db.Exec("INSERT INTO delegations (timestamp, amount, delegator, level) VALUES (?, ?, ?, ?)",
		"2023-01-01T00:00:00Z", 1000, "tz1TestAddress", 123456)
	assert.NoError(t, err, "Error inserting test record")

	log.Println("Test record inserted successfully")

	// Perform a GET request
	resp, err := http.Get(server.URL + "/xtz/delegations")
	assert.NoError(t, err, "Error performing GET request")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response should be OK")

	expected := `{"data":[{"timestamp":"2023-01-01T00:00:00Z","amount":1000,"delegator":"tz1TestAddress","level":123456}]}`
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Error reading response body")
	assert.JSONEq(t, expected, string(bodyBytes), "Response body should match")

	log.Printf("Response body: %s", string(bodyBytes))
}
