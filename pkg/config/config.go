package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config is a struct for holding configuration options.
// Configurations are read from JSON at runtime, so they can be changed without recompiling the entire program.
type Config struct {
	Filepaths Filepaths `json:"filepaths"`
	DB        Database  `json:"db"`       // All database settings and configurations
	OpenAI    OpenAI    `json:"open-ai"`  // OpenAI settings
	Clusters  Clusters  `json:"clusters"` // Clustering options
}

// Filepaths is a struct for holding file paths used in database initialization.
type Filepaths struct {
	CatalogFilepath string `json:"catalog-filepath"`
	PromptsFilepath string `json:"prompts-filepath"`
	SecretsFilepath string `json:"secrets-filepath"`
}

type Database struct {
	ProductTable     string      `json:"product-table"`
	ColumnNames      ColumnNames `json:"column-names"`
	VectorDimensions int         `json:"vector-dimensions"`
}

// ColumnNames is a struct for holding column names used database table.
type ColumnNames struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Price            string `json:"price"`
	Link             string `json:"link"`
	Image            string `json:"image"`
	StyleText        string `json:"style-text"`
	UseCaseText      string `json:"use-case-text"`
	StyleEmbedding   string `json:"style-embedding"`
	UseCaseEmbedding string `json:"use-case-embedding"`
	StyleCluster     string `json:"style-cluster"`
	UseCaseCluster   string `json:"use-case-cluster"`
	Attributes       string `json:"attributes"`
}

// OpenAI is a struct for holding OpenAI settings.
type OpenAI struct {
	GPTModel        string `json:"gpt-model"`
	EmbeddingsModel string `json:"embeddings-model"`
}

// Clusters is a struct for holding clustering options related to clustering
type Clusters struct {
	Algorithm string        `json:"algorithm"`
	KMeans    KMeansOptions `json:"k-means"`
	DBSCAN    DBSCANOptions `json:"dbscan"`
}

type KMeansOptions struct {
	StyleK   int `json:"style-k"`    // k used for clustering products based on style
	UseCaseK int `json:"use-case-k"` // k used for clustering products based on use case
}

type DBSCANOptions struct {
	MinPoints int     `json:"min-points"` // Minimum number of points to consider point a central point
	Epsilon   float64 `json:"epsilon"`    // Radius of area to check for other points
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
