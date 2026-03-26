package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"health": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "ok", nil
					},
				},
			},
		}),
	})
	if err != nil {
		panic(err)
	}

	r := gin.Default()
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

