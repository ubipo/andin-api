package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/graphql-go/handler"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq" // Postgres driver
)

// Serve serves the andin api over http
func Serve() {
	dbpass, exists := os.LookupEnv("DB_PASS")
	if !exists {
		panic(fmt.Errorf("must set DB_PASS enviroment variable"))
	}
	connStr := fmt.Sprintf("user=andin_migrate password=%s dbname=andin_dev sslmode=disable", dbpass)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	schema := generateSchema(db)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8980", nil)
}
