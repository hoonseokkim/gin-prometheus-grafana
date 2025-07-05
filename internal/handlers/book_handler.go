package handlers

import (
	"gin-prometheus-grafana/internal/models"
	"gin-prometheus-grafana/internal/repository"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	repo *repository.BookRepository
}

func NewBookHandler(repo *repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) CreateBook(c *gin.Context) {
	var req models.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.repo.CreateBook(&req)
	if err != nil {
		log.Printf("Failed to create book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create book"})
		return
	}

	log.Printf("Successfully created book: %+v", book)
	c.JSON(http.StatusCreated, book)
}

func (h *BookHandler) GetBookByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid book ID: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	book, err := h.repo.GetBookByID(id)
	if err != nil {
		log.Printf("Failed to get book by ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	log.Printf("Successfully retrieved book: %+v", book)
	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) GetAllBooks(c *gin.Context) {
	books, err := h.repo.GetAllBooks()
	if err != nil {
		log.Printf("Failed to get all books: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve books"})
		return
	}

	log.Printf("Successfully retrieved %d books", len(books))
	c.JSON(http.StatusOK, books)
}

func (h *BookHandler) UpdateBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid book ID: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var req models.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.repo.UpdateBook(id, &req)
	if err != nil {
		log.Printf("Failed to update book ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	log.Printf("Successfully updated book: %+v", book)
	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid book ID: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	err = h.repo.DeleteBook(id)
	if err != nil {
		log.Printf("Failed to delete book ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	log.Printf("Successfully deleted book ID %d", id)
	c.JSON(http.StatusNoContent, nil)
}