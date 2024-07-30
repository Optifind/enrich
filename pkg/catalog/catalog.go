package catalog

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	openaisdk "github.com/lattots/openai-sdk"
	_ "github.com/lib/pq"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/product"
	"github.com/lattots/enrich/pkg/prompts"
	"github.com/lattots/enrich/pkg/vector"
)

// Catalog is a struct that represents a group of products.
type Catalog struct {
	InputFilepath string             // Path to the input CSV file
	Products      []*product.Product // Slice of pointers to products that belong to the catalog
	DB            *sql.DB            // Database handle for product database
	ProductTable  string             // Name of the product table
	ColumnNames   config.ColumnNames // Milvus collection's column names
}

// New creates a new Catalog object according to configuration options.
// Returns a pointer to Catalog object and an error.
func New(conf config.Config) (*Catalog, error) {
	// Database connection is opened to Postgresql database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	catalog := &Catalog{
		InputFilepath: conf.Filepaths.CatalogFilepath,
		DB:            db,
		ProductTable:  conf.DB.ProductTable,
		ColumnNames:   conf.DB.ColumnNames,
	}

	return catalog, nil
}

// Load fetches product's in Catalog's product database to memory.
//
// `count` (required) determines the maximum amount of products that will be loaded.
//
// `ids` (optional) determines the products that will be loaded.
//
// Returns an error.
func (c *Catalog) Load(count int, ids ...string) error {
	if count <= 0 {
		return errors.New("count must be greater than 0")
	}
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT ")
	queryBuilder.WriteString(strings.Join([]string{
		c.ColumnNames.ID,
		c.ColumnNames.Title,
		c.ColumnNames.Description,
		c.ColumnNames.Price,
		c.ColumnNames.Link,
		c.ColumnNames.Image,
		c.ColumnNames.StyleText,
		c.ColumnNames.UseCaseText,
		c.ColumnNames.StyleEmbedding,
		c.ColumnNames.UseCaseEmbedding,
	}, ", "))
	queryBuilder.WriteString(" FROM ")
	queryBuilder.WriteString(c.ProductTable)

	var args []interface{}
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args = append(args, id)
		}
		queryBuilder.WriteString(" WHERE id IN (")
		queryBuilder.WriteString(strings.Join(placeholders, ", "))
		queryBuilder.WriteString(")")
	}

	queryBuilder.WriteString(" LIMIT ?")
	args = append(args, count)

	query := queryBuilder.String()

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products, err := product.FromSQLRows(rows)
	if err != nil {
		return fmt.Errorf("error parsing product query results: %w", err)
	}

	c.Products = products

	return nil
}

// CreateProducts reads input CSV file and creates product objects based on the input.
// Returns an error.
func (c *Catalog) CreateProducts() error {
	fmt.Println("Reading products from file...")
	// Input file is opened.
	file, err := os.Open(c.InputFilepath)
	if err != nil {
		return err
	}
	defer file.Close() // Input file is closed when possible

	// CSV reader is created. This reader will read the file row by row,
	// allowing for efficient processing of large CSV files.
	r := csv.NewReader(file)

	// Header row is skipped.
	_, err = r.Read()
	if err != nil {
		return err
	}

	// Row's are converted to Product's until end-of-file is reached.
	for {
		// File is read row by row with the CSV reader.
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Product object is created based on the row.
		p, err := product.FromRow(row)
		if err != nil {
			log.Printf("error creating product from row: %s\n%s\n", row, err)
		}

		// Created Product is added to catalog's products.
		c.Products = append(c.Products, p)
	}

	return nil
}

// ProcessProducts processes all products. After this all products will have style and use case descriptions and embeddings for some of their fields.
// See Product.Process() for additional information.
// Returns an error.
func (c *Catalog) ProcessProducts(conf config.OpenAI, prompts prompts.Prompts) error {
	fmt.Println("Processing products...")
	openAIClient := openaisdk.APIClient{APIKey: os.Getenv("OPENAI_TOKEN")}
	for _, p := range c.Products {
		err := p.Process(openAIClient, conf.GPTModel, conf.EmbeddingsModel, prompts)
		if err != nil {
			return err
		}
	}
	return nil
}

// InitDatabase initializes a new database collection with the catalogs' information.
// Returns an error.
func (c *Catalog) InitDatabase() error {
	fmt.Println("Initializing database collection...")

	// Database table for products is created.
	err := c.CreateTable()
	if err != nil {
		return fmt.Errorf("error creating database table: %w", err)
	}

	// Products are inserted to database.
	err = c.InsertToDB()
	if err != nil {
		return fmt.Errorf("error inserting products: %w", err)
	}

	return nil
}

func (c *Catalog) CreateTable() error {
	fmt.Println("Creating table...")
	query := fmt.Sprintf(`CREATE TABLE %s IF NOT EXISTS (
		%s VARCHAR(15) PRIMARY KEY NOT NULL,
		%s VARCHAR(255),
		%s TEXT,
		%s NUMERIC(10, 2),
    	%s VARCHAR(255),
    	%s VARCHAR(255),
		%s TEXT,
		%s TEXT,
		%s VECTOR(3072),
		%s VECTOR(3072),);`,
		c.ProductTable,
		c.ColumnNames.ID,
		c.ColumnNames.Title,
		c.ColumnNames.Description,
		c.ColumnNames.Price,
		c.ColumnNames.Link,
		c.ColumnNames.Image,
		c.ColumnNames.StyleText,
		c.ColumnNames.UseCaseText,
		c.ColumnNames.StyleEmbedding,
		c.ColumnNames.UseCaseEmbedding,
	)

	_, err := c.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}
	return nil
}

