package controllers

import (
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	// Реализация регистрации
	c.JSON(200, gin.H{"message": "Register endpoint not implemented"})
}

func Login(c *gin.Context) {
	// Реализация логина
	c.JSON(200, gin.H{"message": "Login endpoint not implemented"})
}
