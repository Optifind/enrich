package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lattots/enrich/pkg/catalog"
	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/prompts"
)

const configFilepath = "./data/config.json"

func main() {
	conf, err := config.Load(configFilepath)
	if err != nil {
		log.Fatalln("error loading config file:", err)
	}

	err = godotenv.Load(conf.Init.SecretsFilepath)
	if err != nil {
		log.Fatalln("error loading secrets to environment variables:", err)
	}

	cat, err := catalog.New(conf.Init)
	if err != nil {
		log.Fatalln("error creating catalog object:", err)
	}

	err = cat.CreateProducts()
	if err != nil {
		log.Fatalln("error creating product objects from CSV file:", err)
	}

	prmts, err := prompts.Load(conf.Init.PromptsFilepath)
	if err != nil {
		log.Fatalln("error reading prompts from file:", err)
	}

	err = cat.ProcessProducts(conf.Init, prmts)
	if err != nil {
		log.Fatalln("error processing products:", err)
	}

	err = cat.InitCollection()
	if err != nil {
		log.Fatalln("error initializing Milvus database collection:", err)
	}

	err = cat.InsertToMilvus(conf.Init)
	if err != nil {
		log.Fatalln("error inserting catalog to Milvus:", err)
	}

	err = cat.CreateIndex(conf.Init)
	if err != nil {
		log.Fatalln("error creating index for Milvus collection:", err)
	}
}
