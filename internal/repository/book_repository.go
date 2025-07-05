package repository

import (
	"database/sql"
	"fmt"
	"gin-prometheus-grafana/internal/models"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "db_query_duration_seconds",
			Help: "Duration of database queries in seconds",
		},
		[]string{"operation", "table"},
	)

	dbQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) CreateBook(book *models.CreateBookRequest) (*models.Book, error) {
	start := time.Now()
	defer func() {
		dbQueryDuration.WithLabelValues("create", "books").Observe(time.Since(start).Seconds())
	}()

	query := `
		INSERT INTO books (title, author, isbn, price, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, author, isbn, price, published_at, created_at, updated_at
	`
	
	now := time.Now()
	row := r.db.QueryRow(query, book.Title, book.Author, book.ISBN, book.Price, book.PublishedAt, now, now)
	
	var result models.Book
	err := row.Scan(&result.ID, &result.Title, &result.Author, &result.ISBN, &result.Price, &result.PublishedAt, &result.CreatedAt, &result.UpdatedAt)
	
	if err != nil {
		dbQueryTotal.WithLabelValues("create", "books", "error").Inc()
		log.Printf("Error creating book: %v", err)
		return nil, err
	}
	
	dbQueryTotal.WithLabelValues("create", "books", "success").Inc()
	log.Printf("Created book: ID=%d, Title=%s", result.ID, result.Title)
	return &result, nil
}

func (r *BookRepository) GetBookByID(id int) (*models.Book, error) {
	start := time.Now()
	defer func() {
		dbQueryDuration.WithLabelValues("select", "books").Observe(time.Since(start).Seconds())
	}()

	query := `
		SELECT id, title, author, isbn, price, published_at, created_at, updated_at
		FROM books WHERE id = $1
	`
	
	row := r.db.QueryRow(query, id)
	var book models.Book
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Price, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			dbQueryTotal.WithLabelValues("select", "books", "not_found").Inc()
			return nil, fmt.Errorf("book with id %d not found", id)
		}
		dbQueryTotal.WithLabelValues("select", "books", "error").Inc()
		log.Printf("Error getting book by ID %d: %v", id, err)
		return nil, err
	}
	
	dbQueryTotal.WithLabelValues("select", "books", "success").Inc()
	log.Printf("Retrieved book: ID=%d, Title=%s", book.ID, book.Title)
	return &book, nil
}

func (r *BookRepository) GetAllBooks() ([]models.Book, error) {
	start := time.Now()
	defer func() {
		dbQueryDuration.WithLabelValues("select_all", "books").Observe(time.Since(start).Seconds())
	}()

	query := `
		SELECT id, title, author, isbn, price, published_at, created_at, updated_at
		FROM books ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		dbQueryTotal.WithLabelValues("select_all", "books", "error").Inc()
		log.Printf("Error getting all books: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Price, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			dbQueryTotal.WithLabelValues("select_all", "books", "error").Inc()
			log.Printf("Error scanning book row: %v", err)
			return nil, err
		}
		books = append(books, book)
	}
	
	// Ensure we return an empty slice instead of nil for consistent JSON serialization
	if books == nil {
		books = []models.Book{}
	}
	
	dbQueryTotal.WithLabelValues("select_all", "books", "success").Inc()
	log.Printf("Retrieved %d books", len(books))
	return books, nil
}

func (r *BookRepository) UpdateBook(id int, req *models.UpdateBookRequest) (*models.Book, error) {
	start := time.Now()
	defer func() {
		dbQueryDuration.WithLabelValues("update", "books").Observe(time.Since(start).Seconds())
	}()

	existing, err := r.GetBookByID(id)
	if err != nil {
		return nil, err
	}
	
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Author != nil {
		existing.Author = *req.Author
	}
	if req.ISBN != nil {
		existing.ISBN = *req.ISBN
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.PublishedAt != nil {
		existing.PublishedAt = *req.PublishedAt
	}
	existing.UpdatedAt = time.Now()
	
	query := `
		UPDATE books 
		SET title = $1, author = $2, isbn = $3, price = $4, published_at = $5, updated_at = $6
		WHERE id = $7
		RETURNING id, title, author, isbn, price, published_at, created_at, updated_at
	`
	
	row := r.db.QueryRow(query, existing.Title, existing.Author, existing.ISBN, existing.Price, existing.PublishedAt, existing.UpdatedAt, id)
	
	var result models.Book
	err = row.Scan(&result.ID, &result.Title, &result.Author, &result.ISBN, &result.Price, &result.PublishedAt, &result.CreatedAt, &result.UpdatedAt)
	
	if err != nil {
		dbQueryTotal.WithLabelValues("update", "books", "error").Inc()
		log.Printf("Error updating book ID %d: %v", id, err)
		return nil, err
	}
	
	dbQueryTotal.WithLabelValues("update", "books", "success").Inc()
	log.Printf("Updated book: ID=%d, Title=%s", result.ID, result.Title)
	return &result, nil
}

func (r *BookRepository) DeleteBook(id int) error {
	start := time.Now()
	defer func() {
		dbQueryDuration.WithLabelValues("delete", "books").Observe(time.Since(start).Seconds())
	}()

	query := `DELETE FROM books WHERE id = $1`
	result, err := r.db.Exec(query, id)
	
	if err != nil {
		dbQueryTotal.WithLabelValues("delete", "books", "error").Inc()
		log.Printf("Error deleting book ID %d: %v", id, err)
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbQueryTotal.WithLabelValues("delete", "books", "error").Inc()
		return err
	}
	
	if rowsAffected == 0 {
		dbQueryTotal.WithLabelValues("delete", "books", "not_found").Inc()
		return fmt.Errorf("book with id %d not found", id)
	}
	
	dbQueryTotal.WithLabelValues("delete", "books", "success").Inc()
	log.Printf("Deleted book: ID=%d", id)
	return nil
}