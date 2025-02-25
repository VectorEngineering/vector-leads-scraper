## Advanced Scraping Techniques

To maximize the effectiveness of the Google Maps Scraper, consider these advanced techniques:

```mermaid
graph TB
    subgraph "Advanced Techniques"
        direction TB
        geotargeting[Geo-targeting<br>Strategies] --> precision[Improved Precision]
        query[Query<br>Optimization] --> coverage[Better Coverage]
        session[Session<br>Management] --> stability[Enhanced Stability]
        parsing[Advanced<br>Parsing] --> data[Rich Data Extraction]
    end
    
    subgraph "Implementation Methods"
        direction TB
        grid[Geographic Grid<br>Search Pattern] --> geotargeting
        keywords[Keyword<br>Combination Matrix] --> query
        cookies[Cookie<br>Management] --> session
        js[JavaScript<br>DOM Extraction] --> parsing
    end
    
    subgraph "Benefits"
        precision --> results[Comprehensive<br>Results]
        coverage --> results
        stability --> efficiency[Operational<br>Efficiency]
        data --> quality[Higher Quality<br>Data]
    end
    
    classDef technique fill:#f9f,stroke:#333,stroke-width:2px
    classDef method fill:#bbf,stroke:#333,stroke-width:2px
    classDef benefit fill:#bfb,stroke:#333,stroke-width:2px
    class geotargeting,query,session,parsing technique
    class grid,keywords,cookies,js method
    class precision,coverage,stability,data,results,efficiency,quality benefit
```

### 1. Geographic Grid Search Pattern

For thorough coverage of a geographic area, implement a grid-based search pattern:

```python
# Example implementation in Python (pseudocode)
def generate_grid_points(center_lat, center_lng, radius_km, grid_size):
    """Generate grid points covering a circular area."""
    # Convert radius from km to degrees (approximate)
    radius_lat = radius_km / 111.0  # 1 degree ~ 111km
    radius_lng = radius_km / (111.0 * math.cos(math.radians(center_lat)))
    
    # Calculate grid boundaries
    min_lat = center_lat - radius_lat
    max_lat = center_lat + radius_lat
    min_lng = center_lng - radius_lng
    max_lng = center_lng + radius_lng
    
    # Calculate step size
    lat_step = 2 * radius_lat / grid_size
    lng_step = 2 * radius_lng / grid_size
    
    # Generate grid points
    grid_points = []
    for i in range(grid_size):
        for j in range(grid_size):
            lat = min_lat + i * lat_step
            lng = min_lng + j * lng_step
            
            # Check if point is within the circular area
            dist = haversine_distance(center_lat, center_lng, lat, lng)
            if dist <= radius_km:
                grid_points.append((lat, lng))
    
    return grid_points

# Example usage
grid_points = generate_grid_points(37.7749, -122.4194, 10, 5)
for lat, lng in grid_points:
    # Run the scraper with this geo point
    # ./google-maps-scraper -geo f"{lat},{lng}" -radius 2000 -input queries.txt
```

### 2. Query Expansion and Refinement

Improve search coverage by automatically expanding and refining search queries:

```go
// Example query expansion in Go
func expandQueries(baseQueries []string, modifiers []string) []string {
    var expandedQueries []string
    
    for _, query := range baseQueries {
        // Add the base query
        expandedQueries = append(expandedQueries, query)
        
        // Add modified queries
        for _, modifier := range modifiers {
            expandedQueries = append(expandedQueries, query+" "+modifier)
        }
    }
    
    return expandedQueries
}

// Example usage
baseQueries := []string{"restaurants", "cafes", "bars"}
modifiers := []string{"best", "open now", "top rated", "cheap", "family friendly"}
expanded := expandQueries(baseQueries, modifiers)

// Result will include:
// restaurants, restaurants best, restaurants open now, ...
// cafes, cafes best, cafes open now, ...
// etc.
```

### 3. Browser Fingerprint Randomization

Minimize detection by randomizing browser fingerprints:

```go
// Example browser fingerprint randomization in Go
func randomUserAgent() string {
    userAgents := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
        "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
    }
    
    return userAgents[rand.Intn(len(userAgents))]
}

// Apply to browser options
func getBrowserOptions(proxy string) []chromedp.ExecAllocatorOption {
    options := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.UserAgent(randomUserAgent()),
        chromedp.WindowSize(rand.Intn(100)+900, rand.Intn(100)+800),
    )
    
    if proxy != "" {
        options = append(options, chromedp.ProxyServer(proxy))
    }
    
    return options
}
```

### 4. Advanced HTML Parsing

Extract hidden or dynamic data from Google Maps listings:

```go
// Extract data from Google Maps using advanced selectors
func extractAdvancedData(ctx context.Context, browser *Browser) (*Business, error) {
    var business Business
    
    // Extract structured data (JSON-LD)
    var jsonLD string
    err := chromedp.Run(browser.ctx,
        chromedp.EvaluateAsDevTools(`
            Array.from(document.querySelectorAll('script[type="application/ld+json"]'))
                .map(el => el.textContent)
                .join('\n')
        `, &jsonLD),
    )
    if err != nil {
        return nil, err
    }
    
    // Process JSON-LD
    if jsonLD != "" {
        // JSON-LD often contains rich business information
        // Parse and extract relevant fields
    }
    
    // Extract data from DOM
    var phoneNodes []*cdp.Node
    err = chromedp.Run(browser.ctx,
        chromedp.Nodes(`button[data-item-id="phone"] span`, &phoneNodes, chromedp.ByQueryAll),
    )
    if err == nil && len(phoneNodes) > 0 {
        // Extract phone number
    }
    
    // Extract hidden data using JavaScript
    var hiddenData string
    err = chromedp.Run(browser.ctx,
        chromedp.EvaluateAsDevTools(`
            // Attempt to access Google Maps internal data
            // This is advanced and may break with Google updates
            (function() {
                try {
                    // Search for specific global objects
                    for (let key in window) {
                        if (key.startsWith('_') && window[key] && window[key].hasOwnProperty('entities')) {
                            const entities = window[key].entities;
                            // Look for business data
                            for (let id in entities) {
                                const entity = entities[id];
                                if (entity && entity.hasOwnProperty('address')) {
                                    return JSON.stringify(entity);
                                }
                            }
                        }
                    }
                } catch (e) {
                    return '';
                }
                return '';
            })()
        `, &hiddenData),
    )
    
    if hiddenData != "" {
        // Parse the extracted data
    }
    
    return &business, nil
}
```

### 5. Review Extraction Optimization

Extract more reviews by handling pagination:

