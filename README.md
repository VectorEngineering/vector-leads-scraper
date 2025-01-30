# Google maps scraper
![build](https://github.com/Vector/vector-leads-scraper/actions/workflows/build.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/Vector/vector-leads-scraper)](https://goreportcard.com/report/github.com/Vector/vector-leads-scraper)

> A free and open-source Google Maps scraper with both command line and web UI options. This tool is easy to use and allows you to extract data from Google Maps efficiently.

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
**[CapSolver](https://www.capsolver.com/?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos)** automates CAPTCHA solving for efficient web scraping. It supports [reCAPTCHA V2](https://docs.capsolver.com/guide/captcha/ReCaptchaV2.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), [reCAPTCHA V3](https://docs.capsolver.com/guide/captcha/ReCaptchaV3.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), [hCaptcha](https://docs.capsolver.com/guide/captcha/HCaptcha.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos), and more. With API and extension options, it’s perfect for any web scraping project.

<hr>

[Evomi](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps) is your Swiss Quality Proxy Provider, starting at **$0.49/GB**

[![Evomi Banner](https://my.evomi.com/images/brand/cta.png)](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps)

<hr>


## What Google maps scraper does

A command line and web based google maps scraper build using 

[scrapemate](https://github.com/gosom/scrapemate) web crawling framework.

You can use this repository either as is, or you can use its code as a base and
customize it to your needs

![Example GIF](img/example.gif)

### Web UI:

```
mkdir -p gmapsdata && docker run -v $PWD/gmapsdata:/gmapsdata -p 8080:8080 gosom/google-maps-scraper -data-folder /gmapsdata
```

Or dowload the [binary](https://github.com/Vector/vector-leads-scraper/releases) for your platform and run it.

Note: The results will take at least 3 minutes to appear, even if you add only one keyword. This is the minimum configured runtime.

Note: for MacOS the docker command should not work. **HELP REQUIRED**


### Command line:

```
touch results.csv && docker run -v $PWD/example-queries.txt:/example-queries -v $PWD/results.csv:/results.csv gosom/google-maps-scraper -depth 1 -input /example-queries -results /results.csv -exit-on-inactivity 3m
```

file `results.csv` will contain the parsed results.

**If you want emails use additionally the `-email` parameter*

### REST API
The Google Maps Scraper provides a RESTful API for programmatic management of scraping tasks.

### Key Endpoints

- POST /api/v1/jobs: Create a new scraping job
- GET /api/v1/jobs: List all jobs
- GET /api/v1/jobs/{id}: Get details of a specific job
- DELETE /api/v1/jobs/{id}: Delete a job
- GET /api/v1/jobs/{id}/download: Download job results as CSV

For detailed API documentation, refer to the OpenAPI 3.0.3 specification available through Swagger UI or Redoc when running the app http://localhost:8080/api/docs


## 🌟 Support the Project!

If you find this tool useful, consider giving it a **star** on GitHub. 
Feel free to check out the **Sponsor** button on this repository to see how you can further support the development of this project. 
Your support helps ensure continued improvement and maintenance.


## Features

- Extracts many data points from google maps
- Exports the data to CSV, JSON or PostgreSQL 
- Performance about 120 urls per minute (-depth 1 -c 8)
- Extendable to write your own exporter
- Dockerized for easy run in multiple platforms
- Scalable in multiple machines
- Optionally extracts emails from the website of the business
- SOCKS5/HTTP/HTTPS proxy support
- Serverless execution via AWS Lambda functions (experimental & no documentation yet)
- Fast Mode (BETA)

## Notes on email extraction

By default email extraction is disabled. 

If you enable email extraction (see quickstart) then the scraper will visit the 
website of the business (if exists) and it will try to extract the emails from the
page.

For the moment it only checks only one page of the website (the one that is registered in Gmaps). At some point, it will be added support to try to extract from other pages like about, contact, impressum etc. 


Keep in mind that enabling email extraction results to larger processing time, since more
pages are scraped. 

## Fast Mode

Fast mode returns you at most 21 search results per query ordered by distance from the **latitude** and **longitude** provided.
All the results are within the specified **radius**

It does not contain all the data points but basic ones. 
However it provides the ability to extract data really fast. 

When you use the fast mode ensure that you have provided:
- zoom
- radius (in meters)
- latitude
- longitude


**Fast mode is Beta, you may experience blocking**

## Extracted Data Points

```
input_id
link
title
category
address
open_hours
popular_times
website
phone
plus_code
review_count
review_rating
reviews_per_rating
latitude
longitude
cid
status
descriptions
reviews_link
thumbnail
timezone
price_range
data_id
images
reservations
order_online
menu
owner
complete_address
about
user_reviews
emails
```

**Note**: email is empty by default (see Usage)

**Note**: Input id is an ID that you can define per query. By default it's a UUID
In order to define it you can have an input file like:

```
Matsuhisa Athens #!#MyIDentifier
```

## Quickstart

### Using docker:

```
touch results.csv && docker run -v $PWD/example-queries.txt:/example-queries -v $PWD/results.csv:/results.csv gosom/google-maps-scraper -depth 1 -input /example-queries -results /results.csv -exit-on-inactivity 3m
```

file `results.csv` will contain the parsed results.

**If you want emails use additionally the `-email` parameter**


### On your host

(tested only on Ubuntu 22.04)


```
git clone https://github.com/Vector/vector-leads-scraper.git
cd google-maps-scraper
go mod download
go build
./google-maps-scraper -input example-queries.txt -results restaurants-in-cyprus.csv -exit-on-inactivity 3m
```

Be a little bit patient. In the first run it downloads required libraries.

The results are written when they arrive in the `results` file you specified

**If you want emails use additionally the `-email` parameter**

### Command line options

try `./google-maps-scraper -h` to see the command line options available:
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

## Using a custom writer

In cases the results need to be written in a custom format or in another system like a db a message queue or basically anything the Go plugin system can be utilized.

Write a Go plugin (see an example in examples/plugins/example_writeR.go) 

Compile it using (for Linux):

```
go build -buildmode=plugin -tags=plugin -o ~/mytest/plugins/example_writer.so examples/plugins/example_writer.go
```

and then run the program using the `-writer` argument. 

See an example:

1. Write your plugin (use the examples/plugins/example_writer.go as a reference)
2. Build your plugin `go build -buildmode=plugin -tags=plugin -o ~/myplugins/example_writer.so plugins/example_writer.go`
3. Download the lastes [release](https://github.com/Vector/vector-leads-scraper/releases/) or build the program
4. Run the program like `./google-maps-scraper -writer ~/myplugins:DummyPrinter -input example-queries.txt`


### Plugins and Docker

It is possible to use the docker image and use tha plugins.
In such case make sure that the shared library is build using a compatible GLIB version with the docker image.
otherwise you will encounter an error like:

```
/lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found (required by /plugins/example_writer.so)
```


## Using Database Provider (postgreSQL)

For running in your local machine:

```
docker-compose -f docker-compose.dev.yaml up -d
```

The above starts a PostgreSQL container and creates the required tables

to access db:

```
psql -h localhost -U postgres -d postgres
```

Password is `postgres`

Then from your host run:

```
go run main.go -dsn "postgres://postgres:postgres@localhost:5432/postgres" -produce -input example-queries.txt --lang el
```

(configure your queries and the desired language)

This will populate the table `gmaps_jobs` . 

you may run the scraper using:

```
go run main.go -c 2 -depth 1 -dsn "postgres://postgres:postgres@localhost:5432/postgres"
```

If you have a database server and several machines you can start multiple instances of the scraper as above.

### Kubernetes

You may run the scraper in a kubernetes cluster. This helps to scale it easier.

Assuming you have a kubernetes cluster and a database that is accessible from the cluster:

1. First populate the database as shown above
2. Create a deployment file `scraper.deployment`

```
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

Please replace the values or the command args accordingly 

Note: Keep in mind that because the application starts a headless browser it requires CPU and memory. 
Use an appropriate kubernetes cluster

## Telemetry

Anonymous usage statistics are collected for debug and improvement reasons. 
You can opt out by setting the env variable `DISABLE_TELEMETRY=1`

## Deployment

You can deploy the scraper using the helm chart in the `charts` folder.

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

## Perfomance

Expected speed with concurrency of 8 and depth 1 is 120 jobs/per minute.
Each search is 1 job + the number or results it contains.

Based on the above: 
if we have 1000 keywords to search with each contains 16 results => 1000 * 16 = 16000 jobs.

We expect this to take about 16000/120 ~ 133 minutes ~ 2.5 hours

If you want to scrape many keywords then it's better to use the Database Provider in
combination with Kubernetes for convenience and start multiple scrapers in more than 1 machines.

## References

For more instruction you may also read the following links

- https://blog.gkomninos.com/how-to-extract-data-from-google-maps-using-golang
- https://blog.gkomninos.com/distributed-google-maps-scraping
- https://github.com/omkarcloud/google-maps-scraper/tree/master (also a nice project) [many thanks for the idea to extract the data by utilizing the JS objects]


## Licence

This code is licensed under the MIT License


## Contributing

Please open an ISSUE or make a Pull Request


Thank you for considering support for the project. Every bit of assistance helps maintain momentum and enhances the scraper’s capabilities!




## Sponsors

### Special Thanks to:

[Evomi](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps) is your Swiss Quality Proxy Provider, starting at **$0.49/GB**

- 👩‍💻 **$0.49 per GB Residential Proxies**: Our price is unbeatable
- 👩‍💻 **24/7 Expert Support**: We will join your Slack Channel
- 🌍 **Global Presence**: Available in 150+ Countries
- ⚡ **Low Latency**
- 🔒 **Swiss Quality and Privacy**
- 🎁 **Free Trial**
- 🛡️ **99.9% Uptime**
- 🤝 **Special IP Pool selection**: Optimize for fast, quality or quantity of ips
- 🔧 **Easy Integration**: Compatible with most software and programming languages

[![Evomi Banner](https://my.evomi.com/images/brand/cta.png)](https://evomi.com?utm_source=github&utm_medium=banner&utm_campaign=gosom-maps)


<div align="center">
	<p>
		<a href="https://www.capsolver.com/?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos" rel="nofollow">
                  <div>
                    <img src="https://raw.githubusercontent.com/gosom/google-maps-scraper/main/img/capsolver-banner.png" alt="Capsolver banner"/>
                  </div>
                  <b><a href="https://www.capsolver.com/?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos" rel="nofollow">CapSolver</a> automates CAPTCHA solving for efficient web scraping. It supports <a href="https://docs.capsolver.com/guide/captcha/ReCaptchaV2.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos" rel="nofolow">reCAPTCHA V2</a>, <a href="https://docs.capsolver.com/guide/captcha/ReCaptchaV3.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos" rel="nofollow">reCAPTCHA V3</a>, <a href="https://docs.capsolver.com/guide/captcha/HCaptcha.html?utm_source=github&utm_medium=banner_repo&utm_campaign=scraping&utm_term=giorgos" rel="nofollow">hCaptcha</a>, and more. With API and extension options, it’s perfect for any web scraping project. </b>
		</a>
		</a>
                <br>
                <br>
                		<a href="https://www.searchapi.io/google-maps?via=gosom" rel="nofollow">
                  <div>
                    <img src="https://www.searchapi.io/press/v1/svg/searchapi_logo_black_h.svg" width="300" alt="Google Maps API for easy SERP scraping"/>
                  </div>
                  <b>Google Maps API for easy SERP scraping</b>
		<br>
		<br>
	</p>
</div>