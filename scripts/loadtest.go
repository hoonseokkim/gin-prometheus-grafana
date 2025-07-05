package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	BaseURL = "http://localhost:8080/api/v1/books"
	DefaultRequests = 500
	DefaultConcurrency = 10
	DefaultDuration = 60 // seconds
)

type Book struct {
	ID          int     `json:"id,omitempty"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	ISBN        string  `json:"isbn"`
	Price       float64 `json:"price"`
	PublishedAt string  `json:"published_at"`
}

type UpdateBook struct {
	Title       *string  `json:"title,omitempty"`
	Author      *string  `json:"author,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	PublishedAt *string  `json:"published_at,omitempty"`
}

type Stats struct {
	Total     int
	Success   int
	Errors    int
	Creates   int
	Reads     int
	Updates   int
	Deletes   int
	mutex     sync.Mutex
}

func (s *Stats) Increment(operation string, success bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.Total++
	if success {
		s.Success++
	} else {
		s.Errors++
	}
	
	switch operation {
	case "create":
		s.Creates++
	case "read":
		s.Reads++
	case "update":
		s.Updates++
	case "delete":
		s.Deletes++
	}
}

func (s *Stats) Print() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Total Requests: %d\n", s.Total)
	fmt.Printf("Successful: %d\n", s.Success)
	fmt.Printf("Errors: %d\n", s.Errors)
	fmt.Printf("Success Rate: %.2f%%\n", float64(s.Success)/float64(s.Total)*100)
	fmt.Printf("\nOperation Breakdown:\n")
	fmt.Printf("  Creates: %d\n", s.Creates)
	fmt.Printf("  Reads: %d\n", s.Reads)
	fmt.Printf("  Updates: %d\n", s.Updates)
	fmt.Printf("  Deletes: %d\n", s.Deletes)
	fmt.Printf("========================\n")
}

var sampleBooks = []Book{
	{Title: "The Go Programming Language", Author: "Alan Donovan", ISBN: "9780134190440", Price: 49.99, PublishedAt: "2015-11-16T00:00:00Z"},
	{Title: "Clean Code", Author: "Robert C. Martin", ISBN: "9780132350884", Price: 39.99, PublishedAt: "2008-08-11T00:00:00Z"},
	{Title: "Design Patterns", Author: "Gang of Four", ISBN: "9780201633610", Price: 54.99, PublishedAt: "1994-10-21T00:00:00Z"},
	{Title: "Refactoring", Author: "Martin Fowler", ISBN: "9780201485677", Price: 47.99, PublishedAt: "1999-07-08T00:00:00Z"},
	{Title: "Head First Design Patterns", Author: "Eric Freeman", ISBN: "9780596007126", Price: 44.99, PublishedAt: "2004-10-25T00:00:00Z"},
	{Title: "Clean Architecture", Author: "Robert C. Martin", ISBN: "9780134494166", Price: 42.99, PublishedAt: "2017-09-20T00:00:00Z"},
	{Title: "Effective Go", Author: "Go Team", ISBN: "9781234567890", Price: 35.99, PublishedAt: "2020-01-15T00:00:00Z"},
	{Title: "Concurrency in Go", Author: "Katherine Cox-Buday", ISBN: "9781491941195", Price: 39.99, PublishedAt: "2017-07-19T00:00:00Z"},
	{Title: "Go in Action", Author: "William Kennedy", ISBN: "9781617291784", Price: 44.99, PublishedAt: "2015-11-04T00:00:00Z"},
	{Title: "Learning Go", Author: "Jon Bodner", ISBN: "9781492077213", Price: 49.99, PublishedAt: "2021-03-02T00:00:00Z"},
	{Title: "Microservices Patterns", Author: "Chris Richardson", ISBN: "9781617294549", Price: 59.99, PublishedAt: "2018-10-25T00:00:00Z"},
	{Title: "Building Microservices", Author: "Sam Newman", ISBN: "9781491950357", Price: 54.99, PublishedAt: "2015-02-20T00:00:00Z"},
	{Title: "Domain-Driven Design", Author: "Eric Evans", ISBN: "9780321125217", Price: 64.99, PublishedAt: "2003-08-22T00:00:00Z"},
	{Title: "The Pragmatic Programmer", Author: "David Thomas", ISBN: "9780201616224", Price: 49.99, PublishedAt: "1999-10-30T00:00:00Z"},
	{Title: "Code Complete", Author: "Steve McConnell", ISBN: "9780735619678", Price: 59.99, PublishedAt: "2004-06-09T00:00:00Z"},
}

