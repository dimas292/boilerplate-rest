package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standardized API response envelope.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"per_page"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// PaginationQuery represents incoming pagination parameters.
type PaginationQuery struct {
	Page    int `form:"page,default=1" binding:"min=1"`
	PerPage int `form:"per_page,default=10" binding:"min=1,max=100"`
}

// Offset calculates the database offset from page and per_page.
func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Success sends a 200 JSON response.
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 JSON response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

// Paginated sends a 200 JSON response with pagination metadata.
func Paginated(c *gin.Context, message string, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

// Error sends an error JSON response with the given status code.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Status:  statusCode,
		Message: message,
	})
}
