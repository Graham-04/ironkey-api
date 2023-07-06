package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Graham-04/ironkey-api/hash"

	"github.com/go-sql-driver/mysql"
)

var (
	once sync.Once
	// db   *sql.DB
)

type User struct {
	Email     string  `json:"email"`
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	Notes     *string `json:"notes"`
	Password  string  `json:"password"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
}

type RedactedUser struct {
	Email     string  `json:"email"`
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	Notes     *string `json:"notes"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
}

type GetUsersResult struct {
	Users []RedactedUser `json:"users"`
	Total int            `json:"total"`
}

type UpdateUserRequest struct {
	Id        *string `json:"id" binding:"required"`
	Email     string  `json:"email" binding:"required,email"`
	FirstName string  `json:"firstname" binding:"required,min=1,max=100"`
	LastName  string  `json:"lastname" binding:"required,min=1,max=100"`
	Notes     string  `json:"notes"`
}

type AddUserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	FirstName string  `json:"firstname" binding:"required,min=1,max=100"`
	LastName  string  `json:"lastname" binding:"required,min=1,max=100"`
	Notes     *string `json:"notes"`
	Password  string  `json:"password" binding:"required"`
}

type SQLDataStore interface {
	AddUser(user User) (RedactedUser, error)
	InitDB()
	// GetDB() *sql.DB
	UserExists(email string, id string) bool
	GetUser(email string, id string) User
	GetUsers(offset int16) GetUsersResult
	DeleteUser(email string, id string) bool
	GetTotalUserCount() int
	Search(value string) []RedactedUser
	UpdateUser(newUserData UpdateUserRequest) bool
	// UpdateUser(user User) bool
}

func GetDataStore(databaseType string) SQLDataStore {
	switch databaseType {
	case "mysql":
		return &MySQLDataStore{}
	default:
		log.Fatal("No database type specified")
		return nil
	}
}

type MySQLDataStore struct {
	db *sql.DB
}

func (m *MySQLDataStore) InitDB() {
	once.Do(func() {
		var err error
		m.db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/UserData")
		if err != nil {
			log.Fatal("Could not connect to DB. Exiting...")
		}

		err = m.db.Ping()
		if err != nil {
			log.Fatal("[sql.go] Could not ping DB. Exiting...")
		}

		m.db.SetConnMaxIdleTime(time.Minute * 5)
		m.db.SetMaxOpenConns(100)
		m.db.SetMaxIdleConns(100)
		fmt.Println("[sql.go] Connected To MySQL. Connection pool established.")
	})
}

// func (m *MySQLDataStore) GetDB() *sql.DB {
// 	return m.db
// }

