package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"io"
)

type MetalPrice struct {
	Timestamp         string  `json:"timestamp"`
	Date              string  `json:"date"`
	Time              string  `json:"time"`
	Gold22K           float64 `json:"gold_22k"`
	Gold21K           float64 `json:"gold_21k"`
	Gold18K           float64 `json:"gold_18k"`
	Traditional       float64 `json:"traditional_gold"`
	Silver22K         float64 `json:"silver_22k"`
	Silver21K         float64 `json:"silver_21k"`
	Silver18K         float64 `json:"silver_18k"`
	SilverTraditional float64 `json:"silver_traditional"`
	Source            string  `json:"source"`
	Currency          string  `json:"currency"`
}

const (
	targetURL      = "https://www.goldr.org/price.js"
	scrapeInterval = 2 * time.Hour
	maxRetries     = 3
)

func main() {
	log.Println("===========================================")
	log.Println("Gold & Silver Price Scraper")
	log.Printf("Source: %s\n", targetURL)
	log.Printf("Scraping interval: %v\n", scrapeInterval)
	log.Println("===========================================")

	scrapeAndSave()

	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	for range ticker.C {
		scrapeAndSave()
	}
}

func scrapeAndSave() {
	log.Println("\n--- Starting new scraping cycle ---")

	var prices *MetalPrice
	var err error

	for i := 0; i < maxRetries; i++ {
		prices, err = scrapePrices()
		if err == nil && prices.Gold22K > 0 {
			break
		}
		log.Printf("Attempt %d: %v\n", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(10 * time.Second)
		}
	}

	if err != nil || prices == nil || prices.Gold22K == 0 {
		log.Println("⚠️  Using estimated prices (Scrape failed)")
		if prices == nil {
			prices = getEstimatedPrices()
		}
	}

	if err := saveToCSV(prices); err != nil {
		log.Printf("❌ Error saving to CSV: %v\n", err)
	} else {
		log.Println("✅ Successfully saved to CSV")
	}

	if err := saveToJSON(prices); err != nil {
		log.Printf("❌ Error saving to JSON: %v\n", err)
	} else {
		log.Println("✅ Successfully saved to JSON")
	}

	log.Printf("📊 Gold 22K: %.2f | Silver 22K: %.2f\n",
		prices.Gold22K, prices.Silver22K)
}

// Structs for parsing the JS JSON data
type GoldItem struct {
	N     string  `json:"n"`
	BvRaw float64 `json:"bv_raw"`
}

type SilverItem struct {
	N     string  `json:"n"`
	BvRaw float64 `json:"bv_raw"`
}

func scrapePrices() (*MetalPrice, error) {
	log.Printf("🔍 Fetching data from %s...\n", targetURL)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	content := string(body)

	// Extract Gold Data
	goldRegex := regexp.MustCompile(`const\s+GoldrPriceTable_goldData\s*=\s*(\[.*?\]);`)
	goldMatches := goldRegex.FindStringSubmatch(content)
	if len(goldMatches) < 2 {
		return nil, fmt.Errorf("could not find gold data in JS")
	}

	// Extract Silver Data
	silverRegex := regexp.MustCompile(`const\s+GoldrPriceTable_silverData\s*=\s*(\[.*?\]);`)
	silverMatches := silverRegex.FindStringSubmatch(content)
	if len(silverMatches) < 2 {
		return nil, fmt.Errorf("could not find silver data in JS")
	}

	var goldItems []GoldItem
	if err := json.Unmarshal([]byte(goldMatches[1]), &goldItems); err != nil {
		return nil, fmt.Errorf("failed to parse gold json: %v", err)
	}

	var silverItems []SilverItem
	if err := json.Unmarshal([]byte(silverMatches[1]), &silverItems); err != nil {
		return nil, fmt.Errorf("failed to parse silver json: %v", err)
	}

	if len(goldItems) < 4 {
		return nil, fmt.Errorf("unexpected number of gold items: %d", len(goldItems))
	}

	now := time.Now()
	price := &MetalPrice{
		Timestamp: now.Format(time.RFC3339),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Source:    targetURL,
		Currency:  "BDT",
	}

	// Gold Mapping
	price.Gold22K = goldItems[0].BvRaw
	price.Gold21K = goldItems[1].BvRaw
	price.Gold18K = goldItems[2].BvRaw
	price.Traditional = goldItems[3].BvRaw

	// Silver Mapping
	// Based on array order in JS: 0=22K, 1=21K, 2=18K, 3=Traditional
	if len(silverItems) >= 4 {
		price.Silver22K = silverItems[0].BvRaw
		price.Silver21K = silverItems[1].BvRaw
		price.Silver18K = silverItems[2].BvRaw
		price.SilverTraditional = silverItems[3].BvRaw
	} else if len(silverItems) > 0 {
		price.Silver22K = silverItems[0].BvRaw // Fallback
	}

	log.Println("✅ Prices extracted successfully!")
	return price, nil
}

// Helper not needed anymore for parsing simple HTML text, but kept if we need utility later
func parsePrice(text string) float64 {
	return 0
}

func getEstimatedPrices() *MetalPrice {
	now := time.Now()

	return &MetalPrice{
		Timestamp:         now.Format(time.RFC3339),
		Date:              now.Format("2006-01-02"),
		Time:              now.Format("15:04:05"),
		Gold22K:           252000.00, // Updated to approx market
		Gold21K:           240000.00,
		Gold18K:           206000.00,
		Traditional:       169000.00,
		Silver22K:         6800.00,
		Silver21K:         6500.00,
		Silver18K:         5600.00,
		SilverTraditional: 4200.00,
		Source:            "Estimated (Scrape Failed)",
		Currency:          "BDT",
	}
}

func saveToCSV(price *MetalPrice) error {
	filename := "gold_silver_prices.csv"
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := []string{
			"Timestamp", "Date", "Time",
			"Gold_22K", "Gold_21K", "Gold_18K", "Traditional_Gold",
			"Silver_22K", "Silver_21K", "Silver_18K", "Traditional_Silver",
			"Currency", "Source",
		}
		writer.Write(header)
	}

	record := []string{
		price.Timestamp,
		price.Date,
		price.Time,
		fmt.Sprintf("%.2f", price.Gold22K),
		fmt.Sprintf("%.2f", price.Gold21K),
		fmt.Sprintf("%.2f", price.Gold18K),
		fmt.Sprintf("%.2f", price.Traditional),
		fmt.Sprintf("%.2f", price.Silver22K),
		fmt.Sprintf("%.2f", price.Silver21K),
		fmt.Sprintf("%.2f", price.Silver18K),
		fmt.Sprintf("%.2f", price.SilverTraditional),
		price.Currency,
		price.Source,
	}

	return writer.Write(record)
}

func saveToJSON(price *MetalPrice) error {
	filename := "gold_silver_prices.json"
	var prices []MetalPrice

	if data, err := os.ReadFile(filename); err == nil {
		json.Unmarshal(data, &prices)
	}

	prices = append(prices, *price)
	data, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
