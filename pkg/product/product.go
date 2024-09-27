package product

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pgvector/pgvector-go"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/langmod"
	"github.com/lattots/enrich/pkg/prompts"
)

// Product is a struct that represents a single product in a product catalog.
type Product struct {
	ID          string  `json:"id"` // Product ID created and used by FDS
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`

	Link  string `json:"product-url"`   // URL of product page
	Image string `json:"product-image"` // URL of product image

	StyleText   string `json:"style-text"`    // Description of products design style
	UseCaseText string `json:"use-case-text"` // Description of products use case

	StyleEmbedding   []float32 `json:"style-embedding"`    // Embedding of product's design style
	UseCaseEmbedding []float32 `json:"use-case-embedding"` // Embedding of product's use case

	StyleCluster   int `json:"style-cluster"`    // Cluster number of products design style
	UseCaseCluster int `json:"use-case-cluster"` // Cluster number of products use case

	Attributes Attributes `json:"attributes"` // Attributes used to classify the product
}

// Attributes is a struct for holding product attribute data.
// TODO: This feature is currently work in progress and will not be created when the product is processed.
type Attributes struct {
	Material []string `json:"material"`
	Size     string   `json:"size"`
	Category []string `json:"category"`
}

// FromRow converts a single row from product catalog CSV to Product object.
// Returns pointer to Product object and an error.
func FromRow(row []string) (*Product, error) {
	// Six is the expected number of columns in the input CSV.
	// This number will change depending on the product catalog.
	const expectedColumns = 6
	if len(row) != expectedColumns {
		return nil, fmt.Errorf("error converting row with length %d to product", len(row))
	}

	price, err := extractPrice(row[3])
	if err != nil {
		return nil, fmt.Errorf("error extracting price from row: %w", err)
	}

	// Pointer to Product object is created with row's data.
	p := &Product{
		ID:          row[0],
		Title:       row[1],
		Description: row[2],
		Price:       price,
		Link:        row[4],
		Image:       row[5],
	}
	// Pointer to Product object is returned.
	return p, nil
}

// GetById fetches a product corresponding to ID from database. Returns a pointer to product and error.
func GetById(db *sql.DB, config config.Config, id string) (*Product, error) {
	cn := &config.DB.ColumnNames // Makes the code more concise
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE id = $1;",
		cn.ID,
		cn.Title,
		cn.Description,
		cn.Price,
		cn.Link,
		cn.Image,
		cn.StyleText,
		cn.UseCaseText,
		cn.StyleEmbedding,
		cn.UseCaseEmbedding,
		config.DB.ProductTable,
	)

	res := db.QueryRow(query, id)
	if err := res.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Product not found
		}
		return nil, fmt.Errorf("error fetching product from database: %w", err)
	}

	product, err := FromSQLRow(res)
	if err != nil {
		return nil, fmt.Errorf("error parsing product info: %w", err)
	}

	// Pointer to product is returned.
	return product, nil
}

// Process processes single product. This means creating style and
// use case texts and creating embeddings for them.
// Returns an error.
func (p *Product) Process(model langmod.LangMod, prompts prompts.Prompts) error {
	productInfo := fmt.Sprintf("Title: %s\nDescription: %s", p.Title, p.Description)

	styleText, err := model.CreateChatResponse(
		langmod.NewTextInput(prompts.StyleText.Content, prompts.StyleText.Role),
		langmod.NewImageInput(p.Image, "user"),
		langmod.NewTextInput(productInfo, "user"),
	)
	if err != nil {
		return fmt.Errorf("error creating style text: %w", err)
	}
	p.StyleText = styleText

	useCaseText, err := model.CreateChatResponse(
		langmod.NewTextInput(prompts.UseCaseText.Content, prompts.UseCaseText.Role),
		langmod.NewImageInput(p.Image, "user"),
		langmod.NewTextInput(productInfo, "user"),
	)
	if err != nil {
		return fmt.Errorf("error creating use case text: %w", err)
	}
	p.UseCaseText = useCaseText

	// Embeddings are created and saved to product's embeddings fields depending on which embeddings are specified.
	// See CreateEmbeddings for additional information.
	err = p.CreateEmbeddings(model)
	return err
}

// CreateEmbeddings creates embeddings for style text and use case text of the product.
// Created embeddings are saved to Product's fields StyleEmbedding and UseCaseEmbedding.
// Returns an error.
func (p *Product) CreateEmbeddings(model langmod.LangMod) error {
	embeddings, err := model.CreateEmbeddings(
		langmod.NewTextInput(p.StyleText, ""),
		langmod.NewTextInput(p.UseCaseText, ""),
	)
	if err != nil {
		return fmt.Errorf("error creating embeddings: %w", err)
	}

	p.StyleEmbedding = embeddings[0]
	p.UseCaseEmbedding = embeddings[1]

	return nil
}

// CreateAttributes TODO: Method should create product attributes with GPT and parse them to Attributes object.
func (p *Product) CreateAttributes(model langmod.LangMod, prompt prompts.Prompt) error {
	return errors.New("NOT IMPLEMENTED")
}

// Insert inserts the product to database
func (p *Product) Insert(db *sql.DB, tableName string, cn config.ColumnNames) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	values := []any{
		p.ID,
		p.Title,
		p.Description,
		p.Price,
		p.Link,
		p.Image,
		p.StyleText,
		p.UseCaseText,
		p.StyleCluster,
		p.UseCaseCluster,
		pgvector.NewVector(p.StyleEmbedding),
		pgvector.NewVector(p.UseCaseEmbedding),
	}

	var prefixes []string
	for i := range values {
		prefixes = append(prefixes, "$"+strconv.Itoa(i+1))
	}

	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (%s);",
		tableName,
		cn.ID,
		cn.Title,
		cn.Description,
		cn.Price,
		cn.Link,
		cn.Image,
		cn.StyleText,
		cn.UseCaseText,
		cn.StyleCluster,
		cn.UseCaseCluster,
		cn.StyleEmbedding,
		cn.UseCaseEmbedding,
		strings.Join(prefixes, ", "),
	)

	_, err = tx.Exec(stmt, values...)
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// Update updates the database entry of the product
func (p *Product) Update(db *sql.DB, tableName string, cn config.ColumnNames) error {
	stmt := fmt.Sprintf(`
		UPDATE %s 
		SET
			%s = $1,
			%s = $2,
			%s = $3,
			%s = $4,
			%s = $5,
			%s = $6,
			%s = $7,
			%s = $8,
			%s = $9,
			%s = $10
		WHERE %s = $11;
		`,
		tableName,

		cn.Title,
		cn.Description,

		cn.Link,
		cn.Image,

		cn.StyleText,
		cn.UseCaseText,

		cn.StyleEmbedding,
		cn.UseCaseEmbedding,

		cn.StyleCluster,
		cn.UseCaseCluster,

		cn.ID,
	)

	values := []any{
		p.Title,
		p.Description,
		p.Link,
		p.Image,
		p.StyleText,
		p.UseCaseText,
		pgvector.NewVector(p.StyleEmbedding),
		pgvector.NewVector(p.UseCaseEmbedding),
		p.StyleCluster,
		p.UseCaseCluster,
		p.ID,
	}

	_, err := db.Exec(stmt, values...)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	return nil
}
