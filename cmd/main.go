package main

import (
	"library-app/controllers"
	"library-app/database"
	middleware "library-app/middlewares"

	"github.com/gin-gonic/gin"
)

func init() {
	controllers.LoadEnv()
	database.InitDB()
}

func main() {
	r := gin.Default()

	// Маршруты аутентификации
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// Группа маршрутов с проверкой JWT
	protected := r.Group("/books")
	protected.Use(middleware.JWTMiddleware())
	{
		// GET /books — Список книг с авторами и жанрами
		protected.GET("", controllers.GetBooks)

		// POST /books/:id/borrow — Взять книгу
		protected.POST("/:id/borrow", controllers.BorrowBook)

		// POST /books/:id/return — Вернуть книгу
		protected.POST("/:id/return", controllers.ReturnBook)

		// POST /books/:id/review — Добавить отзыв
		protected.POST("/:id/review", controllers.AddReview)

		// Админские маршруты (дополнительная проверка роли)
		admin := protected.Group("")
		admin.Use(middleware.AdminMiddleware())
		{
			// POST /books — Добавить книгу
			admin.POST("", controllers.CreateBook)

			// DELETE /books/:id — Удалить книгу
			admin.DELETE("/:id", controllers.DeleteBook)
		}
	}

	// Запуск сервера
	r.Run(":8080") // запуск на localhost:8080
}
