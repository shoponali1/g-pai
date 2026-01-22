# 🪙 Gold & Silver Price Scraper | সোনা ও রুপার দাম স্ক্র্যাপার

একটি **Go-ভিত্তিক ওয়েব স্ক্র্যাপার** যা goldr.org থেকে সোনা ও রুপার দাম সংগ্রহ করে এবং **CSV ও JSON** ফাইলে সংরক্ষণ করে।

A **Go-based web scraper** that collects gold and silver prices from goldr.org and saves them to **CSV and JSON** files.

---

## 🎯 বৈশিষ্ট্য | Features

| Feature | বিবরণ | Description |
|---------|-------|-------------|
| ✅ | **22K, 21K, 18K সোনার দাম** সংগ্রহ | Collects 22K, 21K, 18K gold prices |
| ✅ | **রুপা/সিলভার** দাম সংগ্রহ | Collects silver prices |
| ⏰ | **প্রতি ২ ঘন্টায়** স্বয়ংক্রিয় আপডেট | Auto-updates every 2 hours |
| 💾 | **CSV ফাইল** সংরক্ষণ | Saves to CSV file |
| 💾 | **JSON ফাইল** সংরক্ষণ | Saves to JSON file |
| 🔄 | **Retry Logic** - ব্যর্থ হলে পুনঃচেষ্টা | Retry on failure |
| 📊 | **Detailed Logging** - সব কিছু লগ করে | Detailed logging |

---

## 📦 ইনস্টলেশন | Installation

### প্রয়োজনীয় | Prerequisites

- **Go 1.21+** ইনস্টল করা থাকতে হবে
- ইন্টারনেট সংযোগ

### ⚡ Quick Start

```bash
# 1️⃣ Repository clone করুন
git clone https://github.com/yourusername/gold-silver-scraper.git
cd gold-silver-scraper

# 2️⃣ Dependencies ডাউনলোড করুন
go mod download

# 3️⃣ প্রথমে connection টেস্ট করুন (Optional)
go run test_connection.go

# 4️⃣ Scraper চালান
go run main.go
```

---

## 🧪 টেস্ট করার আগে | Before Running

### ✅ Step 1: Connection Test

প্রথমে চেক করুন ওয়েবসাইট access করা যাচ্ছে কিনা:

```bash
go run test_connection.go
```

**Expected Output:**
```
Testing goldr.org connectivity...
=====================================
Sending request to: https://www.goldr.org

✅ Success!
Status Code: 200
Status: 200 OK
✅ Website is accessible!
```

### ✅ Step 2: Run Scraper

```bash
go run main.go
```

**Expected Output:**
```
===========================================
Gold & Silver Price Scraper Started
Target: https://www.goldr.org
Scraping interval: 2h0m0s
===========================================

--- Starting new scraping cycle ---
🔍 Fetching data from website...
✅ Website responded: 200
🔎 Searching for prices...
  Found 22K: 7850.50
  Found 21K: 7520.25
  Found 18K: 6430.75
  Found Silver: 95.50
✅ Extracted prices successfully!
✅ Successfully saved to CSV
✅ Successfully saved to JSON
📊 Gold 22K: 7850.50 | Gold 21K: 7520.25 | Gold 18K: 6430.75 | Silver: 95.50
```

---

## 📁 Output Files | আউটপুট ফাইল

### 📄 CSV Format

**File:** `gold_silver_prices.csv`

```csv
Timestamp,Date,Time,Gold_22K,Gold_21K,Gold_18K,Silver,Source
2024-01-22T10:30:00+06:00,2024-01-22,10:30:00,7850.50,7520.25,6430.75,95.50,https://www.goldr.org
2024-01-22T12:30:00+06:00,2024-01-22,12:30:00,7855.00,7525.00,6435.00,96.00,https://www.goldr.org
```

### 📄 JSON Format

**File:** `gold_silver_prices.json`

```json
[
  {
    "timestamp": "2024-01-22T10:30:00+06:00",
    "date": "2024-01-22",
    "time": "10:30:00",
    "gold_22k": 7850.50,
    "gold_21k": 7520.25,
    "gold_18k": 6430.75,
    "silver_price": 95.50,
    "source": "https://www.goldr.org"
  }
]
```

