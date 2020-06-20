package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	if err != nil {
		log.Fatal("connect database error", err)
	}

	createTb := `CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);`

	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table customers", err)
	}
	fmt.Println("create table success")
}

func createCustomersHandler(c *gin.Context) {
	cu := Customer{}
	if err := c.ShouldBindJSON(&cu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3)  RETURNING id", cu.Name, cu.Email, cu.Status)

	err := row.Scan(&cu.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, cu)
}

func getCustomersHandler(c *gin.Context) {
	status := c.Query("status")

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	customers := []*Customer{}
	for rows.Next() {
		cu := &Customer{}

		err := rows.Scan(&cu.ID, &cu.Name, &cu.Email, &cu.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		customers = append(customers, cu)
	}

	cc := []*Customer{}

	for _, item := range customers {
		if status != "" {
			if item.Status == status {
				cc = append(cc, item)
			}
		} else {
			cc = append(cc, item)
		}
	}

	c.JSON(http.StatusOK, cc)
}

func getCustomerByIdHandler(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	row := stmt.QueryRow(id)
	cu := &Customer{}

	err = row.Scan(&cu.ID, &cu.Name, &cu.Email, &cu.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, cu)
}

func updateCustomersHandler(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	row := stmt.QueryRow(id)

	cu := &Customer{}

	err = row.Scan(&cu.ID, &cu.Name, &cu.Email, &cu.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := c.ShouldBindJSON(cu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err = db.Prepare("UPDATE customers SET status=$2, email=$3, name=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if _, err := stmt.Exec(id, cu.Status, cu.Email, cu.Name); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, cu)
}

func deleteCustomersHandler(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("DELETE FROM customers WHERE id = $1")
	if err != nil {
		log.Fatal("can't prepare delete statement", err)
	}

	if _, err := stmt.Exec(id); err != nil {
		log.Fatal("can,t execute delete statement", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func main() {
	r := gin.Default()
	r.POST("/customers", createCustomersHandler)
	r.GET("/customers/:id", getCustomerByIdHandler)
	r.GET("/customers", getCustomersHandler)
	r.PUT("/customers/:id", updateCustomersHandler)
	r.DELETE("/customers/:id", deleteCustomersHandler)
	r.Run(":2019")
}
