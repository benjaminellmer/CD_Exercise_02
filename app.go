package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

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
