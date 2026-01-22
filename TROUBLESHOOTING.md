# 🔧 Troubleshooting Guide | সমস্যা সমাধান গাইড

## 🚨 Common Problems & Solutions

---

### ❌ Problem 1: Website Returns 403 Forbidden

**লক্ষণ | Symptoms:**
```
Error: status code: 403 Forbidden
```

**কারণ | Cause:**
Website blocking bot requests

**সমাধান | Solution:**

1. **User-Agent পরিবর্তন করুন** (`main.go` এ):
```go
req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
```

2. **আরো Headers যোগ করুন**:
```go
req.Header.Set("Accept-Language", "en-US,en;q=0.9")
req.Header.Set("Referer", "https://www.google.com")
```

3. **Delay যোগ করুন** requests এর মধ্যে:
```go
time.Sleep(5 * time.Second)
```

---

### ❌ Problem 2: All Prices Show 0.00

**লক্ষণ | Symptoms:**
```
Gold 22K: 0.00 | Gold 21K: 0.00 | Gold 18K: 0.00 | Silver: 0.00
Source: DEMO
```

**কারণ | Cause:**
HTML structure মিলছে না বা JavaScript দিয়ে data load হচ্ছে

**সমাধান | Solution:**

1. **Website inspect করুন**:
   - Browser এ goldr.org খুলুন
   - Right-click → Inspect
   - দাম কোথায় আছে দেখুন (class, id, tag)

2. **Code আপডেট করুন** যদি দাম specific class এ থাকে:
```go
// Example: যদি দাম "price-22k" class এ থাকে
doc.Find(".price-22k").Each(func(i int, s *goquery.Selection) {
    text := s.Text()
    if val := extractPriceValue(text); val > 0 {
        price.Gold22K = val
    }
})
```

3. **যদি JavaScript rendering দরকার হয়**:
   - `chromedp` বা `selenium` ব্যবহার করুন
   - নিচে example দেওয়া আছে

---

### ❌ Problem 3: Connection Timeout

**লক্ষণ | Symptoms:**
```
Error: context deadline exceeded
Error: i/o timeout
```

**সমাধান | Solution:**

1. **Timeout বাড়ান**:
```go
client := &http.Client{
    Timeout: 60 * time.Second,  // 30 থেকে 60 করুন
}
```

2. **Internet connection check করুন**:
```bash
ping www.goldr.org
curl -I https://www.goldr.org
```

3. **Proxy ব্যবহার করুন** (যদি দরকার হয়)

---

### ❌ Problem 4: "go: command not found"

**সমাধান | Solution:**

**Linux/Mac:**
```bash
# Go install করুন
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Windows:**
- https://go.dev/dl/ থেকে installer ডাউনলোড করুন

---

### ❌ Problem 5: CSV/JSON Files Not Created

**সমাধান | Solution:**

1. **Permission check করুন**:
```bash
ls -la gold_silver_prices.csv
chmod 644 gold_silver_prices.csv
```

2. **Directory writable কিনা দেখুন**:
```bash
touch test.txt
rm test.txt
```

---

## 🔍 Advanced Solutions

### JavaScript-rendered Website এর জন্য

যদি website JavaScript দিয়ে data load করে, **chromedp** ব্যবহার করুন:

#### Install:
```bash
go get -u github.com/chromedp/chromedp
```

#### Example Code:
```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/chromedp/chromedp"
)

func scrapePricesWithChrome() error {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    var price string
    
    err := chromedp.Run(ctx,
        chromedp.Navigate("https://www.goldr.org"),
        chromedp.WaitVisible(".price-class", chromedp.ByQuery),
        chromedp.Text(".price-class", &price, chromedp.ByQuery),
    )
    
    if err != nil {
        return err
    }
    
    log.Println("Price:", price)
    return nil
}
```

---

## 🧪 Testing Checklist

চালানোর আগে এই checklist follow করুন:

- [ ] Go installed (version 1.21+)
- [ ] Internet connection working
- [ ] `go mod download` চালানো হয়েছে
- [ ] `test_connection.go` successfully চলছে
- [ ] goldr.org browser এ accessible
- [ ] No CAPTCHA showing on website
- [ ] Firewall/antivirus blocking নেই

---

## 📊 Debugging Tips

### Enable Verbose Logging

`main.go` তে logging level বাড়ান:

```go
// HTML content দেখার জন্য
log.Printf("HTML Content: %s\n", string(bodyBytes))

// All matches দেখার জন্য
log.Printf("Found matches: %v\n", matches)
```

### Manual Testing

```bash
# Website manually fetch করুন
curl -v https://www.goldr.org

# Save HTML to file
curl https://www.goldr.org > page.html

# HTML inspect করুন
cat page.html | grep -i "gold\|22k\|price"
```

---

## 💡 Alternative Approaches

যদি scraping কাজ না করে:

### Option 1: API ব্যবহার করুন
কিছু website API দেয়। Check করুন goldr.org এর কোনো API আছে কিনা।

### Option 2: RSS Feed
Check করুন RSS feed আছে কিনা।

### Option 3: Browser Extension
Browser extension বানান যেটা page থেকে direct data নিবে।

---

## 📞 Need Help?

এখনো সমস্যা হলে:

1. **GitHub Issue** খুলুন detailed error message সহ
2. Output logs share করুন
3. Website URL verify করুন
4. Go version check করুন: `go version`

---

**Remember:** Web scraping ওয়েবসাইটের structure এর উপর নির্ভর করে। Website পরিবর্তন হলে code update করতে হবে।
