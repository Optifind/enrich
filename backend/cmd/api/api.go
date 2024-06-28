package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/handler"
)

const configFilepath = "./data/config.json"

func main() {
	//
	conf, err := config.Load(configFilepath)
	if err != nil {
		log.Fatalln("error loading config: ", err)
	}

	err = godotenv.Load(conf.API.SecretsFilepath)
	if err != nil {
		log.Fatalln("error loading .env file: ", err)
	}

	// New handler object is created. This is used to store configurations and Milvus database handle.
	h, err := handler.New(conf.API)
	if err != nil {
		log.Fatalln("error creating handler: ", err)
	}

	// New HTTP router is created.
	router := http.NewServeMux()

	// Routes
	router.HandleFunc("GET /product/{id}", h.HandleGetProduct)

	port := ":3000"
	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln("unexpected error: ", err)
	}
}
