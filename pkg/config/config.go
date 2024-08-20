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
	Filepaths Filepaths     `json:"filepaths"`
	DB        Database      `json:"db"`             // All database settings and configurations
	LM        LanguageModel `json:"language-model"` // Language model settings
}

// Filepaths is a struct for holding file paths used in database initialization.
type Filepaths struct {
	CatalogFilepath string `json:"catalog-filepath"`
	PromptsFilepath string `json:"prompts-filepath"`
	SecretsFilepath string `json:"secrets-filepath"`
}

type Database struct {
	ProductTable string      `json:"product-table"`
	ColumnNames  ColumnNames `json:"column-names"`
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
	Attributes       string `json:"attributes"`
}

// LanguageModel is a struct for holding language model settings
type LanguageModel struct {
	ActiveProvider string          `json:"active-provider"`
	Providers      ProviderOptions `json:"provider-options"`
}

// ProviderOptions is a library of options for different language model providers
type ProviderOptions struct {
	Gemini ProviderOption `json:"gemini"`
	OpenAI ProviderOption `json:"openai"`
}

// ProviderOption contains options for given language model provider
type ProviderOption struct {
	GenerativeModel string `json:"generative-model"`
	EmbeddingsModel string `json:"embeddings-model"`
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
