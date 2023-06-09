package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

var app App

// Notes:
//   - TestMain is not a real test, it can be used to setup the tests and more... => therefore no tests found
//   - M is a type passed to a TestMain function to run the actual tests.
func TestMain(m *testing.M) {
	app.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := app.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	app.DB.Exec("DELETE FROM products")
	app.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Expected response Code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var jsonResponse map[string]string
	err := json.Unmarshal(response.Body.Bytes(), &jsonResponse)
	if err != nil {
		t.Error("Found error unmarshalling the response")
	}
	if jsonResponse["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", jsonResponse["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"name": "test product", "price": 11.23}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type:", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var jsonResponse map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &jsonResponse)
	if err != nil {
		t.Error("Found error unmarshalling the response")
	}

	if jsonResponse["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got %s", jsonResponse["name"])
	}

	if jsonResponse["price"] != 11.23 {
		t.Errorf("Expected product price to be '11.23'. Got %s", jsonResponse["price"])
	}

	if jsonResponse["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got %s", jsonResponse["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product -updated name", "price": 11.23}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var jsonResponse map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &jsonResponse)

	if jsonResponse["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], jsonResponse["id"])
	}
	if jsonResponse["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], jsonResponse["name"], jsonResponse["name"])
	}
	if jsonResponse["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], jsonResponse["price"], jsonResponse["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestFindProductsByName(t *testing.T) {
	clearTable()
	addProduct("red ball", 20)
	addProduct("ball", 25)
	addProduct("phone", 500)

	query := url.Values{}
	query.Set("name", "ball")
	requestUrl := "/product/search?" + query.Encode()
	req, _ := http.NewRequest("GET", requestUrl, nil)

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var products []product
	if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
		t.Error("Error parsing body!")
	}

	if len(products) != 2 {
		log.Fatalf("Expected to receive 2 products. Received %d", len(products))
	}

	checkProduct(t, products[0], product{ID: 1, Name: "red ball", Price: 20})
	checkProduct(t, products[1], product{ID: 2, Name: "ball", Price: 25})
}

func TestGetProducts(t *testing.T) {
	clearTable()
	addProduct("red ball", 20)
	addProduct("ball", 25)
	addProduct("phone", 500)

	query := url.Values{}
	query.Set("start", "0")
	query.Set("count", "2")
	requestUrl := "/products?" + query.Encode()
	req, _ := http.NewRequest("GET", requestUrl, nil)

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var products []product
	if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
		t.Error("Error parsing body!")
	}

	if len(products) != 2 {
		log.Fatalf("Expected to receive 2 products. Received %d", len(products))
	}

	checkProduct(t, products[0], product{ID: 1, Name: "red ball", Price: 20})
	checkProduct(t, products[1], product{ID: 2, Name: "ball", Price: 25})
}

func TestGetProductsSortedByName(t *testing.T) {
	clearTable()
	addProduct("red ball", 20)
	addProduct("ball", 25)
	addProduct("phone", 500)

	query := url.Values{}
	query.Set("sortProperty", "name")
	query.Set("sortDirection", "asc")
	requestUrl := "/products?" + query.Encode()
	req, _ := http.NewRequest("GET", requestUrl, nil)

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var products []product
	if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
		t.Error("Error parsing body!")
	}

	if len(products) != 3 {
		log.Fatalf("Expected to receive 3 products. Received %d", len(products))
	}

	checkProduct(t, products[0], product{ID: 2, Name: "ball", Price: 25})
	checkProduct(t, products[1], product{ID: 3, Name: "phone", Price: 500})
	checkProduct(t, products[2], product{ID: 1, Name: "red ball", Price: 20})
}

func TestGetProductsSortedByPrice(t *testing.T) {
	clearTable()
	addProduct("red ball", 20)
	addProduct("ball", 25)
	addProduct("phone", 500)

	query := url.Values{}
	query.Set("sortProperty", "price")
	query.Set("sortDirection", "desc")
	requestUrl := "/products?" + query.Encode()
	req, _ := http.NewRequest("GET", requestUrl, nil)

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var products []product
	if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
		t.Error("Error parsing body!")
	}

	if len(products) != 3 {
		log.Fatalf("Expected to receive 3 products. Received %d", len(products))
	}

	checkProduct(t, products[0], product{ID: 3, Name: "phone", Price: 500})
	checkProduct(t, products[1], product{ID: 2, Name: "ball", Price: 25})
	checkProduct(t, products[2], product{ID: 1, Name: "red ball", Price: 20})
}

func addProducts(count int) {
	for i := 0; i < count; i++ {
		// Note: strconv makes a string out of the int
		addProduct("Product"+strconv.Itoa(i), float64((i+1.0)*10))
	}
}

func addProduct(name string, price float64) {
	app.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", name, price)
}

func checkProduct(t *testing.T, actual product, expected product) {
	if actual.ID != expected.ID {
		t.Errorf("Expected the id %v. Got %v", expected.ID, actual.ID)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected the name %v. Got %v", expected.ID, actual.ID)
	}
	if actual.Price != expected.Price {
		t.Errorf("Expected the price %v. Got %v", expected.ID, actual.ID)
	}
}