var createdBookIDs []int
var idsMutex sync.Mutex

func addBookID(id int) {
	idsMutex.Lock()
	defer idsMutex.Unlock()
	createdBookIDs = append(createdBookIDs, id)
}

func getRandomBookID() int {
	idsMutex.Lock()
	defer idsMutex.Unlock()
	
	if len(createdBookIDs) == 0 {
		return 0
	}
	
	idx := rand.Intn(len(createdBookIDs))
	return createdBookIDs[idx]
}

func removeBookID(id int) {
	idsMutex.Lock()
	defer idsMutex.Unlock()
	
	for i, bookID := range createdBookIDs {
		if bookID == id {
			createdBookIDs = append(createdBookIDs[:i], createdBookIDs[i+1:]...)
			break
		}
	}
}

func getBookCount() int {
	idsMutex.Lock()
	defer idsMutex.Unlock()
	return len(createdBookIDs)
}

// Initialize with some books at startup
func initializeBooks(stats *Stats) {
	fmt.Println("Initializing with sample books...")
	
	for i := 0; i < 10; i++ {
		createBook(stats, true)
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("Initialized with %d books\n", getBookCount())
}

func createBook(stats *Stats, silent ...bool) {
	book := sampleBooks[rand.Intn(len(sampleBooks))]
	
	// Make ISBN unique by adding random suffix
	book.ISBN = book.ISBN + fmt.Sprintf("%d", rand.Intn(10000))
	
	// Randomize price slightly
	book.Price = book.Price + float64(rand.Intn(20)) - 10
	if book.Price < 0 {
		book.Price = 9.99
	}
	
	jsonData, _ := json.Marshal(book)
	
	resp, err := http.Post(BaseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		if len(silent) == 0 {
			log.Printf("Error creating book: %v", err)
		}
		stats.Increment("create", false)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusCreated {
		var createdBook Book
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &createdBook); err == nil {
			addBookID(createdBook.ID)
		}
		stats.Increment("create", true)
	} else {
		stats.Increment("create", false)
	}
}

func readBook(stats *Stats) {
	// 60% chance to read all books, 40% chance to read specific book
	if rand.Float32() < 0.6 {
		// Read all books
		resp, err := http.Get(BaseURL)
		if err != nil {
			log.Printf("Error reading all books: %v", err)
			stats.Increment("read", false)
			return
		}
		defer resp.Body.Close()
		
		stats.Increment("read", resp.StatusCode == http.StatusOK)
	} else {
		// Read specific book
		bookID := getRandomBookID()
		if bookID == 0 {
			// If no books exist, create one first
			createBook(stats, true)
			bookID = getRandomBookID()
		}
		
		if bookID != 0 {
			resp, err := http.Get(fmt.Sprintf("%s/%d", BaseURL, bookID))
			if err != nil {
				log.Printf("Error reading book %d: %v", bookID, err)
				stats.Increment("read", false)
				return
			}
			defer resp.Body.Close()
			
			stats.Increment("read", resp.StatusCode == http.StatusOK)
		} else {
			stats.Increment("read", false)
		}
	}
}

func updateBook(stats *Stats) {
	bookID := getRandomBookID()
	if bookID == 0 {
		// If no books exist, create one first
		createBook(stats, true)
		bookID = getRandomBookID()
	}
	
	if bookID == 0 {
		stats.Increment("update", false)
		return
	}
	
	// Create random update
	updates := UpdateBook{}
	
	if rand.Float32() < 0.4 {
		newTitle := sampleBooks[rand.Intn(len(sampleBooks))].Title + " - Updated"
		updates.Title = &newTitle
	}
	
	if rand.Float32() < 0.3 {
		newAuthor := sampleBooks[rand.Intn(len(sampleBooks))].Author
		updates.Author = &newAuthor
	}
	
	if rand.Float32() < 0.5 {
		newPrice := 19.99 + float64(rand.Intn(60))
		updates.Price = &newPrice
	}
	
	jsonData, _ := json.Marshal(updates)
	
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", BaseURL, bookID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error updating book %d: %v", bookID, err)
		stats.Increment("update", false)
		return
	}
	defer resp.Body.Close()
	
	success := resp.StatusCode == http.StatusOK
	if !success && resp.StatusCode == http.StatusNotFound {
		// Book was deleted by another worker, remove from our list
		removeBookID(bookID)
	}
	
	stats.Increment("update", success)
}

