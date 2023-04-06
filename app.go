package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

const maxPageSize = 10

func (app *App) Initialize(user string, password string, dbname string) {
	connectionString := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, dbname)

	var err error
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	app.Router = mux.NewRouter()
}
func (app *App) Run(addr string) {}

func (app *App) initializeRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")
	app.Router.HandleFunc("/product", app.createProduct).Methods("POST")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.getProduct).Methods("GET")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.updateProduct).Methods("PUT")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.deleteProduct).Methods("DELETE")
}
func (app *App) getProduct(responseWriter http.ResponseWriter, request *http.Request) {
	requestVariables := mux.Vars(request)
	id, err := strconv.Atoi(requestVariables["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product := product{ID: id}
	if err := product.getProduct(app.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(responseWriter, http.StatusNotFound, "Product not found")
		default:
			respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(responseWriter, http.StatusOK, product)
}

func (app *App) getProducts(responseWriter http.ResponseWriter, request *http.Request) {
	count, _ := strconv.Atoi(request.FormValue("count"))
	start, _ := strconv.Atoi(request.FormValue("start"))

	if count > maxPageSize || count < 1 {
		count = maxPageSize
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(app.DB, start, count)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, products)
}

func (app *App) createProduct(responseWriter http.ResponseWriter, request *http.Request) {
	var product product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&product); err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer request.Body.Close()

	if err := product.createProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusCreated, product)
}

func (app *App) updateProduct(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&product); err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer request.Body.Close()
	product.ID = id

	if err := product.updateProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, product)
}

func (app *App) deleteProduct(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	product := product{ID: id}
	if err := product.deleteProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(responseWriter, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithJSON(responseWriter http.ResponseWriter, statusCode int, payload interface{}) {
	response, _ := json.Marshal(payload)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	responseWriter.Write(response)
}

func respondWithError(responseWriter http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(responseWriter, statusCode, map[string]string{"error": message})
}
