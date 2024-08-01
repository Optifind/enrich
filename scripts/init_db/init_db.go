package main

import (
	"log"

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

	cat, err := catalog.New(conf)
	if err != nil {
		log.Fatalln("error creating catalog object:", err)
	}

	err = cat.CreateProducts()
	if err != nil {
		log.Fatalln("error creating product objects from CSV file:", err)
	}

	prmts, err := prompts.Load(conf.Filepaths.PromptsFilepath)
	if err != nil {
		log.Fatalln("error reading prompts from file:", err)
	}

	err = cat.ProcessProducts(conf.OpenAI, prmts)
	if err != nil {
		log.Fatalln("error processing products:", err)
	}

	err = cat.InitDatabase()
	if err != nil {
		log.Fatalln("error initializing database:", err)
	}

	err = cat.InsertToDB()
	if err != nil {
		log.Fatalln("error inserting catalog to database:", err)
	}

	err = cat.UpdateIndex()
	if err != nil {
		log.Fatalln("error creating index for Milvus collection:", err)
	}
}
