# Rick & Morty Backend Service

A Go-based backend that fetches data from the official Rick & Morty API. This service allows for concurrent searches throughout data categories and also has ability to show which character pairings most frequently appear in the series.

---

## Project & Learning Process

To be honest **This is my first project using the Go programming language.**

Go was not originally listed on my CV because I only began learning it about week and a half ago, specifically for this recruitment assignment. I spent the last week and a half studying the documentation and learning how to handle concurrency, which i wish I understood better.

While I am still trying to get used to the language, I have somewhat successfully implemented:
* **Concurrent API Fetching:** I've used `goroutines` and `sync.WaitGroups` to  get the data from multiple endpoints at the same time.
* **Unit Testing:** Implemented a very simple unit test to check if the pairs from the `top-pairs` functionality are counted correctl.
* **Containerization:** I've created a DockerFile 
* **Algorithm Logic:** Creating a custom logic to map and count unique character pairs across the entire episode database.

---

### Running with Docker
**Build the image:**
   ```bash
   docker build -t rick-morty-app .
   ```
Launch the container:

```bash
docker run -p 8080:8080 rick-morty-app
The server will be live at http://localhost:8080.
```
Running Locally
```bash
go mod download
go run .
```

To run the tests:

```bash
go test -v .
```

1. Search
GET /search?term={name}&limit={n}
Searches across Characters, Locations, and Episodes

2. Top Character Pairs
GET /top-pairs?min={n}&max={n}&limit={n}
Analyzes the entire episode database to find which characters share the most occurences together.

min/max: Filter pairs based on the number of shared episodes.

limit: Restrict the number of returned pairs (default: 20).