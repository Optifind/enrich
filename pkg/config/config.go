package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config is a struct for holding configuration options.
// Configurations are read from JSON at runtime, so they can be changed without compiling the entire program.
type Config struct {
	Init Init `json:"init"`
	API  API  `json:"api"`
}

// Init is a struct for holding configuration options related to database initialization.
type Init struct {
	CatalogFilepath string `json:"catalog-filepath"`
	PromptsFilepath string `json:"prompts-filepath"`

	MilvusAddress     string      `json:"milvus-address"`
	MilvusCollection  string      `json:"milvus-collection"`
	MilvusColumnNames ColumnNames `json:"milvus-column-names"`

	SecretsFilepath string `json:"secrets-filepath"`

	GPTModel        string `json:"gpt-model"`
	EmbeddingsModel string `json:"embeddings-model"`
}

type API struct {
	MilvusAddress     string      `json:"milvus-address"`
	MilvusCollection  string      `json:"milvus-collection"`
	MilvusColumnNames ColumnNames `json:"milvus-column-names"`

	SecretsFilepath string `json:"secrets-filepath"`
}

// ColumnNames is a struct for holding column names used in database initialization.
type ColumnNames struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Price          string `json:"price"`
	Link           string `json:"link"`
	Image          string `json:"image"`
	StyleText      string `json:"style-text"`
	UseCaseText    string `json:"use-case-text"`
	StyleEmbedding string `json:"style-embedding"`
	Attributes     string `json:"attributes"`
}

// Load loads the configuration object from JSON. Returns Config object and an error.
func Load(filepath string) (Config, error) {
	fmt.Println("Loading config file...")
	// Config JSON file is opened to read.
	file, err := os.Open(filepath)
	if err != nil {
		return Config{}, err
	}

	// JSON files content is read to byte code.
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	// Config json is read unmarshalled to config struct.
	var config Config
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return Config{}, err
	}

	// Config and nil error is returned.
	return config, nil
}