```go
// Extract reviews with pagination
func extractReviews(ctx context.Context, browser *Browser) ([]Review, error) {
    var reviews []Review
    
    // Click on reviews section
    err := chromedp.Run(browser.ctx,
        chromedp.Click("button[data-tab-index='1']", chromedp.ByQuery),
        chromedp.Sleep(1*time.Second),
    )
    if err != nil {
        return nil, err
    }
    
    // Function to extract visible reviews
    extractVisibleReviews := func() error {
        var reviewsData string
        err := chromedp.Run(browser.ctx,
            chromedp.EvaluateAsDevTools(`
                Array.from(document.querySelectorAll('.jftiEf')).map(el => {
                    const author = el.querySelector('.d4r55')?.textContent || '';
                    const ratingEl = el.querySelector('.kvMYJc');
                    const rating = ratingEl ? parseFloat(ratingEl.getAttribute('aria-label').match(/(\d+\.\d+)/)[1]) : 0;
                    const text = el.querySelector('.wiI7pd')?.textContent || '';
                    const date = el.querySelector('.rsqaWe')?.textContent || '';
                    
                    return { author, rating, text, date };
                })
            `, &reviewsData),
        )
        if err != nil {
            return err
        }
        
        // Parse the extracted reviews
        var batchReviews []Review
        err = json.Unmarshal([]byte(reviewsData), &batchReviews)
        if err != nil {
            return err
        }
        
        reviews = append(reviews, batchReviews...)
        return nil
    }
    
    // Extract initial reviews
    if err := extractVisibleReviews(); err != nil {
        return nil, err
    }
    
    // Scroll and extract more reviews (up to a limit)
    maxScrolls := 5
    for i := 0; i < maxScrolls; i++ {
        // Scroll down in the reviews section
        err := chromedp.Run(browser.ctx,
            chromedp.EvaluateAsDevTools(`
                const reviewsContainer = document.querySelector('.m6QErb[aria-label*="reviews"]');
                if (reviewsContainer) {
                    reviewsContainer.scrollTop = reviewsContainer.scrollHeight;
                    return true;
                }
                return false;
            `, nil),
            chromedp.Sleep(2*time.Second), // Wait for new reviews to load
        )
        if err != nil {
            break
        }
        
        // Extract newly visible reviews
        prevCount := len(reviews)
        if err := extractVisibleReviews(); err != nil {
            break
        }
        
        // If no new reviews were added, we've reached the end
        if len(reviews) == prevCount {
            break
        }
    }
    
    return reviews, nil
}
```

### 6. Smart Proxy Rotation

Implement intelligent proxy rotation based on success rates:

```go
// ProxyManager with smart rotation
type ProxyManager struct {
    proxies       []Proxy
    currentIndex  int
    successRates  map[string]float64
    totalRequests map[string]int
    mu            sync.Mutex
}

type Proxy struct {
    URL      string
    Weight   int
    Tags     []string
    Disabled bool
}

// GetProxy returns the next proxy based on success rates
func (pm *ProxyManager) GetProxy() string {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    // Filter enabled proxies
    var enabledProxies []Proxy
    for _, p := range pm.proxies {
        if !p.Disabled {
            enabledProxies = append(enabledProxies, p)
        }
    }
    
    if len(enabledProxies) == 0 {
        return "" // No proxies available
    }
    
    // Choose proxy based on success rate and weight
    var selectedProxy Proxy
    var maxScore float64 = -1
    
    for _, p := range enabledProxies {
        successRate := pm.successRates[p.URL]
        if pm.totalRequests[p.URL] < 10 {
            // For new proxies, give them priority
            successRate = 0.9
        }
        
        // Calculate score based on success rate and weight
        score := successRate * float64(p.Weight)
        
        if score > maxScore {
            maxScore = score
            selectedProxy = p
        }
    }
    
    return selectedProxy.URL
}

// ReportSuccess reports a successful request for a proxy
func (pm *ProxyManager) ReportSuccess(proxyURL string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.totalRequests[proxyURL]++
    
    // Update success rate with exponential moving average
    currentRate := pm.successRates[proxyURL]
    newRate := currentRate*0.9 + 0.1 // 0.1 = success
    pm.successRates[proxyURL] = newRate
}

// ReportFailure reports a failed request for a proxy
func (pm *ProxyManager) ReportFailure(proxyURL string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.totalRequests[proxyURL]++
    
    // Update success rate with exponential moving average
    currentRate := pm.successRates[proxyURL]
    newRate := currentRate*0.9 + 0.0 // 0.0 = failure
    pm.successRates[proxyURL] = newRate
    
    // Disable proxy if success rate is too low
    if pm.totalRequests[proxyURL] > 10 && newRate < 0.3 {
        for i, p := range pm.proxies {
            if p.URL == proxyURL {
                pm.proxies[i].Disabled = true
                break
            }
        }
    }
}
```

## Machine Learning Enhancements

For users looking to extract even more value from the scraped data, consider adding machine learning components:

```mermaid
flowchart TB
    subgraph "ML Pipeline"
        direction TB
        raw[Raw Scraped Data] --> preprocess[Data Preprocessing]
        preprocess --> features[Feature Extraction]
        features --> models[ML Models]
        
        subgraph "ML Models"
            sentiment[Sentiment Analysis<br>for Reviews]
            categorize[Business<br>Categorization]
            cluster[Geographic<br>Clustering]
            forecast[Trend<br>Forecasting]
        end
        
        models --> sentiment
        models --> categorize
        models --> cluster
        models --> forecast
        
        subgraph "Applications"
            direction LR
            market[Market Analysis]
            targeting[Target Marketing]
            investment[Investment Decisions]
            expansion[Expansion Planning]
        end
        
        sentiment --> market
        sentiment --> targeting
        categorize --> targeting
        categorize --> market
        cluster --> investment
        cluster --> expansion
        forecast --> investment
        forecast --> expansion
    end
    
    classDef data fill:#f9f,stroke:#333,stroke-width:2px
    classDef model fill:#bbf,stroke:#333,stroke-width:2px
    classDef app fill:#bfb,stroke:#333,stroke-width:2px
    class raw,preprocess,features data
    class sentiment,categorize,cluster,forecast model
    class market,targeting,investment,expansion app
```

### 1. Review Sentiment Analysis

Analyze review sentiment to gauge business reputation:

```python
# Example using Python with spaCy and TextBlob
import spacy
import pandas as pd
from textblob import TextBlob

# Load scraped reviews
reviews_df = pd.read_csv("scraped_reviews.csv")

# Initialize spaCy
nlp = spacy.load("en_core_web_sm")

# Define sentiment analysis function
def analyze_sentiment(text):
    if not text or pd.isna(text):
        return 0, 0  # Neutral sentiment, no subjectivity
    
    # Basic cleaning
    text = text.lower()
    
    # Process with spaCy for entity recognition
    doc = nlp(text)
    
    # Extract entities
    entities = [ent.text for ent in doc.ents]
    
    # Calculate sentiment with TextBlob
    blob = TextBlob(text)
    sentiment = blob.sentiment.polarity
    subjectivity = blob.sentiment.subjectivity
    
    return sentiment, subjectivity, entities

# Apply sentiment analysis
sentiment_results = []
for _, row in reviews_df.iterrows():
    sentiment, subjectivity, entities = analyze_sentiment(row['review_text'])
    
    sentiment_results.append({
        'business_id': row['business_id'],
        'review_id': row['review_id'],
        'sentiment': sentiment,
        'subjectivity': subjectivity,
        'entities': entities,
        'original_rating': row['rating']
    })

# Convert to DataFrame and save
sentiment_df = pd.DataFrame(sentiment_results)
sentiment_df.to_csv("review_sentiment_analysis.csv", index=False)

# Aggregate sentiment by business
business_sentiment = sentiment_df.groupby('business_id').agg({
    'sentiment': ['mean', 'std', 'min', 'max'],
    'subjectivity': ['mean', 'std'],
    'original_rating': ['mean', 'count']
})

business_sentiment.to_csv("business_sentiment_summary.csv")
```

### 2. Business Categorization

Use machine learning to categorize businesses more precisely:

