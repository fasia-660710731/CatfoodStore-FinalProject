package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Weight      string  `json:"weight"`
	AgeGroup    string  `json:"age_group"`
	BreedType   string  `json:"breed_type"` // เก็บเป็นสตริงของ array เช่น {"all"}
	Category    string  `json:"category"`
	ImageURL    string  `json:"image_url"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

var db *sql.DB

func main() {
	// โหลด env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("open db error:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("ping db error:", err)
	}

	fmt.Println("✅ Connected to PostgreSQL")

	r := gin.Default()

	// test db
	r.GET("/api/test-db", func(c *gin.Context) {
		var now string
		if err := db.QueryRow("SELECT NOW()").Scan(&now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"db": "connected", "time": now})
	})

	// CRUD
	r.GET("/api/products", getAllProducts)
	r.GET("/api/products/:id", getProductByID)
	r.POST("/api/products", createProduct)
	r.PUT("/api/products/:id", updateProduct)
	r.DELETE("/api/products/:id", deleteProduct)

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}

	r.Run(":" + portEnv)
}

func getAllProducts(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM products ORDER BY id ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := []Product{}

	for rows.Next() {
		var p Product
		// ตามลำดับคอลัมน์ใน init.sql
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Weight,
			&p.AgeGroup,
			&p.BreedType, // varchar[]
			&p.Category,
			&p.ImageURL,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, products)
}

func getProductByID(c *gin.Context) {
	id := c.Param("id")

	var p Product
	err := db.QueryRow("SELECT * FROM products WHERE id = $1", id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Weight,
		&p.AgeGroup,
		&p.BreedType,
		&p.Category,
		&p.ImageURL,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func createProduct(c *gin.Context) {
	var body Product
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(
		`INSERT INTO products
		(name, description, price, weight, age_group, breed_type, category, image_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		body.Name,
		body.Description,
		body.Price,
		body.Weight,
		body.AgeGroup,
		body.BreedType,
		body.Category,
		body.ImageURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created"})
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")

	var body Product
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec(
		`UPDATE products SET
		 name=$1, description=$2, price=$3, weight=$4,
		 age_group=$5, breed_type=$6, category=$7, image_url=$8,
		 updated_at = NOW()
		 WHERE id = $9`,
		body.Name,
		body.Description,
		body.Price,
		body.Weight,
		body.AgeGroup,
		body.BreedType,
		body.Category,
		body.ImageURL,
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func deleteProduct(c *gin.Context) {
	id := c.Param("id")

	res, err := db.Exec("DELETE FROM products WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
