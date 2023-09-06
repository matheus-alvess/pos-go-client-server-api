package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	quotationURL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	defaultPort = ":8080"
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
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println("GetQuotation handler finished")
	//ctx := r.Context()
	defer log.Println("GetQuotation handler finished")

	dolarQuotationResult, err := GetQuotationAPI()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: salvar na base com o timeout do contexto
	// TODO: ver como logar os timeouts do contexto

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dolarQuotationResult.USDBrl.Bid)
}

func GetQuotationAPI() (*dolarQuotation, error) {
	log.Println("GetQuotationAPI started")
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, 200 * time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quotationURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
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

	log.Println("GetQuotationAPI finished")
	return &dolarQuotationResult, nil
}