```python
# Example using scikit-learn for business categorization
import pandas as pd
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report

# Load scraped business data
businesses_df = pd.read_csv("scraped_businesses.csv")

# Prepare data
businesses_df['text_features'] = businesses_df['title'] + ' ' + businesses_df['category'] + ' ' + businesses_df.get('descriptions', '')

# Vectorize text features
vectorizer = TfidfVectorizer(max_features=5000, stop_words='english')
X = vectorizer.fit_transform(businesses_df['text_features'].fillna(''))

# Use existing category as target (for supervised learning)
y = businesses_df['category']

# Train-test split
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size## Technical Deep Dive

This section provides deeper technical insights into the Google Maps Scraper's architecture and implementation.

### Architecture Overview

```mermaid
graph TB
    subgraph "Core Components"
        direction TB
        cmd[Command Line Interface] --> engine
        web[Web Interface] --> engine
        engine[Scraper Engine] --> crawler
        engine --> parser
        crawler[Crawler] --> browser
        browser[Headless Browser] --> maps[Google Maps]
        parser[Parser] --> extraction
        extraction[Data Extraction] --> results[Results Processor]
        results --> exporters
        
        subgraph "Exporters"
            direction LR
            csv[CSV]
            json[JSON]
            db[PostgreSQL]
            plugins[Custom Plugins]
        end
        
        exporters --> csv
        exporters --> json
        exporters --> db
        exporters --> plugins
    end
    
    subgraph "Supporting Systems"
        direction TB
        queue[Job Queue] --> scheduler[Job Scheduler]
        cache[Result Cache] --> dedupe[Deduplication]
        proxy[Proxy Manager] --> rotation[Rotation Strategy]
        metrics[Metrics Collector] --> telemetry[Telemetry]
    end
    
    engine --> queue
    results --> cache
    browser --> proxy
    engine --> metrics
    
    classDef core fill:#f9f,stroke:#333,stroke-width:2px
    classDef support fill:#bbf,stroke:#333,stroke-width:2px
    classDef external fill:#bfb,stroke:#333,stroke-width:2px
    class cmd,web,engine,crawler,browser,parser,extraction,results,exporters,csv,json,db,plugins core
    class queue,scheduler,cache,dedupe,proxy,rotation,metrics,telemetry support
    class maps external
```

### Key Components

#### 1. Scraper Engine

The core of the system orchestrates the entire scraping process:

```go
// Simplified engine implementation
type Engine struct {
    Config     *Config
    JobQueue   *JobQueue
    Results    *ResultsProcessor
    Metrics    *MetricsCollector
    Proxies    *ProxyManager
    Browser    *Browser
}

func (e *Engine) Run(ctx context.Context) error {
    // Initialize components
    if err := e.init(); err != nil {
        return err
    }
    
    // Process jobs until context is cancelled or queue is empty
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            job, err := e.JobQueue.Next()
            if err != nil {
                if errors.Is(err, ErrQueueEmpty) {
                    return nil
                }
                return err
            }
            
            // Process the job
            result, err := e.processJob(ctx, job)
            if err != nil {
                e.Metrics.RecordError(job, err)
                continue
            }
            
            // Handle result
            if err := e.Results.Process(result); err != nil {
                e.Metrics.RecordError(job, err)
            }
        }
    }
}
```

#### 2. Browser Management

The scraper uses headless Chrome/Chromium via the [chromedp](https://github.com/chromedp/chromedp) package:

```go
// Browser wrapper around chromedp
type Browser struct {
    ctx           context.Context
    cancel        context.CancelFunc
    options       []chromedp.ExecAllocatorOption
    browserCtx    context.Context
    browserCancel context.CancelFunc
}

func NewBrowser(ctx context.Context, headless bool, proxyURL string) (*Browser, error) {
    // Default options
    options := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", headless),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    
    // Add proxy if specified
    if proxyURL != "" {
        options = append(options, chromedp.ProxyServer(proxyURL))
    }
    
    // Create allocator context
    allocCtx, cancel := chromedp.NewExecAllocator(ctx, options...)
    
    // Create browser context
    browserCtx, browserCancel := chromedp.NewContext(allocCtx)
    
    return &Browser{
        ctx:           allocCtx,
        cancel:        cancel,
        options:       options,
        browserCtx:    browserCtx,
        browserCancel: browserCancel,
    }, nil
}
```

#### 3. Data Extraction

The scraper extracts data using a combination of DOM traversal and JavaScript execution:

```go
// Extract business data from Google Maps
func extractBusinessData(ctx context.Context, browser *Browser, url string) (*Business, error) {
    var business Business
    
    // Navigate to the business page
    if err := chromedp.Run(browser.browserCtx, chromedp.Navigate(url)); err != nil {
        return nil, fmt.Errorf("navigation error: %w", err)
    }
    
    // Wait for the page to load
    if err := chromedp.Run(browser.browserCtx, chromedp.WaitVisible("div[role=main]", chromedp.ByQuery)); err != nil {
        return nil, fmt.Errorf("wait error: %w", err)
    }
    
    // Extract data via JavaScript
    var jsonData string
    if err := chromedp.Run(browser.browserCtx, chromedp.Evaluate(`
        (() => {
            const data = {};
            // Title
            const titleElement = document.querySelector('h1');
            if (titleElement) data.title = titleElement.textContent;
            
            // Address
            const addressElement = document.querySelector('button[data-item-id="address"]');
            if (addressElement) data.address = addressElement.textContent;
            
            // Reviews
            const reviewElement = document.querySelector('span[aria-label*="stars"]');
            if (reviewElement) {
                const match = reviewElement.getAttribute('aria-label').match(/(\d+\.\d+) stars/);
                if (match) data.rating = parseFloat(match[1]);
            }
            
            // ...more extraction logic...
            
            return JSON.stringify(data);
        })()
    `, &jsonData)); err != nil {
        return nil, fmt.Errorf("extraction error: %w", err)
    }
    
    // Parse the JSON data
    if err := json.Unmarshal([]byte(jsonData), &business); err != nil {
        return nil, fmt.Errorf("parsing error: %w", err)
    }
    
    return &business, nil
}
```

#### 4. Email Extraction

When email extraction is enabled, the scraper visits the business website and uses regex patterns:

```go
// Extract emails from a website
func extractEmails(ctx context.Context, browser *Browser, websiteURL string) ([]string, error) {
    // Navigate to the website
    if err := chromedp.Run(browser.browserCtx, chromedp.Navigate(websiteURL)); err != nil {
        return nil, fmt.Errorf("navigation error: %w", err)
    }
    
    // Wait for the page to load
    if err := chromedp.Run(browser.browserCtx, chromedp.WaitVisible("body", chromedp.ByQuery)); err != nil {
        return nil, fmt.Errorf("wait error: %w", err)
    }
    
    // Get page HTML
    var html string
    if err := chromedp.Run(browser.browserCtx, chromedp.OuterHTML("html", &html)); err != nil {
        return nil, fmt.Errorf("html extraction error: %w", err)
    }
    
    // Use regex to find emails
    var emails []string
    emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
    matches := emailRegex.FindAllString(html, -1)
    
    // Deduplicate emails
    emailMap := make(map[string]struct{})
    for _, email := range matches {
        emailMap[strings.ToLower(email)] = struct{}{}
    }
    
    for email := range emailMap {
        emails = append(emails, email)
    }
    
    return emails, nil
}
```

### Performance Optimization

The scraper employs several performance optimization techniques:

1. **Concurrent Processing**
   - Goroutines for parallel job processing
   - Worker pool pattern for resource management
   - Context-based cancellation for cleanup

2. **Memory Management**
   - Browser recycling to minimize resource usage
   - Incremental result processing to reduce memory footprint
   - Garbage collection hints for long-running operations

3. **Network Optimization**
   - Request batching where possible
   - Connection pooling
   - Proxy rotation strategies

Here's an example of the worker pool implementation:

```go
// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
    done    chan struct{}
    wg      sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan Job),
        results: make(chan Result),
        done:    make(chan struct{}),
    }
}

