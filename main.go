package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MetalPrice struct {
	Timestamp   string  `json:"timestamp"`
	Date        string  `json:"date"`
	Time        string  `json:"time"`
	Gold22K     float64 `json:"gold_22k_per_bhori"`
	Gold21K     float64 `json:"gold_21k_per_bhori"`
	Gold18K     float64 `json:"gold_18k_per_bhori"`
	Gold24K     float64 `json:"gold_24k_per_bhori"`
	Traditional float64 `json:"traditional_gold_per_bhori"`
	SilverPrice float64 `json:"silver_per_kg"`
	Source      string  `json:"source"`
	Currency    string  `json:"currency"`
}

const (
	// BAJUS - Bangladesh Jewellers Association (Official)
	bajusURL = "http://www.bajus.org"
	
	// Alternative: goldprice.org API endpoint
	goldPriceAPI = "https://www.goldprice.org/api/rates"
	
	scrapeInterval = 2 * time.Hour
	maxRetries     = 3
)

func main() {
	log.Println("===========================================")
	log.Println("Bangladesh Gold & Silver Price Scraper")
	log.Println("Source: BAJUS + International Prices")
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

	// Try BAJUS first
	for i := 0; i < maxRetries; i++ {
		prices, err = scrapePrices()
		if err == nil && prices.Gold22K > 0 {
			break
		}
		log.Printf("Attempt %d: %v\n", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(5 * time.Second)
		}
	}

	// If BAJUS fails, use international prices
	if err != nil || prices.Gold22K == 0 {
		log.Println("Trying international prices...")
		prices = getInternationalPrices()
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

	log.Printf("📊 Gold 22K: %.2f BDT/bhori | Gold 21K: %.2f | Gold 18K: %.2f | Silver: %.2f BDT/kg\n",
		prices.Gold22K, prices.Gold21K, prices.Gold18K, prices.SilverPrice)
}

func scrapePrices() (*MetalPrice, error) {
	log.Println("🔍 Fetching BAJUS Bangladesh prices...")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", bajusURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	price := &MetalPrice{
		Timestamp: now.Format(time.RFC3339),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Source:    "BAJUS",
		Currency:  "BDT",
	}

	// Find prices in tables
	doc.Find("table tr, div, span, p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		extractBDPrices(text, price)
	})

	if price.Gold22K == 0 {
		return nil, fmt.Errorf("no prices found")
	}

	log.Println("✅ BAJUS prices extracted")
	return price, nil
}

func extractBDPrices(text string, price *MetalPrice) {
	text = strings.ToLower(text)

	// 22 Carat
	if strings.Contains(text, "22") && (strings.Contains(text, "carat") || strings.Contains(text, "k")) {
		if val := extractBDPrice(text); val > 0 && price.Gold22K == 0 {
			price.Gold22K = val
			log.Printf("  22K: %.2f BDT/bhori\n", val)
		}
	}

	// 21 Carat
	if strings.Contains(text, "21") && (strings.Contains(text, "carat") || strings.Contains(text, "k")) {
		if val := extractBDPrice(text); val > 0 && price.Gold21K == 0 {
			price.Gold21K = val
			log.Printf("  21K: %.2f BDT/bhori\n", val)
		}
	}

	// 18 Carat
	if strings.Contains(text, "18") && (strings.Contains(text, "carat") || strings.Contains(text, "k")) {
		if val := extractBDPrice(text); val > 0 && price.Gold18K == 0 {
			price.Gold18K = val
			log.Printf("  18K: %.2f BDT/bhori\n", val)
		}
	}

	// Traditional
	if strings.Contains(text, "traditional") || strings.Contains(text, "পুরাতন") {
		if val := extractBDPrice(text); val > 0 && price.Traditional == 0 {
			price.Traditional = val
			log.Printf("  Traditional: %.2f BDT/bhori\n", val)
		}
	}

	// Silver
	if strings.Contains(text, "silver") || strings.Contains(text, "রুপা") {
		if val := extractBDPrice(text); val > 0 && price.SilverPrice == 0 {
			price.SilverPrice = val
			log.Printf("  Silver: %.2f BDT/kg\n", val)
		}
	}
}

func extractBDPrice(text string) float64 {
	// BD prices: 50,000 - 150,000 BDT range
	re := regexp.MustCompile(`(\d{1,3}(?:,\d{3})*(?:\.\d{2})?|\d+)`)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		cleaned := strings.ReplaceAll(match, ",", "")
		if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
			// Reasonable BD price range
			if val >= 50000 && val <= 200000 {
				return val
			}
		}
	}
	return 0
}

func getInternationalPrices() *MetalPrice {
	log.Println("🌍 Using international prices...")

	now := time.Now()
	
	// Current approximate prices (will be updated with real scraping)
	// 1 USD = ~110 BDT (approximate)
	// International gold ~$2050/oz = ~$66/gram
	// 1 bhori = 11.664 grams
	// So gold per bhori ≈ $66 * 11.664 * 110 BDT
	
	goldPerBhoriUSD := 2050.0 / 31.1035 * 11.664 // ~$770 per bhori
	bdtRate := 110.0 // USD to BDT
	
	price := &MetalPrice{
		Timestamp:   now.Format(time.RFC3339),
		Date:        now.Format("2006-01-02"),
		Time:        now.Format("15:04:05"),
		Gold22K:     goldPerBhoriUSD * 0.9167 * bdtRate, // 22/24
		Gold21K:     goldPerBhoriUSD * 0.875 * bdtRate,  // 21/24
		Gold18K:     goldPerBhoriUSD * 0.75 * bdtRate,   // 18/24
		Gold24K:     goldPerBhoriUSD * bdtRate,
		Traditional: goldPerBhoriUSD * 0.9167 * bdtRate,
		SilverPrice: 25.0 * 11.664 * bdtRate * 1000 / 11.664, // ~per kg
		Source:      "International (Estimated)",
		Currency:    "BDT",
	}

	log.Printf("  Estimated 22K: %.2f BDT/bhori\n", price.Gold22K)
	return price
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
		header := []string{"Timestamp", "Date", "Time", "Gold_22K", "Gold_21K", "Gold_18K", "Traditional", "Silver", "Currency", "Source"}
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
		fmt.Sprintf("%.2f", price.SilverPrice),
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