func (m *MySQLDataStore) AddUser(user User) (RedactedUser, error) {

	userExists := m.UserExists(user.Email, "")
	if !userExists {
		fmt.Printf("[sql.go] Email %v does not exist. Adding user...\n", user.Email)
	} else {
		fmt.Printf("[sql.go] Email %v already exists\n", user.Email)
		return RedactedUser{}, errors.New("Email already exists")
	}

	start := time.Now()
	stmt, err := m.db.Prepare(`INSERT INTO Users (email, passwordHash, firstName, lastName, notes) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatal("[sql.go] Could not prepare AddUser SQL statement. Exiting...")
	}
	defer stmt.Close()

	user.Password = hash.GeneratePasswordHash(user.Password)

	result, err := stmt.Exec(user.Email, user.Password, user.FirstName, user.LastName, user.Notes)
	if err != nil {
		// log.Fatal("[sql.go] Could not execute insert user query. Exiting...")
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return RedactedUser{}, fmt.Errorf("Attempted to insert duplicate email %v", user.Email)

		} else {
			elapsed := time.Since(start)
			fmt.Printf("[sql.go] Error inserting email: %v (Duration: %v)\n", user.Email, elapsed)
			return RedactedUser{}, err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("[sql.go] Could not get RowsAffected for AddUser. Exiting...")
	} else {
		if rowsAffected > 0 {
			u := m.GetUser(user.Email, "")
			rd := RedactedUser{
				Email: u.Email,
				FirstName: u.FirstName,
				LastName: u.LastName,
				Id: u.Id,
				Notes: u.Notes,
				CreatedAt: u.CreatedAt,
			}
			return rd, nil
		} else {
			return RedactedUser{}, err
		}

	}

	return RedactedUser{}, err
}

func (m *MySQLDataStore) DeleteUser(email string, id string) bool {
	userExists := m.UserExists(email, id)
	if !userExists {
		fmt.Printf("[sql.go] User not found. Email: %v. ID: %v\n", email, id)
		return false
	}

	stmt, err := m.db.Prepare("DELETE FROM Users WHERE email = ? OR ID = ?")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare DeleteUser query. Exiting...")
	}
	defer stmt.Close()

	result, err := stmt.Exec(email, id)
	if err != nil {
		log.Println("[sql.go] Could not delete user. Email: %v. ID: %v\n", email, id)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("[sql.go] Could not get RowsAffected(). Exiting...")

	} else {
		fmt.Println("rowsAffected", rowsAffected)
		if rowsAffected == 1 {
			fmt.Printf("[sql.go] Deleted user. Email: %v. ID: %v\n", email, id)
			return true
		} else if rowsAffected == 0 {
			fmt.Printf("[sql.go] User not found. Email: %v. ID: %v\n", email, id)
			return false
		} else {
			return false
		}
	}

	return false
}

func (m *MySQLDataStore) UserExists(email string, id string) bool {
	var exists bool = false
	stmt, err := m.db.Prepare("SELECT EXISTS (SELECT 1 FROM Users WHERE email = ? OR id = ?)")
	if err != nil {
		log.Fatal("[sq.go] Coudl not prepare UserExists query. Exiting...")
	}
	defer stmt.Close()

	if err != nil {
		log.Fatal("[sql.go] Could not prepare UserExists query. Exiting...")
	}

	row := stmt.QueryRow(email, id)

	err = row.Scan(&exists)
	if err != nil {
		log.Fatal("[sql.go] Could not scan UserExists query. Exiting...")
	}

	fmt.Println("exists: ", exists)
	return exists

}

func (m *MySQLDataStore) GetUser(email string, id string) User {
	stmt, err := m.db.Prepare("SELECT * FROM Users WHERE email = ? OR id = ?")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare GetUser query. Exiting...")
	}
	defer stmt.Close()
	var user User

	err = stmt.QueryRow(email, id).Scan(&user.Email, &user.Id, &user.Password, &user.FirstName, &user.LastName, &user.Notes, &user.CreatedAt)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal("[sql.go] Could not scan GetUser query. Exiting...", err)
	}
	// err = row.Scan(&user.Email, &user.Id, &user.Password, &user.FirstName, &user.LastName, &user.CreatedAt)
	// if err != nil && err != sql.ErrNoRows {
	// 	log.Fatal("[sql.go] Could not scan GetUser query. Exiting...")
	// }
	return user
}

func (m *MySQLDataStore) GetTotalUserCount() int {
	var count int
	stmt, err := m.db.Prepare("SELECT COUNT(id) from Users")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare GetTotalUserCount query. Exiting...")
	}
	defer stmt.Close()
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		log.Fatal("[sql.go] Could not scan GetTotalUserCount. Exiting...")
	}
	return count
}

func (m *MySQLDataStore) GetUsers(offset int16) GetUsersResult {
	var result GetUsersResult
	stmt, err := m.db.Prepare("SELECT * FROM Users LIMIT 10 OFFSET ?")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare GetUsers query. Exiting...")
	}
	defer stmt.Close()

	rows, err := stmt.Query(offset)
	if err != nil {
		log.Fatal("[sql.go] Could not query rows for GetUsers. Exiting...")
	}

	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Email, &user.Id, &user.Password, &user.FirstName, &user.LastName, &user.Notes, &user.CreatedAt)
		if err != nil {
			fmt.Println(err)
			log.Fatal("[sql.go] Could not scan GetUsers rows. Exiting...")
		}
		result.Users = append(result.Users, RedactedUser{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Notes:     user.Notes,
			CreatedAt: user.CreatedAt,
			Id:        user.Id,
		})
	}

	result.Total = m.GetTotalUserCount()
	return result
}

func (m *MySQLDataStore) UpdateUser(user UpdateUserRequest) bool {
	stmt, err := m.db.Prepare("UPDATE Users SET firstName = ?, lastName = ?, notes = ?, email = ? WHERE id = ?")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare UpdateUser query. Exiting...")
	}

	defer stmt.Close()

	row, err := stmt.Exec(user.FirstName, user.LastName, user.Notes, user.Email, user.Id)
	if err != nil {
		log.Fatal("[sql.go] Could not execute UpdateUser query. Exiting...")
	}

	affected, err := row.RowsAffected()
	if err != nil {
		fmt.Println("[sql.go] Could not get RowsAffected for UpdateUser. Exiting...")
	}

	fmt.Println(affected)
	if affected == 1 {
		return true
	} else {
		return false
	}

}

func (m *MySQLDataStore) Search(value string) []RedactedUser {
	var result []RedactedUser
	stmt, err := m.db.Prepare("SELECT email, id, firstName, lastName, notes, createdAt FROM Users WHERE email LIKE ? OR id LIKE ? OR passwordHash LIKE ? OR firstName LIKE ? OR lastName LIKE ? OR CAST(createdAt AS CHAR) LIKE ?")
	if err != nil {
		log.Fatal("[sql.go] Could not prepare Search query. Exiting", err)
	}

	defer stmt.Close()

	value = "%" + value + "%"
	fmt.Println(value)

	rows, err := stmt.Query(value, value, value, value, value, value)
	if err != nil {
		log.Fatal("[sql.go] Could not scan Search query. Exiting...", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user RedactedUser
		err = rows.Scan(&user.Email, &user.Id, &user.FirstName, &user.LastName, &user.Notes, &user.CreatedAt)
		if err != nil {
			log.Fatal("[sql.go] Could not scan Search query. Exiting...", err)
		}
		result = append(result, user)
	}

	fmt.Println("result:", len(result))
	return result
}
