package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/lattots/enrich/pkg/catalog"
	"github.com/lattots/enrich/pkg/config"
)

const configFilepath = "./data/config.json"

func main() {
	conf, err := config.Load(configFilepath)
	if err != nil {
		log.Fatalln("error loading config:", err)
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "local" { // If environment is local, secrets are loaded to environment variables
		err = godotenv.Load(conf.Filepaths.SecretsFilepath)
		if err != nil {
			log.Fatalln("error loading .env file:", err)
		}
	}

	cat, err := catalog.New(conf)
	if err != nil {
		log.Fatalln("error creating catalog:", err)
	}

	err = cat.Load(0)
	if err != nil {
		log.Fatalln("error loading catalog:", err)
	}

	err = cat.Cluster(conf.Clusters)
	if err != nil {
		log.Fatalln("error clustering catalog:", err)
	}

	err = cat.UpdateProducts()
	if err != nil {
		log.Fatalln("error updating products:", err)
	}
}
