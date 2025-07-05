package models

import (
	"time"
)

type Book struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Author      string    `json:"author" db:"author"`
	ISBN        string    `json:"isbn" db:"isbn"`
	Price       float64   `json:"price" db:"price"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateBookRequest struct {
	Title       string    `json:"title" binding:"required"`
	Author      string    `json:"author" binding:"required"`
	ISBN        string    `json:"isbn" binding:"required"`
	Price       float64   `json:"price" binding:"required,min=0"`
	PublishedAt time.Time `json:"published_at" binding:"required"`
}

type UpdateBookRequest struct {
	Title       *string    `json:"title,omitempty"`
	Author      *string    `json:"author,omitempty"`
	ISBN        *string    `json:"isbn,omitempty"`
	Price       *float64   `json:"price,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}