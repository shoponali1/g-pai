package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// এই স্ক্রিপ্ট চালিয়ে দেখুন ওয়েবসাইট accessible কিনা

func TestConnectionMain() {
	fmt.Println("Testing goldr.org connectivity...")
	fmt.Println("=====================================")

	url := "https://www.goldr.org"

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	fmt.Println("Sending request to:", url)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("❌ Error:", err)
	}
	defer resp.Body.Close()

	fmt.Printf("\n✅ Success!\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Content-Length: %s\n", resp.Header.Get("Content-Length"))

	fmt.Println("\n✅ Website is accessible!")
	fmt.Println("You can now run the main scraper with: go run main.go")
}
