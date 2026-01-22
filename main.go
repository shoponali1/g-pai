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
	Gold22K     float64 `json:"gold_22k"`
	Gold21K     float64 `json:"gold_21k"`
	Gold18K     float64 `json:"gold_18k"`
	SilverPrice float64 `json:"silver_price"`
	Source      string  `json:"source"`
}

const (
	baseURL        = "https://www.goldr.org"
	scrapeInterval = 2 * time.Hour
	maxRetries     = 3
)

func main() {
	log.Println("===========================================")
	log.Println("Gold & Silver Price Scraper Started")
	log.Printf("Target: %s\n", baseURL)
	log.Printf("Scraping interval: %v\n", scrapeInterval)
	log.Println("===========================================")

	// প্রথমবার স্ক্র্যাপ করুন
	scrapeAndSave()

	// প্রতি ২ ঘন্টায় স্ক্র্যাপ করুন
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

	// Retry logic
	for i := 0; i < maxRetries; i++ {
		prices, err = scrapePrices()
		if err == nil {
			break
		}
		log.Printf("Attempt %d failed: %v\n", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(10 * time.Second)
		}
	}

	if err != nil {
		log.Printf("❌ All scraping attempts failed: %v\n", err)
		return
	}

	// CSV-তে সংরক্ষণ করুন
	if err := saveToCSV(prices); err != nil {
		log.Printf("❌ Error saving to CSV: %v\n", err)
	} else {
		log.Println("✅ Successfully saved to CSV")
	}

	// JSON-এ সংরক্ষণ করুন
	if err := saveToJSON(prices); err != nil {
		log.Printf("❌ Error saving to JSON: %v\n", err)
	} else {
		log.Println("✅ Successfully saved to JSON")
	}

	log.Printf("📊 Gold 22K: %.2f | Gold 21K: %.2f | Gold 18K: %.2f | Silver: %.2f\n",
		prices.Gold22K, prices.Gold21K, prices.Gold18K, prices.SilverPrice)
	log.Printf("Next scraping in %v at %s\n", scrapeInterval, time.Now().Add(scrapeInterval).Format("15:04:05"))
}

func scrapePrices() (*MetalPrice, error) {
	log.Println("🔍 Fetching data from website...")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Real browser headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	log.Printf("✅ Website responded: %d\n", resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	now := time.Now()
	price := &MetalPrice{
		Timestamp: now.Format(time.RFC3339),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Source:    baseURL,
	}

	log.Println("🔎 Searching for prices...")

	// খুঁজুন বিভিন্ন উপায়ে
	doc.Find("table tr, table td, div, span, p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		extractPricesFromText(text, price)
	})

	// যদি কিছু না পাওয়া যায়
	if price.Gold22K == 0 && price.Gold21K == 0 {
		log.Println("⚠️  Using demo data (couldn't extract from site)")
		price.Gold22K = 7850.50
		price.Gold21K = 7520.25
		price.Gold18K = 6430.75
		price.SilverPrice = 95.50
		price.Source = "DEMO"
	} else {
		log.Println("✅ Extracted prices successfully!")
	}

	return price, nil
}

func extractPricesFromText(text string, price *MetalPrice) {
	lower := strings.ToLower(text)

	// 22K
	if (strings.Contains(lower, "22") || strings.Contains(lower, "22k")) && 
	   strings.Contains(lower, "gold") && price.Gold22K == 0 {
		if val := extractPriceValue(text); val > 0 {
			price.Gold22K = val
			log.Printf("  Found 22K: %.2f\n", val)
		}
	}

	// 21K
	if (strings.Contains(lower, "21") || strings.Contains(lower, "21k")) && 
	   strings.Contains(lower, "gold") && price.Gold21K == 0 {
		if val := extractPriceValue(text); val > 0 {
			price.Gold21K = val
			log.Printf("  Found 21K: %.2f\n", val)
		}
	}

	// 18K
	if (strings.Contains(lower, "18") || strings.Contains(lower, "18k")) && 
	   strings.Contains(lower, "gold") && price.Gold18K == 0 {
		if val := extractPriceValue(text); val > 0 {
			price.Gold18K = val
			log.Printf("  Found 18K: %.2f\n", val)
		}
	}

	// Silver
	if strings.Contains(lower, "silver") && price.SilverPrice == 0 {
		if val := extractPriceValue(text); val > 0 {
			price.SilverPrice = val
			log.Printf("  Found Silver: %.2f\n", val)
		}
	}
}

func extractPriceValue(text string) float64 {
	re := regexp.MustCompile(`(\d{1,3}(?:,\d{3})*(?:\.\d{2})?|\d+\.\d{2}|\d+)`)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		cleaned := strings.ReplaceAll(match, ",", "")
		if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
			if (val >= 5000 && val <= 15000) || (val >= 50 && val <= 200) {
				return val
			}
		}
	}
	return 0
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
		header := []string{"Timestamp", "Date", "Time", "Gold_22K", "Gold_21K", "Gold_18K", "Silver", "Source"}
		writer.Write(header)
	}

	record := []string{
		price.Timestamp,
		price.Date,
		price.Time,
		fmt.Sprintf("%.2f", price.Gold22K),
		fmt.Sprintf("%.2f", price.Gold21K),
		fmt.Sprintf("%.2f", price.Gold18K),
		fmt.Sprintf("%.2f", price.SilverPrice),
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
