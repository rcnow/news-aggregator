# News Aggregator

A simple news aggregator built with Go that fetches news from various RSS feeds and displays them on a web interface.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/dl/) (version 1.21 or later)
- A web browser

## Getting Started

Follow these steps to set up and run the project locally.

### 1. Clone the Repository

Clone the repository to your local machine using the following command:

```bash
git clone https://github.com/rcnow/news-aggregator.git
```
### 2. Navigate to the Project Directory

Change into the project directory:

```bash
cd news-aggregator
```
### 3. Add Your RSS Links

Open the `handlers.go` file and add your RSS feed URLs to the `feeds` slice. For example:

```go
feeds := []string{
    "https://cointelegraph.com/rss", //Example one
    "https://bitcoinmagazine.com/feed", //Example two
    "https://example.com/feed", // Add your custom feed here
}
```

You should see output indicating that the web server is starting:

### 4. Access the Application

Open your web browser and navigate to:
http://localhost:8080