func deleteBook(stats *Stats) {
	bookID := getRandomBookID()
	if bookID == 0 {
		stats.Increment("delete", false)
		return
	}
	
	// Only delete if we have more than 5 books to maintain some inventory
	if getBookCount() <= 5 {
		stats.Increment("delete", false)
		return
	}
	
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", BaseURL, bookID), nil)
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error deleting book %d: %v", bookID, err)
		stats.Increment("delete", false)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNoContent {
		removeBookID(bookID)
		stats.Increment("delete", true)
	} else {
		if resp.StatusCode == http.StatusNotFound {
			removeBookID(bookID)
		}
		stats.Increment("delete", false)
	}
}

func worker(wg *sync.WaitGroup, stats *Stats, duration time.Duration) {
	defer wg.Done()
	
	startTime := time.Now()
	
	for time.Since(startTime) < duration {
		// Weight operations: 35% create, 40% read, 20% update, 5% delete
		operation := rand.Float32()
		
		switch {
		case operation < 0.35:
			createBook(stats)
		case operation < 0.75:
			readBook(stats)
		case operation < 0.95:
			updateBook(stats)
		default:
			deleteBook(stats)
		}
		
		// Random delay between requests (50-300ms)
		time.Sleep(time.Duration(rand.Intn(250)+50) * time.Millisecond)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	// Parse command line arguments
	requests := DefaultRequests
	concurrency := DefaultConcurrency
	duration := DefaultDuration
	
	if len(os.Args) > 1 {
		if r, err := strconv.Atoi(os.Args[1]); err == nil {
			requests = r
		}
	}
	
	if len(os.Args) > 2 {
		if c, err := strconv.Atoi(os.Args[2]); err == nil {
			concurrency = c
		}
	}
	
	if len(os.Args) > 3 {
		if d, err := strconv.Atoi(os.Args[3]); err == nil {
			duration = d
		}
	}
	
	// Check if API is reachable
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		log.Fatal("Error: API server is not reachable. Make sure it's running on localhost:8080")
	}
	resp.Body.Close()
	
	stats := &Stats{}
	
	// Initialize with some books
	initializeBooks(stats)
	
	fmt.Printf("\nStarting load test with %d concurrent workers for %d seconds...\n", concurrency, duration)
	fmt.Printf("Target: ~%d requests\n", requests)
	fmt.Printf("API: %s\n", BaseURL)
	fmt.Printf("Press Ctrl+C to stop early\n\n")
	
	// Calculate duration
	testDuration := time.Duration(duration) * time.Second
	
	startTime := time.Now()
	
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(&wg, stats, testDuration)
	}
	
	// Print progress every 5 seconds
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				stats.mutex.Lock()
				elapsed := time.Since(startTime).Seconds()
				rps := float64(stats.Total) / elapsed
				fmt.Printf("Progress: %d requests in %.1fs (%.1f req/s) | Books: %d\n", 
					stats.Total, elapsed, rps, getBookCount())
				stats.mutex.Unlock()
			case <-done:
				return
			}
		}
	}()
	
	wg.Wait()
	close(done)
	
	elapsed := time.Since(startTime)
	
	fmt.Printf("\nTest completed in %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("Average RPS: %.2f\n", float64(stats.Total)/elapsed.Seconds())
	fmt.Printf("Final book count: %d\n", getBookCount())
	
	stats.Print()
	
	fmt.Printf("\nMetrics are available at: http://localhost:8080/metrics\n")
	fmt.Printf("Prometheus: http://localhost:9090\n")
	fmt.Printf("Grafana: http://localhost:3000\n")
}