# Bangladesh Gold & Silver Price Scraper

A Go-based automated scraper that fetches real-time gold and silver prices in Bangladesh from [goldr.org](https://www.goldr.org/).

## Features

- **Automated Scraping**: Runs every 2 hours (via GitHub Actions).
- **Data Source**: [goldr.org](https://www.goldr.org/price.js)
- **Detailed Prices**:
  - **Gold**: 22K, 21K, 18K, Traditional (Sanatan)
  - **Silver**: 22K, 21K, 18K, Traditional (Sanatan)
- **Data Format**: Saves data to both `CSV` and `JSON` formats in the `Data/` folder.

## Data Fields

The scraper collects the following data points (price per Bhori in BDT):

| Field | Description |
|-------|-------------|
| `Gold_22K` | Price of 22 Karat Gold |
| `Gold_21K` | Price of 21 Karat Gold |
| `Gold_18K` | Price of 18 Karat Gold |
| `Traditional_Gold` | Price of Traditional Gold |
| `Silver_22K` | Price of 22 Karat Silver |
| `Silver_21K` | Price of 21 Karat Silver |
| `Silver_18K` | Price of 18 Karat Silver |
| `Traditional_Silver` | Price of Traditional Silver |

## Usage

### Run Locally

1. **Install Go**: Ensure Go 1.21+ is installed.
2. **Run Scraper**:
   ```bash
   go run main.go
   ```
3. **Check Data**: Data will be saved to `gold_silver_prices.csv` and `.json`.

### GitHub Actions

The repository includes a GitHub Actions workflow (`.github/workflows/auto_update.yml`) that:
1. Runs automatically on a schedule.
2. Scrapes the data.
3. Commits the updated `CSV` and `JSON` files back to the repository in the `Data/` folder.