// InsertToDB inserts all products of the catalog into databases product table.
// Returns an error.
func (c *Catalog) InsertToDB() error {
	prefixes := make([]string, len(c.Products))
	values := make([]interface{}, len(c.Products)*10)
	for _, p := range c.Products {
		prefixes = append(prefixes, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		values = append(
			values,
			p.ID,
			p.Title,
			p.Description,
			p.Price,
			p.Link,
			p.Image,
			p.StyleText,
			p.UseCaseText,
			p.StyleEmbedding,
			p.UseCaseEmbedding,
		)
	}
	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES %s",
		c.ProductTable,
		c.ColumnNames.ID,
		c.ColumnNames.Title,
		c.ColumnNames.Description,
		c.ColumnNames.Price,
		c.ColumnNames.Link,
		c.ColumnNames.Image,
		c.ColumnNames.StyleText,
		c.ColumnNames.UseCaseText,
		c.ColumnNames.StyleEmbedding,
		c.ColumnNames.UseCaseEmbedding,
		strings.Join(prefixes, ", "),
	)

	_, err := c.DB.Exec(stmt, values...)
	if err != nil {
		return fmt.Errorf("error inserting products into table: %w", err)
	}
	return nil
}

// UpdateIndex creates index for style embeddings and use case embeddings. Returns an error.
func (c *Catalog) UpdateIndex() error {
	// Old indices are dropped if they already exist.
	dropStyleIndex := fmt.Sprintf(
		"DROP INDEX IF EXISTS %s_%s_idx",
		c.ProductTable,
		c.ColumnNames.StyleEmbedding,
	)
	_, err := c.DB.Exec(dropStyleIndex)
	if err != nil {
		return fmt.Errorf("error dropping index: %w", err)
	}
	dropUseCaseIndex := fmt.Sprintf(
		"DROP INDEX IF EXISTS %s_%s_idx",
		c.ProductTable,
		c.ColumnNames.UseCaseEmbedding,
	)
	_, err = c.DB.Exec(dropUseCaseIndex)
	if err != nil {
		return fmt.Errorf("error dropping index: %w", err)
	}

	// Statements for creating new indices
	// TODO: Currently we use default values for m and ef_construction in index creation.
	//  It might be smart to try out other parameters
	createStyleIndex := fmt.Sprintf(
		"CREATE INDEX ON %s USING hnsw (%s vector_cosine_ops)",
		c.ProductTable,
		c.ColumnNames.StyleEmbedding,
	)
	_, err = c.DB.Exec(createStyleIndex)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	createUseCaseIndex := fmt.Sprintf(
		"CREATE INDEX ON %s USING hnsw (%s vector_cosine_ops)",
		c.ProductTable,
		c.ColumnNames.UseCaseEmbedding,
	)
	_, err = c.DB.Exec(createUseCaseIndex)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	return nil
}

// SearchSimilarStyle searches for products most similar to Catalog's Products in design style.
//
// Returns up to `count` number of product.Product pointers and error.
func (c *Catalog) SearchSimilarStyle(count int) ([]*product.Product, error) {
	// Temporary slice for holding all product embeddings in Catalog.
	productVecs := make([][]float32, len(c.Products))
	for i, p := range c.Products {
		productVecs[i] = p.StyleEmbedding
	}
	// Average product embedding is calculated.
	// This vector will be used in the search.
	avgVec := vector.GetAverageVec(productVecs)

	results, err := c.vectorSearch(avgVec, c.ColumnNames.StyleEmbedding, count)
	if err != nil {
		return nil, fmt.Errorf("error vector searching db: %w", err)
	}

	return results, nil
}

// SearchSimilarUseCase searches for products most similar to Catalog's Products in design style.
//
// Returns up to `count` number of product.Product pointers and error.
func (c *Catalog) SearchSimilarUseCase(count int) ([]*product.Product, error) {
	// Temporary slice for holding all product embeddings in Catalog.
	productVecs := make([][]float32, len(c.Products))
	for i, p := range c.Products {
		productVecs[i] = p.UseCaseEmbedding
	}
	// Average product embedding is calculated.
	// This vector will be used in the search.
	avgVec := vector.GetAverageVec(productVecs)

	results, err := c.vectorSearch(avgVec, c.ColumnNames.UseCaseEmbedding, count)
	if err != nil {
		return nil, fmt.Errorf("error vector searching db: %w", err)
	}

	return results, nil
}

// vectorSearch searches the database with the given vector on the given column. It returns up to count number of products.
func (c *Catalog) vectorSearch(vec []float32, vecColumn string, count int) ([]*product.Product, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s ORDER BY %s <=> %v LIMIT %d",
		c.ColumnNames.ID,
		c.ColumnNames.Title,
		c.ColumnNames.Description,
		c.ColumnNames.Price,
		c.ColumnNames.Link,
		c.ColumnNames.Image,
		c.ColumnNames.StyleText,
		c.ColumnNames.UseCaseText,
		c.ColumnNames.StyleEmbedding,
		c.ColumnNames.UseCaseEmbedding,
		c.ProductTable,
		vecColumn,
		vec,
		count,
	)

	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error searching products: %w", err)
	}
	defer rows.Close()

	products, err := product.FromSQLRows(rows)
	if err != nil {
		return nil, fmt.Errorf("error parsing query results: %w", err)
	}

	return products, nil
}
