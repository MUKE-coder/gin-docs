package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MUKE-coder/gin-docs/gindocs"
	"github.com/gin-gonic/gin"
)

// ============================================================
// Models
// ============================================================

// User represents a user account.
type User struct {
	ID        uint      `json:"id" gorm:"primarykey" docs:"description:Unique identifier"`
	Name      string    `json:"name" binding:"required,min=2,max=100" docs:"example:John Doe"`
	Email     string    `json:"email" binding:"required,email" gorm:"size:200;uniqueIndex" docs:"example:user@example.com"`
	Role      string    `json:"role" binding:"oneof=admin user moderator" gorm:"default:'user'" docs:"description:User role"`
	Bio       string    `json:"bio,omitempty" gorm:"type:text" docs:"description:User biography"`
	Avatar    string    `json:"avatar,omitempty" docs:"format:uri,description:URL to user avatar"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" docs:"description:Account creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Post represents a blog post.
type Post struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Title       string    `json:"title" binding:"required,min=1,max=200" docs:"example:Getting Started with Go"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;size:250" docs:"description:URL-friendly identifier"`
	Content     string    `json:"content" binding:"required" docs:"description:Post body in markdown"`
	Excerpt     string    `json:"excerpt,omitempty" gorm:"size:500" docs:"description:Short summary for listings"`
	AuthorID    uint      `json:"author_id" docs:"description:ID of the post author"`
	CategoryID  uint      `json:"category_id,omitempty" docs:"description:ID of the post category"`
	Published   bool      `json:"published" gorm:"default:false"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Tags        []string  `json:"tags,omitempty" gorm:"serializer:json" docs:"description:List of tags"`
	ViewCount   int       `json:"view_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Body      string    `json:"body" binding:"required,min=1" docs:"description:Comment text"`
	PostID    uint      `json:"post_id" docs:"description:ID of the parent post"`
	UserID    uint      `json:"user_id" docs:"description:ID of the commenter"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Tag represents a content tag.
type Tag struct {
	ID   uint   `json:"id" gorm:"primarykey"`
	Name string `json:"name" binding:"required" gorm:"uniqueIndex;size:100"`
	Slug string `json:"slug" gorm:"uniqueIndex;size:100"`
}

// Category represents a post category.
type Category struct {
	ID          uint   `json:"id" gorm:"primarykey"`
	Name        string `json:"name" binding:"required" gorm:"uniqueIndex;size:100"`
	Slug        string `json:"slug" gorm:"uniqueIndex;size:100"`
	Description string `json:"description,omitempty" gorm:"size:500"`
}

// LoginRequest represents login credentials.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" docs:"example:user@example.com"`
	Password string `json:"password" binding:"required,min=8" docs:"example:********"`
}

// LoginResponse represents a successful login.
type LoginResponse struct {
	Token     string `json:"token" docs:"description:JWT access token"`
	ExpiresAt string `json:"expires_at" docs:"description:Token expiration time"`
}

// RegisterRequest represents registration data.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100" docs:"example:Jane Doe"`
	Email    string `json:"email" binding:"required,email" docs:"example:jane@example.com"`
	Password string `json:"password" binding:"required,min=8" docs:"example:SecureP@ss1"`
}

// PaginatedResponse wraps a list with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total" docs:"description:Total number of records"`
	Page       int         `json:"page" docs:"description:Current page number"`
	PerPage    int         `json:"per_page" docs:"description:Records per page"`
	TotalPages int         `json:"total_pages" docs:"description:Total number of pages"`
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error   string `json:"error" docs:"description:Error message"`
	Code    string `json:"code,omitempty" docs:"description:Machine-readable error code"`
	Details string `json:"details,omitempty" docs:"description:Additional error details"`
}

// SearchResult represents a search result.
type SearchResult struct {
	Type    string      `json:"type" docs:"description:Result type (user or post),enum:user|post"`
	ID      uint        `json:"id"`
	Title   string      `json:"title" docs:"description:Result title or name"`
	Excerpt string      `json:"excerpt,omitempty" docs:"description:Text snippet"`
	Score   float64     `json:"score" docs:"description:Relevance score"`
}

// ============================================================
// Handlers
// ============================================================

// --- Auth ---

func registerHandler(c *gin.Context) {
	var input RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    User{ID: 1, Name: input.Name, Email: input.Email, Role: "user", IsActive: true},
	})
}