// Start starts the worker pool
func (p *WorkerPool) Start(ctx context.Context, processor JobProcessor) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go func(workerID int) {
            defer p.wg.Done()
            
            // Create a dedicated browser for this worker
            browser, err := NewBrowser(ctx, true, "")
            if err != nil {
                log.Printf("Worker %d: Failed to create browser: %v", workerID, err)
                return
            }
            defer browser.Close()
            
            for {
                select {
                case <-ctx.Done():
                    return
                case <-p.done:
                    return
                case job, ok := <-p.jobs:
                    if !ok {
                        return
                    }
                    
                    // Process the job
                    result, err := processor.Process(ctx, browser, job)
                    if err != nil {
                        log.Printf("Worker %d: Failed to process job %v: %v", workerID, job.ID, err)
                        continue
                    }
                    
                    // Send the result
                    select {
                    case <-ctx.Done():
                        return
                    case p.results <- result:
                        // Result sent successfully
                    }
                }
            }
        }(i)
    }
}
```

### Custom Plugin System

The Google Maps Scraper supports custom Go plugins for extending its functionality. Here's an example of the plugin interface:

```go
// WriterPlugin defines the interface for custom writer plugins
type WriterPlugin interface {
    // Name returns the name of the plugin
    Name() string
    
    // Open initializes the plugin with the provided configuration
    Open(config map[string]interface{}) error
    
    // Write writes a business record to the destination
    Write(business *Business) error
    
    // Close closes the plugin
    Close() error
}
```

Example plugin implementation:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

// DummyPrinter is a simple plugin that prints businesses to files
type DummyPrinter struct {
    outputDir string
}

// Name returns the name of the plugin
func (d *DummyPrinter) Name() string {
    return "DummyPrinter"
}

// Open initializes the plugin
func (d *DummyPrinter) Open(config map[string]interface{}) error {
    // Get output directory from config, or use a default
    if dir, ok := config["outputDir"].(string); ok {
        d.outputDir = dir
    } else {
        d.outputDir = "output"
    }
    
    // Create the output directory if it doesn't exist
    if err := os.MkdirAll(d.outputDir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }
    
    return nil
}

// Write writes a business record to a file
func (d *DummyPrinter) Write(business *Business) error {
    if business == nil {
        return fmt.Errorf("nil business")
    }
    
    // Create a filename using the business ID or title
    filename := business.CID
    if filename == "" {
        // Sanitize the title for use in a filename
        filename = sanitizeFilename(business.Title)
    }
    
    // Add timestamp to ensure uniqueness
    timestamp := time.Now().Format("20060102-150405")
    filename = fmt.Sprintf("%s-%s.json", filename, timestamp)
    
    // Create the full path
    path := filepath.Join(d.outputDir, filename)
    
    // Marshal the business to JSON
    data, err := json.MarshalIndent(business, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal business: %w", err)
    }
    
    // Write to the file
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    fmt.Printf("Wrote business %q to %s\n", business.Title, path)
    return nil
}

// Close closes the plugin
func (d *DummyPrinter) Close() error {
    // No resources to clean up
    return nil
}

// sanitizeFilename sanitizes a string for use in a filename
func sanitizeFilename(s string) string {
    // Replace invalid characters with underscores
    invalidChars := []rune{'/', '\\', '?', '%', '*', ':', '|', '"', '<', '>', '.', ' '}
    result := s
    for _, c := range invalidChars {
        result = strings.ReplaceAll(result, string(c), "_")
    }
    return result
}

// DummyPrinter is the exported plugin symbol
var DummyPrinter DummyPrinter
```

## Database Schema

When using the PostgreSQL storage option, the scraper uses the following schema:

```mermaid
erDiagram
    GMAPS_JOBS ||--o{ GMAPS_RESULTS : "produces"
    GMAPS_JOBS {
        uuid id PK
        string query
        string status
        timestamp created_at
        timestamp updated_at
        string language
        int depth
        boolean email_extraction
        float radius
        string geo
        int zoom
        boolean fast_mode
        json metadata
    }
    
    GMAPS_RESULTS {
        uuid id PK
        uuid job_id FK
        string title
        string category
        string address
        string website
        string phone
        float latitude
        float longitude
        float review_rating
        int review_count
        string cid
        timestamp created_at
        json raw_data
    }
    
    GMAPS_EMAILS {
        uuid id PK
        uuid result_id FK
        string email
        timestamp created_at
    }
    
    GMAPS_REVIEWS {
        uuid id PK
        uuid result_id FK
        string author
        float rating
        string text
        timestamp review_date
        timestamp created_at
    }
```

SQL for creating these tables:

```sql
-- Jobs table
CREATE TABLE gmaps_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    language VARCHAR(10) DEFAULT 'en',
    depth INTEGER DEFAULT 10,
    email_extraction BOOLEAN DEFAULT FALSE,
    radius FLOAT DEFAULT 10000,
    geo TEXT,
    zoom INTEGER DEFAULT 15,
    fast_mode BOOLEAN DEFAULT FALSE,
    metadata JSONB
);

-- Results table
CREATE TABLE gmaps_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES gmaps_jobs(id) ON DELETE CASCADE,
    title TEXT,
    category TEXT,
    address TEXT,
    website TEXT,
    phone TEXT,
    latitude FLOAT,
    longitude FLOAT,
    review_rating FLOAT,
    review_count INTEGER,
    cid TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    raw_data JSONB,
    UNIQUE (job_id, cid)
);

-- Emails table
CREATE TABLE gmaps_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    result_id UUID REFERENCES gmaps_results(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (result_id, email)
);

-- Reviews table
CREATE TABLE gmaps_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    result_id UUID REFERENCES gmaps_results(id) ON DELETE CASCADE,
    author TEXT,
    rating FLOAT,
    text TEXT,
    review_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_gmaps_jobs_status ON gmaps_jobs(status);
CREATE INDEX idx_gmaps_jobs_created_at ON gmaps_jobs(created_at);
CREATE INDEX idx_gmaps_results_job_id ON gmaps_results(job_id);
CREATE INDEX idx_gmaps_emails_result_id ON gmaps_emails(result_id);
CREATE INDEX idx_gmaps_reviews_result_id ON gmaps_reviews(result_id);
```

## AWS Lambda Integration

For serverless execution, the Google Maps Scraper supports running on AWS Lambda:

```mermaid
sequenceDiagram
    participant User
    participant Invoker as Lambda Invoker
    participant Queue as SQS Queue
    participant Lambda
    participant S3
    
    User->>Invoker: Submit jobs
    Invoker->>Invoker: Split into chunks
    loop For each chunk
        Invoker->>Queue: Enqueue job chunk
    end
    
    Lambda->>Queue: Poll for messages
    Queue-->>Lambda: Job chunk
    Lambda->>Lambda: Process jobs
    Lambda->>S3: Store results
    
    User->>S3: Retrieve results
```

### Lambda Handler

