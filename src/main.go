package main

import (
	"fmt"

	"github.com/Graham-04/ironkey-api/handlers"
	"github.com/Graham-04/ironkey-api/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dataStore := sql.GetDataStore("mysql")
	dataStore.InitDB()

	dataStore.AddUser(sql.User{
		Email:     "email@email.com",
		Password:  "password hash",
		FirstName: "first name",
		LastName:  "last name",
	})

	user := dataStore.GetUser("email@meail.com", "");
	fmt.Println(user)

	dataStore.Search("us")

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("datastore", dataStore)
		c.Next()
	})

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

	r.GET("/user", handlers.GetUser)
	r.GET("users", handlers.GetUsers)
	r.POST("/user", handlers.AddUser)
	r.PATCH("/user", handlers.UpdateUser)
	r.DELETE("/user", handlers.DeleteUser)
	r.GET("/search", handlers.Search)

	r.Run("127.0.0.1:8080")

}