func loginHandler(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, LoginResponse{
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}

// --- Users ---

func listUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []User{{ID: 1, Name: "John Doe", Email: "john@example.com", Role: "admin", IsActive: true}},
		Total:      1,
		Page:       page,
		PerPage:    perPage,
		TotalPages: 1,
	})
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, User{ID: 1, Name: "John Doe", Email: "john@example.com", Role: "admin", IsActive: true, Bio: "Admin user", CreatedAt: time.Now().Add(-24 * time.Hour * 30)})
	_ = id
}

func createUser(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	input.ID = uint(rand.Intn(1000) + 1)
	input.CreatedAt = time.Now()
	c.JSON(http.StatusCreated, input)
}

func updateUser(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, input)
}

func deleteUser(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// --- Posts ---

func listPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       []Post{{ID: 1, Title: "Hello World", Slug: "hello-world", Content: "First post!", AuthorID: 1, Published: true}},
		Total:      1,
		Page:       page,
		PerPage:    perPage,
		TotalPages: 1,
	})
}

func getPost(c *gin.Context) {
	c.JSON(http.StatusOK, Post{ID: 1, Title: "Hello World", Slug: "hello-world", Content: "Full post content here.", AuthorID: 1, Published: true, ViewCount: 42})
}

func createPost(c *gin.Context) {
	var input Post
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	input.ID = uint(rand.Intn(1000) + 1)
	c.JSON(http.StatusCreated, input)
}

func updatePost(c *gin.Context) {
	var input Post
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, input)
}

func deletePost(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// --- Comments ---

func listPostComments(c *gin.Context) {
	c.JSON(http.StatusOK, []Comment{{ID: 1, Body: "Great post!", PostID: 1, UserID: 1, CreatedAt: time.Now()}})
}

func createPostComment(c *gin.Context) {
	var input Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	input.ID = uint(rand.Intn(1000) + 1)
	c.JSON(http.StatusCreated, input)
}

// --- Tags & Categories ---

func listTags(c *gin.Context) {
	c.JSON(http.StatusOK, []Tag{
		{ID: 1, Name: "Go", Slug: "go"},
		{ID: 2, Name: "Web", Slug: "web"},
	})
}

func listCategories(c *gin.Context) {
	c.JSON(http.StatusOK, []Category{
		{ID: 1, Name: "Tutorials", Slug: "tutorials", Description: "Step-by-step guides"},
		{ID: 2, Name: "News", Slug: "news", Description: "Latest updates"},
	})
}

// --- Search ---

func searchHandler(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "query parameter 'q' is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   q,
		"results": []SearchResult{},
		"total":   0,
	})
}

// --- Upload ---

func uploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"filename": file.Filename,
		"size":     file.Size,
		"url":      fmt.Sprintf("/uploads/%s", file.Filename),
	})
}

// ============================================================
// Middleware
// ============================================================

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error: "missing or invalid authorization token",
				Code:  "UNAUTHORIZED",
			})
			return
		}
		c.Set("user_id", uint(1))
		c.Next()
	}
}

// ============================================================
// Main
// ============================================================