---

## ⚙️ Configuration | কনফিগারেশন

### Scraping Interval পরিবর্তন করুন

`main.go` ফাইলে:

```go
const scrapeInterval = 2 * time.Hour  // ২ ঘন্টা (Default)
```

**অন্য অপশন:**
```go
const scrapeInterval = 1 * time.Hour      // ১ ঘন্টা
const scrapeInterval = 30 * time.Minute   // ৩০ মিনিট
const scrapeInterval = 6 * time.Hour      // ৬ ঘন্টা
```

### URL পরিবর্তন করুন

```go
const baseURL = "https://www.goldr.org"  // আপনার URL দিন
```

---

## 🐳 Docker দিয়ে চালান (Optional)

```bash
# Build
docker build -t gold-scraper .

# Run
docker run -v $(pwd)/data:/app/data gold-scraper
```

---

## 🖥️ Background Service হিসেবে চালান

### Linux (systemd)

```bash
# 1. Binary তৈরি করুন
go build -o scraper main.go

# 2. Service file কপি করুন
sudo cp gold-scraper.service /etc/systemd/system/

# 3. Service start করুন
sudo systemctl start gold-scraper
sudo systemctl enable gold-scraper

# 4. Status check করুন
sudo systemctl status gold-scraper
```

### Windows (Task Scheduler)

1. `scraper.exe` build করুন: `go build -o scraper.exe main.go`
2. Task Scheduler খুলুন
3. "Create Basic Task" করুন
4. Trigger: "At startup"
5. Action: Start the program `scraper.exe`

---

## ⚠️ Important Notes | গুরুত্বপূর্ণ নোট

### 🔍 Scraping সম্পর্কে

- ওয়েবসাইট যদি **JavaScript দিয়ে ডেটা লোড** করে, তাহলে এই scraper কাজ নাও করতে পারে
- সেক্ষেত্রে **Selenium** বা **Playwright** ব্যবহার করতে হবে
- ওয়েবসাইটের **HTML structure** পরিবর্তন হলে scraper আপডেট করতে হবে

### ✅ যদি কাজ না করে

1. **Connection test** চালান: `go run test_connection.go`
2. Website কি accessible?
3. Website কি **CAPTCHA** বা **anti-bot** protection ব্যবহার করছে?
4. HTML structure পরীক্ষা করুন

### 📝 Demo Data

যদি scraper actual ডেটা extract করতে না পারে, এটি **demo/test ডেটা** ব্যবহার করবে এবং Source column-এ "DEMO" লিখবে।

---

## 🛠️ Troubleshooting

### Problem: "Status code: 403"
**Solution:** Website blocking করছে। User-Agent বা headers পরিবর্তন করুন।

### Problem: সব দাম 0 দেখাচ্ছে
**Solution:** HTML structure চেক করুন। Website inspect করে দেখুন কোথায় দাম আছে।

### Problem: "Connection timeout"
**Solution:** Internet connection চেক করুন বা timeout বাড়ান।

---

## 📊 Project Structure

```
gold-silver-scraper/
├── main.go                     # মূল scraper
├── test_connection.go          # Connection tester
├── go.mod                      # Dependencies
├── README.md                   # Documentation
├── Dockerfile                  # Docker support
├── .gitignore                  # Git ignore
├── gold-scraper.service        # Linux service
└── .github/
    └── workflows/
        └── build.yml           # CI/CD
```

---

## 📚 Dependencies

```
github.com/PuerkitoBio/goquery v1.8.1
```

---

## 🤝 Contributing | অবদান

Pull requests স্বাগতম! বড় পরিবর্তনের জন্য প্রথমে একটি **issue** খুলুন।

---

## 📄 License

MIT License - শিক্ষামূলক উদ্দেশ্যে ব্যবহার করুন।

---

## ⚖️ Legal Notice

এই scraper **শিক্ষামূলক উদ্দেশ্যে**। ওয়েবসাইটের **Terms of Service** এবং **robots.txt** মেনে চলুন।

---

## 📞 Support

সমস্যা হলে **GitHub Issues** তে রিপোর্ট করুন।

---

**Made with ❤️ for the Go community**
