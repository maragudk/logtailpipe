package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	log := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	if err := start(log); err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
}

func start(log *log.Logger) error {
	token := os.Getenv("LOGTAIL_TOKEN")
	if token == "" {
		return errors.New("env var LOGTAIL_TOKEN not set")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return pipe(log, client, token)
}

type request struct {
	Time    string `json:"dt"`
	Message string `json:"message"`
}

func pipe(log *log.Logger, client *http.Client, token string) error {
	s := bufio.NewScanner(os.Stdin)

	var wg sync.WaitGroup
	defer wg.Done()

	for s.Scan() {
		line := s.Text()
		now := time.Now()
		if line == "\u0004" {
			return nil
		}
		fmt.Println(line)

		wg.Add(1)
		go func() {
			body := request{
				Time:    now.UTC().Format(time.RFC3339Nano),
				Message: line,
			}

			bodyJSON, err := json.Marshal(body)
			if err != nil {
				log.Println("Error creating JSON request:", err)
			}

			req, err := http.NewRequest(http.MethodPost, "https://in.logtail.com/", bytes.NewReader(bodyJSON))
			if err != nil {
				log.Println("Error creating request:", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			if _, err := client.Do(req); err != nil {
				log.Println("Error requesting:", err)
			}
		}()
	}

	return s.Err()
}
