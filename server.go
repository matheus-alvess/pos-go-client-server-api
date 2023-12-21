package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	quotationURL      = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	defaultPort       = ":8080"
	dolarAPITimeout   = 200 * time.Millisecond
	insertDbTimeout   = 10 * time.Millisecond
	createTableScript = `
		CREATE TABLE IF NOT EXISTS quotations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value REAL,
			date_created TIMESTAMP                                 
		);
	`
)

type dolarQuotation struct {
	USDBrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", GetQuotationHandler)
	_ = http.ListenAndServe(defaultPort, nil)
}

func GetQuotationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dolarQuotationResult, err := getQuotationAPI(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveQuotationInDb(ctx, *dolarQuotationResult)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dolarQuotationResult.USDBrl.Bid)
}

func getQuotationAPI(ctx context.Context) (*dolarQuotation, error) {
	ctx, cancel := context.WithTimeout(ctx, dolarAPITimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quotationURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout exceeded for call Dolar API")
		}
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dolarQuotationResult dolarQuotation
	err = json.Unmarshal(body, &dolarQuotationResult)
	if err != nil {
		return nil, err
	}

	return &dolarQuotationResult, nil
}

func saveQuotationInDb(ctx context.Context, dolarQuotation dolarQuotation) error {
	ctx, cancel := context.WithTimeout(ctx, insertDbTimeout)
	defer cancel()

	db, err := sql.Open("sqlite3", "quotation.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(createTableScript)
	if err != nil {
		return err
	}

	if err := insertDataWithTimeout(ctx, db, dolarQuotation); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout exceeded for insert in SQlLite")
		}
		return err
	}

	return nil
}

func insertDataWithTimeout(ctx context.Context, db *sql.DB, dolarQuotation dolarQuotation) error {
	insertData := `
		INSERT INTO quotations (name, value, date_created) VALUES (?, ?, ?);
	`
	result, err := db.ExecContext(ctx, insertData, dolarQuotation.USDBrl.Name, dolarQuotation.USDBrl.Bid, time.Now())
	if err != nil {
		return err
	}

	lastIDInserted, _ := result.LastInsertId()
	fmt.Printf("ID of new register: %d\n", lastIDInserted)
	return nil
}
