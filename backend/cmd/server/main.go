package main

import (
	"log"
	"net/http"
	"os"

	"smart-study-assist-api/internal/auth"
	"smart-study-assist-api/internal/database"
	"smart-study-assist-api/internal/handlers"
	"smart-study-assist-api/internal/study"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize MongoDB
	client := database.ConnectDB()
	_ = client
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	studyService := study.NewService(client)
	schema, err := study.BuildSchema(studyService)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	authHandler := handlers.NewAuthHandler(client)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	protected := r.Group("/api")
	protected.Use(auth.AuthMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/graphql", func(c *gin.Context) {
		var req struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  req.Query,
			VariableValues: req.Variables,
		})
		c.JSON(http.StatusOK, result)
	})

	_ = r.Run(":" + port)
}

