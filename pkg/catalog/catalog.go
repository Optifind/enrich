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
	"time"

	_ "github.com/lib/pq"
	"github.com/mpraski/clusters"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/langmod"
	"github.com/lattots/enrich/pkg/product"
	"github.com/lattots/enrich/pkg/prompts"
	"github.com/lattots/enrich/pkg/vector"
)

// Catalog is a struct that represents a group of products.
type Catalog struct {
	InputFilepath string             // Path to the input CSV file
	Products      []*product.Product // Slice of pointers to products that belong to the catalog
	DB            *sql.DB            // Database handle for product database
	DBConfig      config.Database    // Database configurations like product table name and column names
	EmbeddingDim  int                // Number of dimensions used by embeddings
	langMod       langmod.LangMod    // Language model used by this catalog to create products
}

// New creates a new Catalog object according to configuration options.
// Returns a pointer to Catalog object and an error.
func New(conf config.Config) (*Catalog, error) {
	// Database connection is opened to Postgresql database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	var langMod langmod.LangMod
	if conf.LM.ActiveProvider == "openai" {
		var err error
		langMod, err = langmod.NewOpenAIFromOption(os.Getenv("OPENAI_TOKEN"), conf.LM.Providers.OpenAI)
		if err != nil {
			return nil, fmt.Errorf("error creating OpenAI client: %w", err)
		}
	} else if conf.LM.ActiveProvider == "gemini" {
		var err error
		langMod, err = langmod.NewGeminiFromOption(os.Getenv("GEMINI_TOKEN"), conf.LM.Providers.Gemini)
		if err != nil {
			return nil, fmt.Errorf("error creating Gemini client: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unknown language model provider: %s", conf.LM.ActiveProvider)
	}

	dim, err := langMod.GetEmbeddingDimensions()
	if err != nil {
		return nil, fmt.Errorf("error getting embedding dimensions: %w", err)
	}

	catalog := &Catalog{
		InputFilepath: conf.Filepaths.CatalogFilepath,
		DB:            db,
		DBConfig:      conf.DB,
		langMod:       langMod,
		EmbeddingDim:  dim,
	}

	return catalog, nil
}

// Load fetches product's in Catalog's product database to memory.
//
// `count` (optional) determines the maximum amount of products that will be loaded. 0 will load all available products.
//
// `ids` (optional) determines the products that will be loaded.
//
// Returns an error.
func (c *Catalog) Load(count int, ids ...string) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT ")
	queryBuilder.WriteString(strings.Join([]string{
		c.DBConfig.ColumnNames.ID,
		c.DBConfig.ColumnNames.Title,
		c.DBConfig.ColumnNames.Description,
		c.DBConfig.ColumnNames.Price,
		c.DBConfig.ColumnNames.Link,
		c.DBConfig.ColumnNames.Image,
		c.DBConfig.ColumnNames.StyleText,
		c.DBConfig.ColumnNames.UseCaseText,
		c.DBConfig.ColumnNames.StyleEmbedding,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
	}, ", "))
	queryBuilder.WriteString(" FROM ")
	queryBuilder.WriteString(c.DBConfig.ProductTable)

	var args []interface{}
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, id)
		}
		queryBuilder.WriteString(" WHERE id IN (")
		queryBuilder.WriteString(strings.Join(placeholders, ", "))
		queryBuilder.WriteString(")")
	}

	if count > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d;", len(ids)+1))
		args = append(args, count)
	}
	if count < 0 {
		return errors.New("count must be 0 or larger")
	}

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
func (c *Catalog) ProcessProducts(prompts prompts.Prompts) error {
	fmt.Println("Processing products...")

	// errgroup is package for handling errors in a concurrent application
	var pErrGroup errgroup.Group // product processor error group
	var iErrGroup errgroup.Group // product inserter error group

	productCh := make(chan *product.Product) // Channel for passing products to inserter
	inserter := product.NewInserter(c.DB, productCh, c.DBConfig)
	iErrGroup.Go(inserter.ListenAndInsert) // Product inserter is started

	// All products are processed and inserted to database
	for i, p := range c.Products {
		// Sub-process in the product error group is started
		pErrGroup.Go(func() error {
			start := time.Now()
			// Product is processed
			// If processing fails, it is retried once
			if err := p.Process(c.langMod, prompts); err != nil {
				fmt.Printf("error processing product %s: %s\nTrying again...\n", p.ID, err)
				// Processing is retried once
				// If this fails, method exits with errors
				err = p.Process(c.langMod, prompts)
				if err != nil {
					return fmt.Errorf("product %d: %w", i, err)
				}
			}
			fmt.Printf("Product %d processed in %s\n", i, time.Since(start))
			productCh <- p // Processed product is passed to inserter via the channel
			return nil
		})
		// Program sleeps to prevent hitting API limiter
		// TODO: Optimize sleep time to better utilize available API capacity
		time.Sleep(200 * time.Millisecond)
	}

	// Wait for all goroutines to finish and return the first returned error
	if err := pErrGroup.Wait(); err != nil {
		return fmt.Errorf("error processing product: %w", err)
	}
	close(productCh) // Closing the product channel stops the inserter

	// If inserter stops with no errors, method exits with nil
	if err := iErrGroup.Wait(); err != nil {
		return fmt.Errorf("error inserting products: %w", err)
	}
	return nil
}

