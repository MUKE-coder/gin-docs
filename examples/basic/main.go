// Package main demonstrates the simplest possible Gin Docs setup.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MUKE-coder/gin-docs/gindocs"
	"github.com/gin-gonic/gin"
)

type Item struct {
	ID    uint   `json:"id" gorm:"primarykey"`
	Name  string `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gte=0"`
}

func main() {
	r := gin.Default()

	r.GET("/api/items", func(c *gin.Context) {
		c.JSON(http.StatusOK, []Item{
			{ID: 1, Name: "Widget", Price: 9.99},
		})
	})

	r.GET("/api/items/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, Item{ID: 1, Name: "Widget", Price: 9.99})
	})

	r.POST("/api/items", func(c *gin.Context) {
		var item Item
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	})

	// One line to add docs!
	gindocs.Mount(r, nil)

	fmt.Println("Docs at http://localhost:8080/docs")
	log.Fatal(r.Run(":8080"))
}
