package handlers

import (
	"fmt"
	"github.com/Graham-04/ironkey-api/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddUser(c *gin.Context) {
	dataStore := c.MustGet("datastore").(sql.SQLDataStore)
	var newUser sql.AddUserRequest
	var addUserResult sql.RedactedUser
	if err := c.BindJSON(&newUser); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	addUserResult, err := dataStore.AddUser(sql.User{
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Notes:     newUser.Notes,
		Password:  newUser.Password,
	})

	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	fmt.Printf("Email: %v", addUserResult.Email)
	c.JSON(http.StatusOK, addUserResult)
	return

}

func DeleteUser(c *gin.Context) {
	dataStore := c.MustGet("datastore").(sql.SQLDataStore)
	var deleteUserRequest struct {
		Email string `form:"email" binding:"omitempty,email,min=0,max=50"`
		Id    string `form:"id" binding:"omitempty,min=0,max=50"`
	}
	fmt.Println("email: ", deleteUserRequest.Email, " id: ", deleteUserRequest.Id)

	if err := c.BindQuery(&deleteUserRequest); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Needed because you cannot make both fields optional on the deleteUserRequest
	if deleteUserRequest.Email == "" && deleteUserRequest.Id == "" {
		fmt.Println("[handlers.go] Could not delete user. Missing `email` and `id` params in request.")
		c.Status(http.StatusBadRequest)
		return
	}

	deleteUserResult := dataStore.DeleteUser(deleteUserRequest.Email, deleteUserRequest.Id)
	fmt.Println("deleteUserResult: ", deleteUserResult)
	if deleteUserResult {
		fmt.Println("[handlers.go] Successfully deleted user.")
		c.Status(http.StatusOK)
		return
	} else {
		c.Status(http.StatusInternalServerError)
		return
	}

}

func GetUser(c *gin.Context) {
	dataStore := c.MustGet("datastore").(sql.SQLDataStore)
	var getUserRequest struct {
		Email string `form:"email" binding:"omitempty,email,min=0,max=50"`
		Id    string `form:"id" binding:"omitempty,min=0,max=50"`
	}

	if err := c.BindQuery(&getUserRequest); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if getUserRequest.Email == "" && getUserRequest.Id == "" {
		fmt.Println("[handlers.go] Could not get user. Missing `email` and `id` params in request.")
		c.Status(http.StatusBadRequest)
		return
	}

	fmt.Println("email: ", getUserRequest.Email, " id: ", getUserRequest.Id)
	user := dataStore.GetUser(getUserRequest.Email, getUserRequest.Id)
	if user.Id == nil {
		fmt.Println(user)
		c.Status(http.StatusNotFound)
		return
	}
	redactedUser := sql.RedactedUser{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		Id:        user.Id,
	}
	c.JSON(http.StatusOK, redactedUser)
	return
}

func GetUsers(c *gin.Context) {

	var getUsersRequest struct {
		Offset *int16 `form:"offset" binding:"required,numeric,min=0,max=1000"`
	}

	if err := c.BindQuery(&getUsersRequest); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	dataStore := c.MustGet("datastore").(sql.SQLDataStore)

	users := dataStore.GetUsers(*getUsersRequest.Offset)
	fmt.Println(users)
	if len(users.Users) >= 1 {
		c.JSON(http.StatusOK, users)
		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

func UpdateUser(c *gin.Context) {
	dataStore := c.MustGet("datastore").(sql.SQLDataStore)
	var user sql.UpdateUserRequest
	if err := c.BindJSON(&user); err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	fmt.Printf("{\nEmail: %s\nId: %v\nNotes: %v\nFirstName: %s\nLastName: %s\n}\n",
		user.Email,
		user.Notes,
		user.Id,
		user.FirstName,
		user.LastName)

	updateUserResult := dataStore.UpdateUser(user)
	if updateUserResult {
		c.Status(200)
		return
	} else {
		c.Status(500)
		return
	}

}

func Search(c *gin.Context) {
	dataStore := c.MustGet("datastore").(sql.SQLDataStore)
	value := c.Query("value")
	fmt.Println("value:", value)
	result := dataStore.Search(value)
	if len(result) >= 1 {
		c.JSON(http.StatusOK, result)
		return
	}

	c.JSON(http.StatusOK, []int{})
	return
}
