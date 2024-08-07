package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

	env := os.Getenv("ENVIRONMENT")
	if env == "local" { // If environment is local, secrets are loaded to environment variables
		fmt.Println("Running in local mode")
		err := godotenv.Load(conf.Filepaths.SecretsFilepath)
		if err != nil {
			log.Fatalln("error loading .env file:", err)
		}
	}

	// New handler object is created. This is used to store configurations and Milvus database handle.
	h, err := handler.New(conf)
	if err != nil {
		log.Fatalln("error creating handler: ", err)
	}

	// New HTTP router is created.
	router := http.NewServeMux()

	// Routes.
	router.HandleFunc("/product/{id}", h.HandleGetProduct)
	router.HandleFunc("/products", h.HandleGetProducts)

	// Wrap the router with CORS middleware
	corsHandler := corsMiddleware(router)

	// Get PORT from Heroku env
	port := ":" + os.Getenv("PORT")
	fmt.Printf("Server started on port %s\n", port)
	if err := http.ListenAndServe(port, corsHandler); err != nil {
		log.Fatalln("unexpected error: ", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If it's a preflight request, stop here
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve the request to the next middleware/handler
		next.ServeHTTP(w, r)
	})
}
