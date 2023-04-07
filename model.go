package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (p *product) getProduct(db *sql.DB) error {
	// Scan copies the values of the first matched element to the given target
	return db.QueryRow("SELECT name,price FROM products WHERE id=$1", p.ID).Scan(&p.Name, &p.Price)
}

func (p *product) updateProduct(db *sql.DB) error {
	_, err := db.Exec("UPDATE products SET name=$1, price=$2 WHERE id=$3", p.Name, p.Price, p.ID)

	return err
}

func (p *product) deleteProduct(db *sql.DB) error {
	_, err := db.Exec("DELETE from products WHERE id=$1", p.ID)

	return err
}

func (p *product) createProduct(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO products(name, price) VALUES($1, $2) RETURNING id", p.Name, p.Price).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getProducts(db *sql.DB, start int, count int, sortProperty string, sortDirection string) ([]product, error) {
	var query string
	if sortProperty == "" && (strings.ToLower(sortDirection) != "asc" || strings.ToLower(sortDirection) != "desc") {
		query = "SELECT id, name, price FROM products LIMIT $1 OFFSET $2"
	} else {
		query = fmt.Sprintf("SELECT id, name, price FROM products ORDER BY %s %s LIMIT $1 OFFSET $2", sortProperty, sortDirection)
	}
	rows, err := db.Query(query, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	products := []product{}

	for rows.Next() {
		var product product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			return nil, err
		}

		products = append(products, product)
	}
	return products, nil
}

func findProductByName(db *sql.DB, searchTerm string) ([]product, error) {
	rows, err := db.Query("SELECT id, name, price FROM products")

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	products := []product{}

	regex := regexp.MustCompile(searchTerm)

	for rows.Next() {
		var product product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			return nil, err
		}

		if regex.MatchString(strings.ToLower(product.Name)) {
			products = append(products, product)
		}
	}
	return products, nil
}
