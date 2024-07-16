package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/handler"
)

const configFilepath = "./data/config.json"

func main() {
	// Configuration file is loaded.
	conf, err := config.Load(configFilepath)
	if err != nil {
		log.Fatalln("error loading config: ", err)
	}

	// New handler object is created. This is used to store configurations and Milvus database handle.
	h, err := handler.New(conf)
	if err != nil {
		log.Fatalln("error creating handler: ", err)
	}

	// New HTTP router is created.
	router := http.NewServeMux()

	// Routes.
	router.HandleFunc("GET /product/{id}", h.HandleGetProduct)
	router.HandleFunc("GET /products", h.HandleGetProducts)

	// Get PORT from Heroku env
	port := os.Getenv("PORT")
	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln("unexpected error: ", err)
	}
}
