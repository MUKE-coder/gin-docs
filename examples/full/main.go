// Package main demonstrates a fully configured Gin Docs setup.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MUKE-coder/gin-docs/gindocs"
	"github.com/gin-gonic/gin"
)

// --- Models ---

type Product struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" binding:"required,min=1,max=200" docs:"example:Wireless Mouse"`
	Description string    `json:"description,omitempty" gorm:"type:text" docs:"description:Product description"`
	Price       float64   `json:"price" binding:"required,gte=0" docs:"example:29.99"`
	SKU         string    `json:"sku" gorm:"uniqueIndex;size:50" docs:"description:Stock keeping unit"`
	InStock     bool      `json:"in_stock" gorm:"default:true"`
	CategoryID  uint      `json:"category_id,omitempty"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProductCategory struct {
	ID   uint   `json:"id" gorm:"primarykey"`
	Name string `json:"name" binding:"required" gorm:"uniqueIndex;size:100"`
	Slug string `json:"slug" gorm:"uniqueIndex;size:100"`
}

type Order struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	CustomerID uint      `json:"customer_id" binding:"required"`
	Status     string    `json:"status" binding:"oneof=pending processing shipped delivered cancelled" gorm:"default:'pending'"`
	Total      float64   `json:"total" binding:"required,gte=0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// --- Handlers ---

func listProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"products": []Product{}, "total": 0})
}

func getProduct(c *gin.Context) {
	c.JSON(http.StatusOK, Product{ID: 1, Name: "Wireless Mouse", Price: 29.99, SKU: "WM-001", InStock: true})
}

func createProduct(c *gin.Context) {
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func updateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func deleteProduct(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func listCategories(c *gin.Context) {
	c.JSON(http.StatusOK, []ProductCategory{
		{ID: 1, Name: "Electronics", Slug: "electronics"},
	})
}

func listOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"orders": []Order{}, "total": 0})
}

func getOrder(c *gin.Context) {
	c.JSON(http.StatusOK, Order{ID: 1, CustomerID: 1, Status: "pending", Total: 59.98})
}

func createOrder(c *gin.Context) {
	var o Order
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func main() {
	r := gin.Default()

	// --- Product routes ---
	products := r.Group("/api/products")
	{
		products.GET("", listProducts)
		products.POST("", createProduct)
		products.GET("/:id", getProduct)
		products.PUT("/:id", updateProduct)
		products.DELETE("/:id", deleteProduct)
	}

	// --- Category routes ---
	r.GET("/api/categories", listCategories)

	// --- Order routes ---
	orders := r.Group("/api/orders")
	{
		orders.GET("", listOrders)
		orders.POST("", createOrder)
		orders.GET("/:id", getOrder)
	}

	// --- Mount Gin Docs with full configuration ---
	docs := gindocs.Mount(r, nil, gindocs.Config{
		Title:       "E-Commerce API",
		Description: "A fully configured e-commerce API with Gin Docs.\n\nDemonstrates all configuration options including auth, custom sections, and route overrides.",
		Version:     "2.0.0",
		UI:          gindocs.UIScalar, // Use Scalar as default UI
		DevMode:     true,
		Auth: gindocs.AuthConfig{
			Type:         gindocs.AuthBearer,
			BearerFormat: "JWT",
		},
		Servers: []gindocs.ServerInfo{
			{URL: "http://localhost:8080", Description: "Local development"},
			{URL: "https://api.example.com", Description: "Production"},
		},
		Contact: gindocs.ContactInfo{
			Name:  "Dev Team",
			Email: "dev@example.com",
			URL:   "https://example.com",
		},
		License: gindocs.LicenseInfo{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
		Models: []interface{}{
			Product{},
			ProductCategory{},
			Order{},
		},
		CustomSections: []gindocs.Section{
			{
				Title:   "Getting Started",
				Content: "1. Browse products via GET /api/products\n2. Create an order via POST /api/orders\n3. Track your order via GET /api/orders/:id",
			},
		},
	})

	// --- Route overrides ---
	docs.Route("POST /api/products").
		Summary("Add a new product").
		RequestBody(Product{}).
		Response(201, Product{}, "Product created")

	docs.Route("POST /api/orders").
		Summary("Place a new order").
		RequestBody(Order{}).
		Response(201, Order{}, "Order placed")

	fmt.Println("Full example running at http://localhost:8080/docs")
	log.Fatal(r.Run(":8080"))
}