func main() {
	router := gin.Default()

	// --- Public auth routes ---
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", registerHandler)
		auth.POST("/login", loginHandler)
	}

	// --- Public read routes ---
	router.GET("/api/users", listUsers)
	router.GET("/api/users/:id", getUser)
	router.GET("/api/posts", listPosts)
	router.GET("/api/posts/:id", getPost)
	router.GET("/api/posts/:id/comments", listPostComments)
	router.GET("/api/tags", listTags)
	router.GET("/api/categories", listCategories)
	router.GET("/api/search", searchHandler)

	// --- Protected write routes ---
	protected := router.Group("/api")
	protected.Use(authMiddleware())
	{
		protected.POST("/users", createUser)
		protected.PUT("/users/:id", updateUser)
		protected.DELETE("/users/:id", deleteUser)

		protected.POST("/posts", createPost)
		protected.PUT("/posts/:id", updatePost)
		protected.DELETE("/posts/:id", deletePost)

		protected.POST("/posts/:id/comments", createPostComment)
		protected.POST("/upload", uploadHandler)
	}

	// --- Mount Gin Docs ---
	docs := gindocs.Mount(router, nil, gindocs.Config{
		Title:       "Blog API",
		Description: "A full-featured blog API built with Go and Gin — auto-documented by Gin Docs.\n\nThis API provides endpoints for user management, blog posts, comments, tags, categories, and search.",
		Version:     "1.0.0",
		UI:          gindocs.UIScalar,
		DevMode:     true,
		Auth: gindocs.AuthConfig{
			Type:         gindocs.AuthBearer,
			BearerFormat: "JWT",
		},
		Servers: []gindocs.ServerInfo{
			{URL: "http://localhost:8080", Description: "Local development"},
		},
		Contact: gindocs.ContactInfo{
			Name:  "API Support",
			Email: "support@example.com",
			URL:   "https://github.com/MUKE-coder/gin-docs",
		},
		License: gindocs.LicenseInfo{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
		Models: []interface{}{
			User{},
			Post{},
			Comment{},
			Tag{},
			Category{},
		},
		ExcludePrefixes: []string{"/api/upload"},
		CustomSections: []gindocs.Section{
			{
				Title: "Authentication",
				Content: "This API uses JWT Bearer tokens for authentication.\n\n" +
					"1. Register a new account via POST /api/auth/register\n" +
					"2. Login via POST /api/auth/login to receive a JWT token\n" +
					"3. Include the token in the Authorization header: Bearer <token>\n\n" +
					"Read endpoints are public. Write endpoints require authentication.",
			},
			{
				Title: "Pagination",
				Content: "List endpoints support pagination via query parameters:\n\n" +
					"- page: Page number (default: 1)\n" +
					"- per_page: Items per page (default: 20, max: 100)\n\n" +
					"Responses include pagination metadata: total, page, per_page, total_pages.",
			},
		},
	})

	// --- Route overrides for richer documentation ---
	docs.Route("POST /api/auth/register").
		Summary("Register a new account").
		Description("Creates a new user account. Returns the created user object.").
		RequestBody(RegisterRequest{}).
		Response(201, User{}, "User created successfully").
		Response(400, ErrorResponse{}, "Validation error").
		Response(409, ErrorResponse{}, "Email already in use").
		Tags("Authentication")

	docs.Route("POST /api/auth/login").
		Summary("Login to get a JWT token").
		Description("Authenticates with email and password. Returns a JWT token for API access.").
		RequestBody(LoginRequest{}).
		Response(200, LoginResponse{}, "Login successful").
		Response(400, ErrorResponse{}, "Invalid credentials").
		Tags("Authentication")

	docs.Route("POST /api/users").
		Summary("Create a new user").
		RequestBody(User{}).
		Response(201, User{}, "User created").
		Response(400, ErrorResponse{}, "Validation error")

	docs.Route("GET /api/users").
		Summary("List all users").
		Response(200, PaginatedResponse{}, "Paginated list of users")

	docs.Route("GET /api/users/:id").
		Summary("Get user by ID").
		Response(200, User{}, "User details").
		Response(404, ErrorResponse{}, "User not found")

	docs.Route("POST /api/posts").
		Summary("Create a new blog post").
		RequestBody(Post{}).
		Response(201, Post{}, "Post created").
		Response(400, ErrorResponse{}, "Validation error")

	docs.Route("GET /api/posts").
		Summary("List all blog posts").
		Response(200, PaginatedResponse{}, "Paginated list of posts")

	docs.Route("GET /api/posts/:id").
		Summary("Get post by ID").
		Response(200, Post{}, "Post details").
		Response(404, ErrorResponse{}, "Post not found")

	docs.Route("GET /api/search").
		Summary("Search posts and users").
		Description("Full-text search across posts and users. Returns results ranked by relevance.").
		Response(200, nil, "Search results")

	// --- Group override for auth-protected routes ---
	docs.Group("/api/users/*").Security("bearerAuth")
	docs.Group("/api/posts/*").Security("bearerAuth")

	fmt.Println("===========================================")
	fmt.Println("  Gin Docs Demo — Blog API")
	fmt.Println("  API:      http://localhost:8080/api")
	fmt.Println("  Docs:     http://localhost:8080/docs")
	fmt.Println("  Swagger:  http://localhost:8080/docs?ui=swagger")
	fmt.Println("  Spec:     http://localhost:8080/docs/openapi.json")
	fmt.Println("  YAML:     http://localhost:8080/docs/openapi.yaml")
	fmt.Println("  Postman:  http://localhost:8080/docs/export/postman")
	fmt.Println("  Insomnia: http://localhost:8080/docs/export/insomnia")
	fmt.Println("===========================================")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
