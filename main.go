package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
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
	Traditional float64 `json:"traditional_gold"`
	SilverPrice float64 `json:"silver_price"`
	Source      string  `json:"source"`
	Currency    string  `json:"currency"`
}

const (
	// bot.tools-time.com URL
	baseURL        = "https://bot.tools-time.com"
	scrapeInterval = 2 * time.Hour
	maxRetries     = 3
)

func main() {
	log.Println("===========================================")
	log.Println("Gold & Silver Price Scraper")
	log.Printf("Source: %s\n", baseURL)
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
		log.Println("⚠️  Using estimated prices")
		prices = getEstimatedPrices()
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

	log.Printf("📊 Gold 22K: %.2f | 21K: %.2f | 18K: %.2f | Silver: %.2f\n",
		prices.Gold22K, prices.Gold21K, prices.Gold18K, prices.SilverPrice)
}

func scrapePrices() (*MetalPrice, error) {
	log.Printf("🔍 Fetching data from %s...\n", baseURL)

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}

	// Try both HTTP and HTTPS
	urls := []string{
		"https://bot.tools-time.com",
		"http://bot.tools-time.com",
		"https://www.bot.tools-time.com",
	}

	var doc *goquery.Document
	var err error
	var workingURL string

	for _, url := range urls {
		req, e := http.NewRequest("GET", url, nil)
		if e != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,bn;q=0.8")

		resp, e := client.Do(req)
		if e != nil {
			log.Printf("  ❌ %s failed: %v\n", url, e)
			continue
		}

		if resp.StatusCode == 200 {
			doc, err = goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			if err == nil {
				workingURL = url
				log.Printf("  ✅ Connected to %s\n", url)
				break
			}
		}
		resp.Body.Close()
	}

	if doc == nil {
		return nil, fmt.Errorf("could not connect to any URL")
	}

	now := time.Now()
	price := &MetalPrice{
		Timestamp: now.Format(time.RFC3339),
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		Source:    workingURL,
		Currency:  "BDT",
	}

	log.Println("🔎 Searching for prices in HTML...")

	// Search in all possible elements
	doc.Find("table, tr, td, div, span, p, li").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			extractPrices(text, price)
		}
	})

	// Also check for data attributes
	doc.Find("[data-price], [data-gold], [data-silver]").Each(func(i int, s *goquery.Selection) {
		if val, exists := s.Attr("data-price"); exists {
			if num := parsePrice(val); num > 0 {
				if price.Gold22K == 0 {
					price.Gold22K = num
				}
			}
		}
	})

	if price.Gold22K == 0 {
		return nil, fmt.Errorf("no prices found on page")
	}

	log.Println("✅ Prices extracted successfully!")
	return price, nil
}

func extractPrices(text string, price *MetalPrice) {
	lower := strings.ToLower(text)

	// 22K Gold
	if (strings.Contains(lower, "22") || strings.Contains(lower, "22k")) && 
	   (strings.Contains(lower, "gold") || strings.Contains(lower, "সোনা") || strings.Contains(lower, "ক্যারেট")) {
		if val := parsePrice(text); val > 0 && price.Gold22K == 0 {
			price.Gold22K = val
			log.Printf("  Found 22K: %.2f\n", val)
		}
	}

	// 21K Gold
	if (strings.Contains(lower, "21") || strings.Contains(lower, "21k")) && 
	   (strings.Contains(lower, "gold") || strings.Contains(lower, "সোনা")) {
		if val := parsePrice(text); val > 0 && price.Gold21K == 0 {
			price.Gold21K = val
			log.Printf("  Found 21K: %.2f\n", val)
		}
	}

	// 18K Gold
	if (strings.Contains(lower, "18") || strings.Contains(lower, "18k")) && 
	   (strings.Contains(lower, "gold") || strings.Contains(lower, "সোনা")) {
		if val := parsePrice(text); val > 0 && price.Gold18K == 0 {
			price.Gold18K = val
			log.Printf("  Found 18K: %.2f\n", val)
		}
	}

	// Traditional/পুরাতন
	if strings.Contains(lower, "traditional") || strings.Contains(lower, "পুরাতন") || strings.Contains(lower, "পুরনো") {
		if val := parsePrice(text); val > 0 && price.Traditional == 0 {
			price.Traditional = val
			log.Printf("  Found Traditional: %.2f\n", val)
		}
	}

	// Silver/রুপা
	if strings.Contains(lower, "silver") || strings.Contains(lower, "রুপা") {
		if val := parsePrice(text); val > 0 && price.SilverPrice == 0 {
			price.SilverPrice = val
			log.Printf("  Found Silver: %.2f\n", val)
		}
	}
}

func parsePrice(text string) float64 {
	// Remove common non-numeric characters
	text = strings.ReplaceAll(text, "৳", "")
	text = strings.ReplaceAll(text, "tk", "")
	text = strings.ReplaceAll(text, "taka", "")
	
	// Find numbers
	re := regexp.MustCompile(`(\d{1,3}(?:,\d{3})*(?:\.\d{2})?|\d+(?:\.\d{2})?)`)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		cleaned := strings.ReplaceAll(match, ",", "")
		if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
			// Bangladesh gold price range: 50,000 - 200,000 BDT
			// Silver price range: 50,000 - 150,000 BDT/kg
			if val >= 1000 && val <= 250000 {
				return val
			}
		}
	}
	return 0
}

func getEstimatedPrices() *MetalPrice {
	now := time.Now()
	
	return &MetalPrice{
		Timestamp:   now.Format(time.RFC3339),
		Date:        now.Format("2006-01-02"),
		Time:        now.Format("15:04:05"),
		Gold22K:     78500.00,
		Gold21K:     75200.00,
		Gold18K:     64300.00,
		Traditional: 78500.00,
		SilverPrice: 95000.00,
		Source:      "Estimated (bot.tools-time.com unavailable)",
		Currency:    "BDT",
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