// InitDatabase initializes a new database collection with the catalogs' information.
// Returns an error.
func (c *Catalog) InitDatabase() error {
	fmt.Println("Initializing database...")

	// Existing product table is dropped.
	err := c.dropTable()
	if err != nil {
		return fmt.Errorf("error dropping table: %w", err)
	}

	// Database table for products is created.
	err = c.createTable()
	if err != nil {
		return fmt.Errorf("error creating database table: %w", err)
	}

	return nil
}

func (c *Catalog) dropTable() error {
	fmt.Println("Dropping table...")
	stmt := fmt.Sprintf("DROP TABLE IF EXISTS %s", c.DBConfig.ProductTable)
	_, err := c.DB.Exec(stmt)
	if err != nil {
		return fmt.Errorf("error dropping table: %w", err)
	}
	return nil
}

func (c *Catalog) createTable() error {
	fmt.Println("Creating table...")
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s VARCHAR(255) PRIMARY KEY NOT NULL,
		%s VARCHAR(255),
		%s TEXT,
		%s NUMERIC(10, 2),
    	%s VARCHAR(255),
    	%s VARCHAR(255),
		%s TEXT,
		%s TEXT,
    	%s SMALLINT,
    	%s SMALLINT,
		%s VECTOR(%d),
		%s VECTOR(%d)
    	);`,
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.ID,
		c.DBConfig.ColumnNames.Title,
		c.DBConfig.ColumnNames.Description,
		c.DBConfig.ColumnNames.Price,
		c.DBConfig.ColumnNames.Link,
		c.DBConfig.ColumnNames.Image,
		c.DBConfig.ColumnNames.StyleText,
		c.DBConfig.ColumnNames.UseCaseText,
		c.DBConfig.ColumnNames.StyleCluster,
		c.DBConfig.ColumnNames.UseCaseCluster,
		c.DBConfig.ColumnNames.StyleEmbedding,
		c.EmbeddingDim,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
		c.EmbeddingDim,
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
	fmt.Printf("Inserting %d products...\n", len(c.Products))
	prefixes := make([]string, len(c.Products))
	values := make([]any, 0)
	var v []any
	for i, p := range c.Products {
		v = []any{
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

		prefixes[i] = getPrefix(i, len(v))

		values = append(values, v...)
	}

	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES %s;",
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.ID,
		c.DBConfig.ColumnNames.Title,
		c.DBConfig.ColumnNames.Description,
		c.DBConfig.ColumnNames.Price,
		c.DBConfig.ColumnNames.Link,
		c.DBConfig.ColumnNames.Image,
		c.DBConfig.ColumnNames.StyleText,
		c.DBConfig.ColumnNames.UseCaseText,
		c.DBConfig.ColumnNames.StyleCluster,
		c.DBConfig.ColumnNames.UseCaseCluster,
		c.DBConfig.ColumnNames.StyleEmbedding,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
		strings.Join(prefixes, ", "),
	)

	_, err := c.DB.Exec(stmt, values...)
	if err != nil {
		return fmt.Errorf("error inserting products into table: %w", err)
	}
	return nil
}

// getPrefix returns the value placeholder string used by Postgres.
// It is used for inserting multiple values to multiple database rows.
func getPrefix(i, numOfValues int) string {
	prefix := "("
	var valPlaceholders []string
	for j := 1; j <= numOfValues; j++ {
		valPlaceholders = append(valPlaceholders, fmt.Sprintf("$%d", i*numOfValues+j))
	}

	prefix += strings.Join(valPlaceholders, ", ")

	prefix += ")"
	return prefix
}

// UpdateProducts updates database entries of all products in catalog. Returns an error.
func (c *Catalog) UpdateProducts() error {
	fmt.Println("Updating products...")
	var err error
	for _, p := range c.Products {
		err = p.Update(c.DB, c.DBConfig.ProductTable, c.DBConfig.ColumnNames)
		if err != nil {
			return fmt.Errorf("error updating product: %w", err)
		}
	}
	return nil
}

// UpdateIndex creates index for style embeddings and use case embeddings. Returns an error.
func (c *Catalog) UpdateIndex() error {
	// Old indices are dropped if they already exist.
	dropStyleIndex := fmt.Sprintf(
		"DROP INDEX IF EXISTS %s_%s_idx",
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.StyleEmbedding,
	)
	_, err := c.DB.Exec(dropStyleIndex)
	if err != nil {
		return fmt.Errorf("error dropping index: %w", err)
	}
	dropUseCaseIndex := fmt.Sprintf(
		"DROP INDEX IF EXISTS %s_%s_idx",
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
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
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.StyleEmbedding,
	)
	_, err = c.DB.Exec(createStyleIndex)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	createUseCaseIndex := fmt.Sprintf(
		"CREATE INDEX ON %s USING hnsw (%s vector_cosine_ops)",
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
	)
	_, err = c.DB.Exec(createUseCaseIndex)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	return nil
}

// Cluster performs clustering for all products in catalog for both style and use case embeddings.
// Adds cluster numbers for all products. Returns an error.
func (c *Catalog) Cluster(conf config.Clusters) error {
	switch conf.Algorithm {
	case "k-means":
		err := c.kMeansClusterColumn(conf.KMeans.StyleK, c.DBConfig.ColumnNames.StyleEmbedding)
		if err != nil {
			return fmt.Errorf("error clustering products based on style: %w", err)
		}
		err = c.kMeansClusterColumn(conf.KMeans.UseCaseK, c.DBConfig.ColumnNames.UseCaseEmbedding)
		if err != nil {
			return fmt.Errorf("error clustering products based on use case: %w", err)
		}
	case "dbscan":
		err := c.dBSCANClusterColumn(conf.DBSCAN.MinPoints, conf.DBSCAN.Epsilon, c.DBConfig.ColumnNames.StyleEmbedding)
		if err != nil {
			return fmt.Errorf("error clustering products based on style: %w", err)
		}
		err = c.dBSCANClusterColumn(conf.DBSCAN.MinPoints, conf.DBSCAN.Epsilon, c.DBConfig.ColumnNames.UseCaseEmbedding)
		if err != nil {
			return fmt.Errorf("error clustering products based on use case: %w", err)
		}
	default:
		return fmt.Errorf("unknown clustering algorithm: expected \"k-means\" or \"dbscan\", got \"%s\"", conf.Algorithm)
	}

	return nil
}

// clusterColumn performs k-means clustering for products in catalog for the specified column.
// Adds cluster number for all products in catalog. Returns an error.
func (c *Catalog) kMeansClusterColumn(k int, column string) error {
	data := make([][]float64, len(c.Products))
	for i, p := range c.Products {
		switch column {
		case c.DBConfig.ColumnNames.StyleEmbedding:
			data[i] = toFloat64(p.StyleEmbedding)
		case c.DBConfig.ColumnNames.UseCaseEmbedding:
			data[i] = toFloat64(p.UseCaseEmbedding)
		default:
			return errors.New("unknown column")
		}
	}

	const maxIterations = 1000
	kmeans, err := clusters.KMeans(maxIterations, k, clusters.EuclideanDistance)
	if err != nil {
		return fmt.Errorf("error creating clusterer: %w", err)
	}

	err = kmeans.Learn(data)
	if err != nil {
		return fmt.Errorf("error clustering products: %w", err)
	}

	for idx, cluster := range kmeans.Guesses() {
		switch column {
		case c.DBConfig.ColumnNames.StyleEmbedding:
			c.Products[idx].StyleCluster = cluster
		case c.DBConfig.ColumnNames.UseCaseEmbedding:
			c.Products[idx].UseCaseCluster = cluster
		}
	}

	return nil
}

// dBSCANClusterColumn performs DBSCAN clustering for products in catalog for the specified column.
// Adds cluster number for all products in catalog. Returns an error.
func (c *Catalog) dBSCANClusterColumn(minpts int, eps float64, column string) error {
	data := make([][]float64, len(c.Products))
	for i, p := range c.Products {
		switch column {
		case c.DBConfig.ColumnNames.StyleEmbedding:
			data[i] = toFloat64(p.StyleEmbedding)
		case c.DBConfig.ColumnNames.UseCaseEmbedding:
			data[i] = toFloat64(p.UseCaseEmbedding)
		}
	}

	// Clusterer object is created
	dbscan, err := clusters.DBSCAN(minpts, eps, 0, clusters.EuclideanDistance)
	if err != nil {
		return fmt.Errorf("error creating clusterer: %w", err)
	}

	err = dbscan.Learn(data)
	if err != nil {
		return fmt.Errorf("error clustering products: %w", err)
	}

	for idx, cluster := range dbscan.Guesses() {
		switch column {
		case c.DBConfig.ColumnNames.StyleEmbedding:
			c.Products[idx].StyleCluster = cluster
		case c.DBConfig.ColumnNames.UseCaseEmbedding:
			c.Products[idx].UseCaseCluster = cluster
		}
	}

	return nil
}

// SearchSimilarStyle searches for products most similar to Catalog's Products in design style.
//
// Returns up to `count` number of product.Product pointers and error.
func (c *Catalog) SearchSimilarStyle(count int) ([]*product.Product, error) {
	results, err := c.searchSimilar(count, c.DBConfig.ColumnNames.StyleEmbedding)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// SearchSimilarUseCase searches for products most similar to Catalog's Products in design style.
//
// Returns up to `count` number of product.Product pointers and error.
func (c *Catalog) SearchSimilarUseCase(count int) ([]*product.Product, error) {
	results, err := c.searchSimilar(count, c.DBConfig.ColumnNames.UseCaseEmbedding)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Catalog) searchSimilar(count int, metricColumn string) ([]*product.Product, error) {
	// Slice for holding all product IDs in the Catalog
	ids := make([]string, len(c.Products))
	// Temporary slice for holding all product embeddings in Catalog.
	productVecs := make([][]float32, len(c.Products))
	for i, p := range c.Products {
		ids[i] = p.ID
		switch metricColumn {
		case c.DBConfig.ColumnNames.UseCaseEmbedding:
			productVecs[i] = p.UseCaseEmbedding
		case c.DBConfig.ColumnNames.StyleEmbedding:
			productVecs[i] = p.StyleEmbedding
		default:
			return nil, errors.New(fmt.Sprintf("unknown column: %s", metricColumn))
		}
	}
	// Average product embedding is calculated.
	// This vector will be used in the search.
	avgVec := vector.GetAverageVec(productVecs)

	// Database is searched for
	results, err := c.vectorSearch(avgVec, metricColumn, count, ids)
	if err != nil {
		return nil, fmt.Errorf("error vector searching db: %w", err)
	}

	return results, nil
}

// vectorSearch searches the database with the given vector on the given column. It returns up to count number of products.
func (c *Catalog) vectorSearch(vec []float32, vecColumn string, count int, excludedIDs []string) ([]*product.Product, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s NOT IN (%s) ORDER BY %s <=> $1 LIMIT %d;",
		c.DBConfig.ColumnNames.ID,
		c.DBConfig.ColumnNames.Title,
		c.DBConfig.ColumnNames.Description,
		c.DBConfig.ColumnNames.Price,
		c.DBConfig.ColumnNames.Link,
		c.DBConfig.ColumnNames.Image,
		c.DBConfig.ColumnNames.StyleText,
		c.DBConfig.ColumnNames.UseCaseText,
		c.DBConfig.ColumnNames.StyleEmbedding,
		c.DBConfig.ColumnNames.UseCaseEmbedding,
		c.DBConfig.ProductTable,
		c.DBConfig.ColumnNames.ID,
		placeholderList(len(excludedIDs)), // Generate placeholders
		vecColumn,
		count,
	)

	args := []interface{}{pgvector.NewVector(vec)}
	for _, id := range excludedIDs {
		args = append(args, id)
	}

	rows, err := c.DB.Query(query, args...)
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

// placeholderList generates a comma-separated list of placeholders for PostgreSQL queries.
func placeholderList(n int) string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+2) // Starting from $2 as $1 is used for the vector
	}
	return strings.Join(placeholders, ", ")
}

func toFloat64(vec []float32) []float64 {
	res := make([]float64, len(vec))
	for i, v := range vec {
		res[i] = float64(v)
	}
	return res
}
