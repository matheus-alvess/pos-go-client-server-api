package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	apiUrl     = "http://localhost:8080/cotacao"
	apiTimeout = 300 * time.Millisecond
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout exceeded for call Dolar API")
		}
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	dolarResult := string(body)
	writeInFile(dolarResult)

	fmt.Printf("Dolar API Result value: %s\n", dolarResult)
}

func writeInFile(quotation string) {
	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data := fmt.Sprintf("Dólar: %s", quotation)

	_, err = file.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
}
