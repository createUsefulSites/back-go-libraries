package controllers

import (
	"net/http"
	"strconv"
	"time"

	"library-app/database"
	"library-app/models"

	"github.com/gin-gonic/gin"
)

// BookResponse представляет структуру ответа для книги с авторами и жанром
type BookResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description"`
	PublicationYear *int      `json:"publication_year"`
	ISBN            string    `json:"isbn"`
	Genre           string    `json:"genre"`
	TotalCopies     int       `json:"total_copies"`
	AvailableCopies int       `json:"available_copies"`
	CoverURL        *string   `json:"cover_url"`
	AddedDate       time.Time `json:"added_date"`
	Authors         []string  `json:"authors"`
}

// CreateBookRequest представляет структуру запроса для создания книги
type CreateBookRequest struct {
	Title           string  `json:"title" binding:"required"`
	Description     *string `json:"description"`
	PublicationYear *int    `json:"publication_year"`
	ISBN            string  `json:"isbn" binding:"required"`
	GenreID         uint    `json:"genre_id" binding:"required"`
	TotalCopies     int     `json:"total_copies" binding:"required,gt=0"`
	CoverURL        *string `json:"cover_url"`
	AuthorIDs       []uint  `json:"author_ids" binding:"required"`
}

// ReviewRequest представляет структуру запроса для отзыва
type ReviewRequest struct {
	Rating  int     `json:"rating" binding:"required,gte=1,lte=5"`
	Comment *string `json:"comment"`
}

// GetBooks возвращает список всех книг с авторами и жанрами
func GetBooks(c *gin.Context) {
	var books []models.Book
	if err := database.DB.Preload("Genre").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
		return
	}

	var response []BookResponse
	for _, book := range books {
		// Получаем авторов книги
		var bookAuthors []models.BookAuthor
		if err := database.DB.Where("book_id = ?", book.ID).Preload("Author").Find(&bookAuthors).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authors"})
			return
		}

		authors := make([]string, 0, len(bookAuthors))
		for _, ba := range bookAuthors {
			authors = append(authors, ba.Author.Name)
		}

		response = append(response, BookResponse{
			ID:              book.ID,
			Title:           book.Title,
			Description:     book.Description,
			PublicationYear: book.PublicationYear,
			ISBN:            book.ISBN,
			Genre:           book.Genre.Name,
			TotalCopies:     book.TotalCopies,
			AvailableCopies: book.AvailableCopies,
			CoverURL:        book.CoverURL,
			AddedDate:       book.AddedDate,
			Authors:         authors,
		})
	}

	c.JSON(http.StatusOK, response)
}

// CreateBook создаёт новую книгу (доступно только админу)
func CreateBook(c *gin.Context) {
	var req CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли жанр
	var genre models.Genre
	if err := database.DB.First(&genre, req.GenreID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Genre not found"})
		return
	}

	// Проверяем, существуют ли авторы
	for _, authorID := range req.AuthorIDs {
		var author models.Author
		if err := database.DB.First(&author, authorID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more authors not found"})
			return
		}
	}

	// Создаём книгу
	book := models.Book{
		Title:           req.Title,
		Description:     req.Description,
		PublicationYear: req.PublicationYear,
		ISBN:            req.ISBN,
		GenreID:         req.GenreID,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.TotalCopies,
		CoverURL:        req.CoverURL,
	}

	if err := database.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create book"})
		return
	}

	// Создаём связи с авторами
	for _, authorID := range req.AuthorIDs {
		bookAuthor := models.BookAuthor{
			BookID:   book.ID,
			AuthorID: authorID,
		}
		if err := database.DB.Create(&bookAuthor).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate authors with book"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Book created successfully", "book_id": book.ID})
}

// DeleteBook удаляет книгу по ID (доступно только админу)
func DeleteBook(c *gin.Context) {
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var book models.Book
	if err := database.DB.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	if err := database.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

// BorrowBook позволяет пользователю взять книгу
func BorrowBook(c *gin.Context) {
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var book models.Book
	if err := database.DB.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Проверяем доступность книги
	if book.AvailableCopies <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No available copies of this book"})
		return
	}

	// Проверяем, не взял ли пользователь уже эту книгу
	var existingOrder models.Order
	if err := database.DB.Where("user_id = ? AND book_id = ? AND status = ?", userID, bookID, "issued").First(&existingOrder).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already borrowed this book"})
		return
	}

	// Создаём заказ
	dueDate := time.Now().AddDate(0, 0, 14) // Срок возврата через 14 дней
	order := models.Order{
		UserID:  userID.(uint),
		BookID:  uint(bookID),
		DueDate: dueDate,
		Status:  "issued",
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Обновляем количество доступных копий
	book.AvailableCopies--
	if err := database.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book borrowed successfully", "due_date": dueDate})
}

// ReturnBook позволяет пользователю вернуть книгу
func ReturnBook(c *gin.Context) {
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Ищем активный заказ
	var order models.Order
	if err := database.DB.Where("user_id = ? AND book_id = ? AND status = ?", userID, bookID, "issued").First(&order).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active borrow record found for this book"})
		return
	}

	// Обновляем статус заказа
	now := time.Now()
	order.ReturnDate = &now
	if now.After(order.DueDate) {
		order.Status = "overdue"
	} else {
		order.Status = "returned"
	}

	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	// Обновляем количество доступных копий
	var book models.Book
	if err := database.DB.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book"})
		return
	}

	book.AvailableCopies++
	if err := database.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book returned successfully"})
}

// AddReview позволяет пользователю добавить отзыв на книгу
func AddReview(c *gin.Context) {
	bookID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли книга
	var book models.Book
	if err := database.DB.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Проверяем, брал ли пользователь книгу
	var order models.Order
	if err := database.DB.Where("user_id = ? AND book_id = ? AND status IN (?, ?)", userID, bookID, "returned", "overdue").First(&order).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You must borrow and return the book before reviewing"})
		return
	}

	// Проверяем, не оставлял ли пользователь уже отзыв
	var existingReview models.Review
	if err := database.DB.Where("user_id = ? AND book_id = ?", userID, bookID).First(&existingReview).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reviewed this book"})
		return
	}

	// Создаём отзыв
	review := models.Review{
		UserID:  userID.(uint),
		BookID:  uint(bookID),
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	if err := database.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Review added successfully", "review_id": review.ID})
}
