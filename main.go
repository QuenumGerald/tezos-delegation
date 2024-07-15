package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Sender struct pour contenir les informations sur l'expéditeur
type Sender struct {
	Address string `json:"address"`
}

// Delegation struct pour contenir les informations sur la délégation
type Delegation struct {
	Timestamp string `json:"timestamp"`
	Amount    int64  `json:"amount"`
	Delegator string `json:"delegator"`
	Level     int    `json:"level"`
}

// initDB initialise la base de données et ajoute des index pour améliorer les performances
func initDB() (*sql.DB, error) {
	dbPath := "./delegations.db"
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	statement, err := database.Prepare(`CREATE TABLE IF NOT EXISTS delegations (
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

	// Ajouter des index pour améliorer les performances des requêtes
	_, err = database.Exec(`CREATE INDEX IF NOT EXISTS idx_timestamp ON delegations (timestamp)`)
	if err != nil {
		return nil, err
	}

	_, err = database.Exec(`CREATE INDEX IF NOT EXISTS idx_delegator ON delegations (delegator)`)
	if err != nil {
		return nil, err
	}

	return database, nil
}

// fetchDelegations récupère les délégations et les stocke dans la base de données
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

// getLastTimestamp récupère la dernière timestamp enregistrée
func getLastTimestamp(db *sql.DB) string {
	var lastTimestamp string
	row := db.QueryRow("SELECT timestamp FROM delegations ORDER BY timestamp DESC LIMIT 1")
	switch err := row.Scan(&lastTimestamp); err {
	case sql.ErrNoRows:
		return "2023-08-23T00:00:00Z"
	case nil:
		return lastTimestamp
	default:
		log.Println("Error getting last timestamp:", err)
		return "2023-08-23T00:00:00Z"
	}
}

// getDelegations récupère les délégations stockées et les renvoie sous forme de JSON
func getDelegations(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	year := r.URL.Query().Get("year")
	var rows *sql.Rows
	var err error

	if year != "" {
		rows, err = db.Query("SELECT timestamp, amount, delegator, level FROM delegations WHERE strftime('%Y', timestamp) = ? ORDER BY timestamp DESC", year)
	} else {
		rows, err = db.Query("SELECT timestamp, amount, delegator, level FROM delegations ORDER BY timestamp DESC")
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Database query error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var delegations []Delegation
	for rows.Next() {
		var d Delegation
		if err := rows.Scan(&d.Timestamp, &d.Amount, &d.Delegator, &d.Level); err != nil {
			http.Error(w, fmt.Sprintf("Database scan error: %v", err), http.StatusInternalServerError)
			return
		}
		delegations = append(delegations, d)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Rows iteration error: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string][]Delegation{"data": delegations})
}

// sendJSONResponse envoie une réponse JSON uniforme
func sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("JSON encoding error: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/xtz/delegations", func(w http.ResponseWriter, r *http.Request) {
		getDelegations(w, r, db)
	}).Methods("GET")

	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	wg.Add(1)
	go fetchDelegations(db, &wg, stopChan)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("Shutting down...")

		close(stopChan)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}
		wg.Wait() // Attend que fetchDelegations se termine
		fmt.Println("Server stopped")
	}()

	fmt.Println("Server is running on port 8000")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %s", err)
	}
}
