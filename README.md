# Status
[![build](https://github.com/benjaminellmer/CD_Exercise_02/actions/workflows/go.yml/badge.svg)](https://github.com/benjaminellmer/CD_Exercise_02/actions/workflows/go.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=benjaminellmer_CD_Exercise_02&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=benjaminellmer_CD_Exercise_02)

# Steps
## Postgres Setup
Create postgres container using docker:
```bash
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres 
```

Open postgres shell in container and create Table
```bash
docker exec -it postgres psql -U postgres
CREATE TABLE products(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
);
```

## Setup Github Repo
![create-repo](./docs/images/create-repo.png)

Clone repo:
```bash
git clone git@github.com:benjaminellmer/CD_Exercise_02.git
cd CD_Exercise_02
```

## Setup Go module
```bash
go mod init github.com/benjaminellmer/CD_Exercise_02.git
go get -u github.com/gorilla/mux 
go get -u github.com/lib/pq
```

## Issues that I faced
### Test packages
In the tutorial it is described that the file `main_test.go` should be inside the package `main_test`.  
According to some [research](https://stackoverflow.com/questions/19998250/proper-package-naming-for-testing-with-the-go-language) that I did, it depends on the testing strategy, whether to use a different package for the tests.  
Still the tests did not work for me if I used the package `main_test` and since Mr. Kurz also did use the package `main` in his repo I thought it is fine.

### Postgres Connection String
I had to use this connection String, because the one from the tutorial did not work for me:
```go
connectionString := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, dbname)
```

### Array Initialization
IntelliJ told me that I should use `var products []product` instead of `products := []product{}` to initialize the array in the `getProducts`function.  
But this returns `nil` when no products are added, but we want to get an empty array as response.

## My learnings (cheatsheet)
### Json and structs
It is very simple to parse json using go. We can simply define the "json" tags.
```go
type product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}
```
Find more detailed information [here](https://drstearns.github.io/tutorials/gojson/).

### Constructors
In go constructors have `{}` instead of `()`

### Tests
- go automatically detects which files end with `_test.go` and those files get executed as tests.
- `testing.M` is a type passed to a TestMain function to run the actual tests.
- TestMain is not a real test, it can be used to setup the tests and more...  
Therefore it is okay to have the output `no tests found...` as long as only the TestMain exists.  
Find more detailed information [here](http://cs-guy.com/blog/2015/01/test-main/).

### If with a short statement
In go it is possible to have a statement in front of an if.  
We are using this to assign variables, that should only live inside the if block:
```go
if body := response.Body.String(); body != "[]" {
    t.Errorf("Expected an empty array. Got %s", body)
}
```
Find more detailed information [here](https://go.dev/tour/flowcontrol/6).

### Maps
A map that has string keys and string values is written like this:
`var jsonResponse map[string]string` 
but a map that has a string key and any object as value (like in json) is declared like this:
`var jsonResponse map[string]interface{}`

### Defer
Defer is some fancy idiom that we can use in go if we want an action to be executed when the function returns.  
It allows us e.g. to (defer) close some stream even before we use it, and then the close is executed as soon as we return.  
Using this idiom we do not have to write the close statement for each separate return path.  
It reminds me a bit on the `finally` statement of a try/catch.  
Find more detailed information [here](https://go.dev/tour/flowcontrol/12).