```go
// Lambda handler function
func HandleRequest(ctx context.Context, event events.SQSEvent) error {
    // Initialize the scraper
    scraper, err := NewScraper(ctx)
    if err != nil {
        return fmt.Errorf("failed to initialize scraper: %w", err)
    }
    defer scraper.Close()
    
    // Process each SQS message
    for _, record := range event.Records {
        // Parse the job from the message
        var job Job
        if err := json.Unmarshal([]byte(record.Body), &job); err != nil {
            log.Printf("Failed to parse job: %v", err)
            continue
        }
        
        log.Printf("Processing job: %s", job.Query)
        
        // Process the job
        result, err := scraper.ProcessJob(ctx, job)
        if err != nil {
            log.Printf("Failed to process job: %v", err)
            continue
        }
        
        // Store the result in S3
        if err := storeResultInS3(ctx, result); err != nil {
            log.Printf("Failed to store result: %v", err)
            continue
        }
    }
    
    return nil
}

// Store result in S3
func storeResultInS3(ctx context.Context, result *Result) error {
    // Initialize S3 client
    s3Client := s3.NewFromConfig(aws.Config{
        Region: os.Getenv("AWS_REGION"),
    })
    
    // Marshal the result to JSON
    data, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal result: %w", err)
    }
    
    // Create the S3 key
    key := fmt.Sprintf("results/%s/%s.json", result.JobID, result.ID)
    
    // Upload to S3
    _, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET")),
        Key:    aws.String(key),
        Body:   bytes.NewReader(data),
    })
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    
    log.Printf("Stored result in S3: %s", key)
    return nil
}
```# Google Maps Scraper
![build](https://github.com/Vector/vector-leads-scraper/actions/workflows/build.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/Vector/vector-leads-scraper)](https://goreportcard.com/report/github.com/Vector/vector-leads-scraper)

> A free and open-source Google Maps scraper with both command line and web UI options. This tool is easy to use and allows you to extract data from Google Maps efficiently.

```mermaid
graph TB
    subgraph "Google Maps Scraper Architecture"
        direction TB
        cli[Command Line Interface] --> engine
        web[Web UI] --> engine
        engine[Scraping Engine] --> providers
        
        subgraph "Providers"
            file[File Output]
            json[JSON Output]
            db[PostgreSQL]
            custom[Custom Plugins]
        end
        
        providers --> file
        providers --> json
        providers --> db
        providers --> custom
        
        engine --> features
        
        subgraph "Features"
            direction LR
            email[Email Extraction]
            proxy[Proxy Support]
            aws[AWS Integration]
            fast[Fast Mode]
        end
    end
```

## Sponsors

### Supported by the Community

[Supported by the community](https://github.com/sponsors/gosom)

### Premium Sponsors

**No time for code? Extract ALL Google Maps listings at country-scale in 2 clicks, without keywords or limits** 👉 [Try it now for free](https://scrap.io?utm_medium=ads&utm_source=github_gosom_gmap_scraper)

[![Extract ALL Google Maps Listings](./img/premium_scrap_io.png)](https://scrap.io?utm_medium=ads&utm_source=github_gosom_gmap_scraper)

<hr>

<table>
<tr>
<td><img src="./img/SerpApi-logo-w.png" alt="SerpApi Logo" width="100"></td>
<td>
<b>At SerpApi, we scrape public data from Google Maps and other top search engines.</b>

You can find the full list of our APIs here: [https://serpapi.com/search-api](https://serpapi.com/search-api)
</td>
</tr>
</table>

[![SerpApi Banner](./img/SerpApi-banner.png)](https://serpapi.com/?utm_source=google-maps-scraper)

<hr>


### Special Thanks to:

[![Google Maps API for easy SERP scraping](https://www.searchapi.io/press/v1/svg/searchapi_logo_black_h.svg)](https://www.searchapi.io/google-maps?via=gosom)
**Google Maps API for easy SERP scraping**

<hr>

[![Capsolver banner](https://raw.githubusercontent.com/gosom/google-maps-scraper/main/img/capsolver-banner.png)](https://www.capsolver.com/?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos)
**[CapSolver](https://www.capsolver.com/?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos)** automates CAPTCHA solving for efficient web scraping. It supports [reCAPTCHA V2](https://docs.capsolver.com/guide/captcha/ReCaptchaV2.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), [reCAPTCHA V3](https://docs.capsolver.com/guide/captcha/ReCaptchaV3.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), [hCaptcha](https://docs.capsolver.com/guide/captcha/HCaptcha.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), and more. With API and extension options, it's perfect for any web scraping project.

<hr>

[Evomi](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps) is your Swiss Quality Proxy Provider, starting at **$0.49/GB**

[![Evomi Banner](https://my.evomi.com/images/brand/cta.png)](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps)

<hr>

## What Google Maps Scraper Does

A command line and web-based Google Maps scraper built using the [scrapemate](https://github.com/gosom/scrapemate) web crawling framework. You can use this repository as is, or you can customize the code to suit your specific needs.

```mermaid
sequenceDiagram
    participant User
    participant Scraper
    participant GMaps as Google Maps
    participant Website as Business Website
    
    User->>Scraper: Input search queries
    Scraper->>GMaps: Search for businesses
    GMaps-->>Scraper: Search results page
    Scraper->>Scraper: Extract basic data
    Scraper->>GMaps: Visit business details
    GMaps-->>Scraper: Business details page
    Scraper->>Scraper: Extract comprehensive data
    alt Email extraction enabled
        Scraper->>Website: Visit business website
        Website-->>Scraper: Website content
        Scraper->>Scraper: Extract email addresses
    end
    Scraper-->>User: Structured data results
```

![Example GIF](img/example.gif)

### Web UI:

```bash
mkdir -p gmapsdata && docker run -v $PWD/gmapsdata:/gmapsdata -p 8080:8080 gosom/google-maps-scraper -data-folder /gmapsdata
```

Or download the [binary](https://github.com/Vector/vector-leads-scraper/releases) for your platform and run it.

> **Note**: The results will take at least 3 minutes to appear, even if you add only one keyword. This is the minimum configured runtime.

> **Note**: For MacOS, the docker command may not work as expected. **HELP REQUIRED**

### Command Line:

```bash
touch results.csv && docker run -v $PWD/example-queries.txt:/example-queries -v $PWD/results.csv:/results.csv gosom/google-maps-scraper -depth 1 -input /example-queries -results /results.csv -exit-on-inactivity 3m
```

The file `results.csv` will contain the parsed results.

**If you want to extract emails, use the `-email` parameter additionally**

### REST API

The Google Maps Scraper provides a RESTful API for programmatic management of scraping tasks.

```mermaid
classDiagram
    class APIController {
        +CreateJob(request)
        +GetJob(id)
        +ListJobs()
        +DeleteJob(id)
        +DownloadResults(id)
    }
    
    class Job {
        +String id
        +String status
        +String[] queries
        +int depth
        +bool emailExtraction
        +Date createdAt
        +Date completedAt
    }
    
    class Result {
        +String businessName
        +String category
        +String address
        +String phone
        +String website
        +float rating
        +int reviewCount
        +String[] emails
    }
    
    APIController --> Job : manages
    Job --> Result : contains
```

#### Key Endpoints

- `POST /api/v1/jobs`: Create a new scraping job
- `GET /api/v1/jobs`: List all jobs
- `GET /api/v1/jobs/{id}`: Get details of a specific job
- `DELETE /api/v1/jobs/{id}`: Delete a job
- `GET /api/v1/jobs/{id}/download`: Download job results as CSV

For detailed API documentation, refer to the OpenAPI 3.0.3 specification available through Swagger UI or Redoc when running the app at http://localhost:8080/api/docs

## 🌟 Support the Project!

If you find this tool useful, consider giving it a **star** on GitHub. 
Feel free to check out the **Sponsor** button on this repository to see how you can further support the development of this project. 
Your support helps ensure continued improvement and maintenance.

## Features

```mermaid
mindmap
  root((Google Maps Scraper))
    Data Extraction
      Business details
      Contact information
      Geolocation data
      Reviews and ratings
      Operating hours
      Price ranges
    Output Options
      CSV export
      JSON export
      PostgreSQL storage
      Custom plugins
    Performance
      120 URLs/minute
      Concurrent processing
      Multi-machine scaling
    Special Features
      Email extraction
      Proxy support
        SOCKS5
        HTTP/HTTPS
      AWS Lambda integration
      Fast Mode (Beta)
```

- Extracts many data points from Google Maps
- Exports the data to CSV, JSON or PostgreSQL 
- Performance about 120 URLs per minute (-depth 1 -c 8)
- Extendable to write your own exporter
- Dockerized for easy run on multiple platforms
- Scalable across multiple machines
- Optionally extracts emails from the website of the business
- SOCKS5/HTTP/HTTPS proxy support
- Serverless execution via AWS Lambda functions (experimental & no documentation yet)
- Fast Mode (BETA)

## Notes on Email Extraction

By default, email extraction is disabled. 

If you enable email extraction (see quickstart), the scraper will visit the business website (if it exists) and try to extract emails from the page.

Currently, it only checks one page of the website (the one registered in Google Maps). Support for extracting from other pages like about, contact, impressum, etc. will be added in the future.

```mermaid
flowchart LR
    scraper[Google Maps Scraper]
    gmaps[Google Maps Page]
    website[Business Website]
    result[Results with Email]
    
    scraper -->|1. Scrape| gmaps
    gmaps -->|2. Get Website URL| scraper
    scraper -->|3. Visit Website| website
    website -->|4. Extract Email| scraper
    scraper -->|5. Add to| result
    
    style scraper fill:#f9f,stroke:#333,stroke-width:2px
    style gmaps fill:#bbf,stroke:#333,stroke-width:2px
    style website fill:#bbf,stroke:#333,stroke-width:2px
    style result fill:#bfb,stroke:#333,stroke-width:2px
```

Keep in mind that enabling email extraction results in longer processing time, since more pages are scraped.

## Fast Mode

Fast mode returns at most 21 search results per query ordered by distance from the provided **latitude** and **longitude** within the specified **radius**.

It doesn't contain all data points but provides basic ones, allowing for much faster data extraction.

When using fast mode, ensure you have provided:
- zoom
- radius (in meters)
- latitude
- longitude

**Fast mode is Beta, you may experience blocking**

## Data Fields Extracted

```mermaid
classDiagram
    class BusinessListing {
        +String input_id
        +String link
        +String title
        +String category
        +String address
        +String open_hours
        +String popular_times
        +String website
        +String phone
        +String plus_code
        +int review_count
        +float review_rating
        +Map reviews_per_rating
        +float latitude
        +float longitude
        +String cid
        +String status
        +String descriptions
        +String reviews_link
        +String thumbnail
        +String timezone
        +String price_range
        +String data_id
        +String[] images
        +String reservations
        +String order_online
        +String menu
        +boolean owner
        +String complete_address
        +String about
        +Map user_reviews
        +String[] emails
    }
```

### Extracted Data Fields

#### 1. `input_id`
- Internal identifier for the input query.

#### 2. `link`
- Direct URL to the business listing on Google Maps.

#### 3. `title`
- Name of the business.

#### 4. `category`
- Business type or category (e.g., Restaurant, Hotel).

#### 5. `address`
- Street address of the business.

#### 6. `open_hours`
- Business operating hours.

#### 7. `popular_times`
- Estimated visitor traffic at different times of the day.

#### 8. `website`
- Official business website.

#### 9. `phone`
- Business contact phone number.

#### 10. `plus_code`
- Shortcode representing the precise location of the business.

#### 11. `review_count`
- Total number of customer reviews.

#### 12. `review_rating`
- Average star rating based on reviews.

#### 13. `reviews_per_rating`
- Breakdown of reviews by each star rating (e.g., number of 5-star, 4-star reviews).

#### 14. `latitude`
- Latitude coordinate of the business location.

#### 15. `longitude`
- Longitude coordinate of the business location.

#### 16. `cid`
- **Customer ID** (CID) used by Google Maps to uniquely identify a business listing. This ID remains stable across updates and can be used in URLs.
- **Example:** `3D3174616216150310598`

#### 17. `status`
- Business status (e.g., open, closed, temporarily closed).

#### 18. `descriptions`
- Brief description of the business.

#### 19. `reviews_link`
- Direct link to the reviews section of the business listing.

#### 20. `thumbnail`
- URL to a thumbnail image of the business.

#### 21. `timezone`
- Time zone of the business location.

#### 22. `price_range`
- Price range of the business (`$`, `$$`, `$$$`).

#### 23. `data_id`
- An internal Google Maps identifier composed of two hexadecimal values separated by a colon.
- **Structure:** `<spatial_hex>:<listing_hex>`
- **Example:** `0x3eb33fecd7dfa167:0x2c0e80a0f5d57ec6`
- **Note:** This value may change if the listing is updated and should not be used for permanent identification.

#### 24. `images`
- Links to images associated with the business.

#### 25. `reservations`
- Link to book reservations (if available).

#### 26. `order_online`
- Link to place online orders.

#### 27. `menu`
- Link to the menu (for applicable businesses).

#### 28. `owner`
- Indicates whether the business listing is claimed by the owner.

#### 29. `complete_address`
- Fully formatted address of the business.

#### 30. `about`
- Additional information about the business.

#### 31. `user_reviews`
- Collection of customer reviews, including text, rating, and timestamp.

#### 32. `emails`
- Email addresses associated with the business, if available.

**Note**: Email is empty by default (see Usage)

**Note**: Input ID is an ID that you can define per query. By default, it's a UUID.
To define it, you can have an input file like:

```
Matsuhisa Athens #!#MyIDentifier
```

## Quickstart

### Using Docker:

```bash
touch results.csv && docker run -v $PWD/example-queries.txt:/example-queries -v $PWD/results.csv:/results.csv gosom/google-maps-scraper -depth 1 -input /example-queries -results /results.csv -exit-on-inactivity 3m
```

The file `results.csv` will contain the parsed results.

**If you want emails, use the `-email` parameter additionally**

### On Your Host

(tested only on Ubuntu 22.04)

```bash
git clone https://github.com/Vector/vector-leads-scraper.git
cd google-maps-scraper
go mod download
go build
./google-maps-scraper -input example-queries.txt -results restaurants-in-cyprus.csv -exit-on-inactivity 3m
```

Be a little bit patient. The first run downloads required libraries.

The results are written when they arrive in the `results` file you specified.

**If you want emails, use the `-email` parameter additionally**

### Command Line Options

Try `./google-maps-scraper -h` to see the command line options available:

```
  -addr string
        address to listen on for web server (default ":8080")
  -aws-access-key string
        AWS access key
  -aws-lambda
        run as AWS Lambda function
  -aws-lambda-chunk-size int
        AWS Lambda chunk size (default 100)
  -aws-lambda-invoker
        run as AWS Lambda invoker
  -aws-region string
        AWS region
  -aws-secret-key string
        AWS secret key
  -c int
        sets the concurrency [default: half of CPU cores] (default 11)
  -cache string
        sets the cache directory [no effect at the moment] (default "cache")
  -data-folder string
        data folder for web runner (default "webdata")
  -debug
        enable headful crawl (opens browser window) [default: false]
  -depth int
        maximum scroll depth in search results [default: 10] (default 10)
  -dsn string
        database connection string [only valid with database provider]
  -email
        extract emails from websites
  -exit-on-inactivity duration
        exit after inactivity duration (e.g., '5m')
  -fast-mode
        fast mode (reduced data collection)
  -function-name string
        AWS Lambda function name
  -geo string
        set geo coordinates for search (e.g., '37.7749,-122.4194')
  -input string
        path to the input file with queries (one per line) [default: empty]
  -json
        produce JSON output instead of CSV
  -lang string
        language code for Google (e.g., 'de' for German) [default: en] (default "en")
  -produce
        produce seed jobs only (requires dsn)
  -proxies string
        comma separated list of proxies to use in the format protocol://user:pass@host:port example: socks5://localhost:9050 or http://user:pass@localhost:9050
  -radius float
        search radius in meters. Default is 10000 meters (default 10000)
  -results string
        path to the results file [default: stdout] (default "stdout")
  -s3-bucket string
        S3 bucket name
  -web
        run web server instead of crawling
  -writer string
        use custom writer plugin (format: 'dir:pluginName')
  -zoom int
        set zoom level (0-21) for search (default 15)
```

## Using a Custom Writer

For cases where results need to be written in a custom format or to another system like a database or message queue, you can utilize the Go plugin system.

```mermaid
flowchart LR
    scraper[Scraper Core]
    plugin[Custom Plugin]
    output[(Custom Output)]
    
    scraper -->|Data| plugin
    plugin -->|Write| output
    
    subgraph "Plugin Implementation"
    interface[Plugin Interface]
    writer[Custom Writer]
    interface --> writer
    end
    
    style scraper fill:#f9f,stroke:#333,stroke-width:2px
    style plugin fill:#bbf,stroke:#333,stroke-width:2px
    style output fill:#bfb,stroke:#333,stroke-width:2px
```

Write a Go plugin (see an example in examples/plugins/example_writer.go) 

Compile it using (for Linux):

```bash
go build -buildmode=plugin -tags=plugin -o ~/mytest/plugins/example_writer.so examples/plugins/example_writer.go
```

and then run the program using the `-writer` argument. 

See an example:

1. Write your plugin (use the examples/plugins/example_writer.go as a reference)
2. Build your plugin `go build -buildmode=plugin -tags=plugin -o ~/myplugins/example_writer.so plugins/example_writer.go`
3. Download the latest [release](https://github.com/Vector/vector-leads-scraper/releases/) or build the program
4. Run the program like `./google-maps-scraper -writer ~/myplugins:DummyPrinter -input example-queries.txt`

### Plugins and Docker

It is possible to use the Docker image with plugins.
Make sure that the shared library is built using a compatible GLIB version with the Docker image;
otherwise, you will encounter an error like:

```
/lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found (required by /plugins/example_writer.so)
```

## Using Database Provider (PostgreSQL)

For running on your local machine:

```bash
docker-compose -f docker-compose.dev.yaml up -d
```

The above starts a PostgreSQL container and creates the required tables.

To access the database:

```bash
psql -h localhost -U postgres -d postgres
```

Password is `postgres`

Then from your host, run:

```bash
go run main.go -dsn "postgres://postgres:postgres@localhost:5432/postgres" -produce -input example-queries.txt --lang el
```

(configure your queries and the desired language)

This will populate the table `gmaps_jobs`. 

You may run the scraper using:

```bash
go run main.go -c 2 -depth 1 -dsn "postgres://postgres:postgres@localhost:5432/postgres"
```

If you have a database server and several machines, you can start multiple instances of the scraper as above.

### Kubernetes

You may run the scraper in a Kubernetes cluster. This helps to scale it more easily.

Assuming you have a Kubernetes cluster and a database that is accessible from the cluster:

1. First, populate the database as shown above
2. Create a deployment file `scraper.deployment`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: google-maps-scraper
spec:
  selector:
    matchLabels:
      app: google-maps-scraper
  replicas: {NUM_OF_REPLICAS}
  template:
    metadata:
      labels:
        app: google-maps-scraper
    spec:
      containers:
      - name: google-maps-scraper
        image: gosom/google-maps-scraper:v0.9.3
        imagePullPolicy: IfNotPresent
        args: ["-c", "1", "-depth", "10", "-dsn", "postgres://{DBUSER}:{DBPASSWD@DBHOST}:{DBPORT}/{DBNAME}", "-lang", "{LANGUAGE_CODE}"]
```

Please replace the values or the command args accordingly.

> **Note**: Keep in mind that because the application starts a headless browser, it requires CPU and memory. Use an appropriate Kubernetes cluster.

## Telemetry

Anonymous usage statistics are collected for debug and improvement reasons. 
You can opt out by setting the env variable `DISABLE_TELEMETRY=1`

## Deployment

You can deploy the scraper using the Helm chart in the `charts` folder.

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        ing[Ingress] --> svc[Service<br>ClusterIP]
        svc --> pod1[Pod: Leads Scraper<br>Port: 8080<br>Resources: 100m-500m CPU<br>128Mi-512Mi Memory]
        svc --> pod2[Pod: Leads Scraper<br>Port: 8080<br>Resources: 100m-500m CPU<br>128Mi-512Mi Memory]
        
        subgraph "Pod Details"
            direction TB
            hc[Health Check<br>/health]
            web[Web Server Mode<br>-web flag]
            scraper[Scraper Engine<br>Concurrency: 11<br>Depth: 5]
        end
        
        subgraph "AWS Integration"
            s3[S3 Bucket<br>leads-scraper-service]
        end
    end
    
    pod1 --> neon[("Neon Postgres DB<br>us-east-2.aws")]
    pod2 --> neon
    pod1 --> s3
    pod2 --> s3
```

```mermaid
flowchart TB
    subgraph "Application Layer"
        direction TB
        client[Client Requests] --> ing[Ingress]
        ing --> svc[K8s Service]
        svc --> pods[Leads Scraper Pods]
        
        subgraph "Scraper Components"
            direction LR
            web[Web Server] --> engine[Scraper Engine]
            engine --> cache[In-Memory Cache]
            engine --> queue[Job Queue]
        end
    end
    
    subgraph "Data Storage"
        direction TB
        neon[("Neon Postgres DB<br>Serverless PostgreSQL")] 
        s3[("AWS S3<br>leads-scraper-service")]
    end
    
    subgraph "External Services"
        gmaps[Google Maps API]
    end
    
    pods --> neon
    pods --> s3
    engine --> gmaps
    
    classDef db fill:#f9f,stroke:#333,stroke-width:2px
    classDef cloud fill:#bbf,stroke:#333,stroke-width:2px
    class neon db
    class s3,gmaps cloud
```

## Job Generation

```mermaid
graph TB
    subgraph "Input Processing"
        tags[Search Tags/Keywords<br>1000 keywords] --> jobs[Job Generation]
        jobs --> total[Total Jobs:<br>1000 keywords × 16 results<br>= 17,000 jobs]
    end

    subgraph "Job Breakdown per Keyword"
        direction LR
        kw[1 Keyword Search] --> initial[Initial Search Job<br>+1 job]
        initial --> results[Results Processing<br>+16 result jobs]
        results --> total_per_kw[Total per Keyword:<br>17 jobs]
    end

    subgraph "Processing Metrics"
        speed[Processing Speed<br>120 jobs/minute<br>with c=8, depth=1]
        total --> time[Processing Time<br>~17,000/120 = ~142 minutes<br>≈ 2.4 hours]
        speed --> time
    end

    classDef metrics fill:#f9f,stroke:#333,stroke-width:2px
    classDef process fill:#bbf,stroke:#333,stroke-width:2px
    class speed,time metrics
    class jobs,results process
```

## Job Processing Architecture

```mermaid
flowchart TB
    subgraph "Job Generation Layer"
        input[Input Keywords] --> parser[Tag Parser]
        parser --> jobgen[Job Generator]
        jobgen --> queue[Job Queue]
        
        subgraph "Job Types"
            direction LR
            search[Search Jobs<br>1 per keyword] 
            detail[Detail Jobs<br>~16 per keyword]
        end
    end

    subgraph "Processing Layer"
        queue --> distributor[Job Distributor]
        distributor --> worker1[Worker 1<br>Concurrency: 11]
        distributor --> worker2[Worker 2<br>Concurrency: 11]
        distributor --> worker3[Worker 3<br>Concurrency: 11]
        
        subgraph "Processing Stats"
            stats[Performance Metrics]
            stats --> rate[120 jobs/minute]
            stats --> depth[Depth: 1-10]
            stats --> conc[Concurrency: 8]
        end
    end

    subgraph "Storage Layer"
        worker1 --> db[("Neon Postgres DB")]
        worker2 --> db
        worker3 --> db
        worker1 --> s3[("AWS S3")]
        worker2 --> s3
        worker3 --> s3
    end

    classDef storage fill:#f9f,stroke:#333,stroke-width:2px
    classDef metrics fill:#bbf,stroke:#333,stroke-width:2px
    class db,s3 storage
    class stats,rate,depth,conc metrics
```

## Data Flow

```mermaid
flowchart TB
    subgraph "Data Extraction Flow"
        direction TB
        gmaps[Google Maps Page] --> parser[Parser Engine]
        
        subgraph "Extracted Data Points"
            direction LR
            basic[Basic Info<br>- Title<br>- Category<br>- Address<br>- Phone<br>- Website] 
            geo[Geo Data<br>- Latitude<br>- Longitude<br>- Plus Code]
            reviews[Review Data<br>- Review Count<br>- Rating<br>- Reviews per Rating]
            meta[Business Meta<br>- Open Hours<br>- Popular Times<br>- Price Range<br>- Timezone]
            media[Media Data<br>- Thumbnail<br>- Images<br>- Menu]
            extra[Extra Info<br>- Descriptions<br>- About<br>- Owner Info<br>- Reservations]
        end

        parser --> |Extracts| basic
        parser --> |Extracts| geo
        parser --> |Extracts| reviews
        parser --> |Extracts| meta
        parser --> |Extracts| media
        parser --> |Extracts| extra

        subgraph "Optional Extraction"
            website[Business Website] --> email[Email Extraction<br>-email flag required]
        end
    end

    classDef data fill:#f9f,stroke:#333,stroke-width:2px
    classDef source fill:#bbf,stroke:#333,stroke-width:2px
    class basic,geo,reviews,meta,media,extra data
    class gmaps,website source
```

## Deployment Modes

```mermaid
flowchart TB
    subgraph "Deployment Modes"
        direction TB
        cli[CLI Mode] --> single[Single Machine<br>Local Execution]
        web[Web Server Mode<br>-web flag] --> k8s[Kubernetes Deployment]
        
        subgraph "Configuration Options"
            direction LR
            basic_conf[Basic Config<br>- Concurrency<br>- Depth<br>- Language] 
            geo_conf[Geo Config<br>- Radius<br>- Zoom Level<br>- Coordinates]
            output[Output Options<br>- CSV<br>- JSON<br>- PostgreSQL<br>- Custom Plugin]
            perf[Performance<br>- Fast Mode<br>- Exit on Inactivity<br>- Debug Mode]
        end

        subgraph "Storage Options"
            direction LR
            files[File Storage<br>- CSV/JSON Files] 
            db[Database<br>- PostgreSQL<br>- Neon Serverless]
            cloud[Cloud Storage<br>- AWS S3]
        end
    end

    subgraph "Integration Points"
        direction TB
        proxy[Proxy Support<br>- SOCKS5<br>- HTTP/HTTPS] --> scraper
        aws[AWS Integration<br>- Lambda Functions<br>- S3 Storage] --> scraper
        scraper[Scraper Engine] --> output_handlers[Output Handlers]
    end

    classDef config fill:#f9f,stroke:#333,stroke-width:2px
    classDef mode fill:#bbf,stroke:#333,stroke-width:2px
    classDef storage fill:#e6e6fa,stroke:#333,stroke-width:2px
    class basic_conf,geo_conf,output,perf config
    class cli,web mode
    class files,db,cloud storage
```

## Performance

Expected speed with concurrency of 8 and depth 1 is 120 jobs/per minute.
Each search is 1 job + the number of results it contains.

Based on the above: 
if we have 1000 keywords to search with each containing 16 results => 1000 * 16 = 16000 jobs.

We expect this to take about 16000/120 ~ 133 minutes ~ 2.5 hours

If you want to scrape many keywords, it's better to use the Database Provider in
combination with Kubernetes for convenience and start multiple scrapers across more than one machine.

```mermaid
graph LR
    subgraph "Performance Benchmarks"
        direction TB
        conf1[Depth: 1<br>Concurrency: 8] --> rate1[120 jobs/minute]
        conf2[Depth: 5<br>Concurrency: 8] --> rate2[~80 jobs/minute]
        conf3[Depth: 10<br>Concurrency: 4] --> rate3[~50 jobs/minute]
    end
    
    subgraph "Scaling Options"
        direction TB
        single[Single Machine] --> limit[Bottlenecked by<br>Browser Instances]
        kube[Kubernetes Cluster] --> distributed[Distributed Processing]
        distributed --> linear[Near-Linear Scaling]
    end
    
    subgraph "Optimization Tips"
        direction TB
        proxy[Rotate Proxies] --> block[Avoid Blocking]
        fast[Use Fast Mode] --> quick[Quick Results]
        batch[Batch Processing] --> efficient[Resource Efficiency]
    end
    
    classDef perf fill:#f9f,stroke:#333,stroke-width:2px
    classDef scale fill:#bbf,stroke:#333,stroke-width:2px
    classDef tips fill:#bfb,stroke:#333,stroke-width:2px
    class conf1,conf2,conf3,rate1,rate2,rate3 perf
    class single,kube,distributed,linear scale
    class proxy,fast,batch,block,quick,efficient tips
```

### Performance Optimization Strategies

1. **Concurrency Tuning**:
   - Start with concurrency (`-c`) set to half your CPU cores
   - Monitor system load and adjust accordingly
   - Too high concurrency can lead to browser failures

2. **Depth Management**:
   - Lower depth values (1-3) provide faster results but fewer listings
   - Higher depth values (5-10) provide more comprehensive results but take longer
   - Balance depth with your specific needs

3. **Proxy Rotation**:
   - Use multiple proxies to avoid rate limiting
   - Consider geographic distribution of proxies for region-specific searches
   - Residential proxies typically perform better than datacenter IPs

4. **Memory Optimization**:
   - Each browser instance requires ~250-300MB RAM
   - Calculate total memory requirements: Concurrency × 300MB + 500MB base
   - Monitor for memory leaks during long-running operations

5. **Batching Strategies**:
   - Break large query sets into manageable batches
   - Process high-priority geographic areas first
   - Consider time-of-day optimization for lower Google Maps traffic

## System Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| CPU       | 2 cores | 4+ cores    | Each concurrent browser uses significant CPU |
| Memory    | 2GB     | 4GB+        | ~300MB per concurrent browser instance |
| Disk      | 1GB     | 5GB+        | For storing results and browser cache |
| Network   | 5Mbps   | 20Mbps+     | Higher bandwidth needed for concurrent requests |
| Docker    | 19.03+  | 20.10+      | For containerized deployment |
| Go        | 1.18+   | 1.21+       | For building from source |

### Kubernetes Resource Recommendations

| Resource | Request | Limit | Notes |
|----------|---------|-------|-------|
| CPU      | 100m    | 500m  | Per pod |
| Memory   | 256Mi   | 512Mi | Per pod |
| Replicas | 2       | 10+   | Scale based on workload |