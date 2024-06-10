package catalog

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/product"
	"github.com/lattots/enrich/pkg/prompts"
	"github.com/lattots/openai-sdk"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"io"
	"log"
	"os"
)

// Catalog is a struct that represents product catalog.
type Catalog struct {
	InputFilepath    string             // Path to the input CSV file
	Products         []*product.Product // Slice of pointers to products that belong to the catalog
	MilvusCollection string             // Name of the Milvus collection
	MilvusClient     client.Client      // Milvus database client
	ColumnNames      config.ColumnNames // Milvus collection's column names
}

// New creates a new Catalog object according to configuration options.
// Returns a pointer to Catalog object and an error.
func New(conf config.Init) (*Catalog, error) {
	// Milvus client configuration is specified.
	milvusConf := client.Config{
		Address: conf.MilvusAddress,
		APIKey:  os.Getenv("MILVUS_TOKEN"),
	}

	// Milvus client is created.
	milvusClient, err := client.NewClient(
		context.Background(),
		milvusConf,
	)
	if err != nil {
		return nil, err
	}

	catalog := &Catalog{
		InputFilepath:    conf.CatalogFilepath,
		MilvusClient:     milvusClient,
		MilvusCollection: conf.MilvusCollection,
		ColumnNames:      conf.MilvusColumnNames,
	}

	return catalog, nil
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
func (c *Catalog) ProcessProducts(conf config.Init, prompts prompts.Prompts) error {
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

// InitCollection initializes a Milvus database collection with the catalogs' information.
// Returns an error.
func (c *Catalog) InitCollection() error {
	fmt.Println("Initializing Milvus collection...")
	// Schema for collection is created. This decides what columns are created.
	schema := c.createSchema()

	// Here the collection is created to the database with the specified schema.
	err := c.MilvusClient.CreateCollection(context.Background(), schema, 2)
	if err != nil {
		return err
	}

	return nil
}

// Struct for holding Milvus insert batch.
type batch struct {
	IDs          []string
	Titles       []string
	Descriptions []string
	Prices       []string

	Images []string
	Links  []string

	StyleTexts   []string
	UseCaseTexts []string

	StyleVectors [][]float32
}

// InsertToMilvus inserts all products of the catalog to Milvus collection.
// Returns an error.
func (c *Catalog) InsertToMilvus(conf config.Init) error {
	fmt.Println("Inserting product catalog to Milvus collection...")
	// This is a batch of information that is inserted to database.
	var batch batch

	// All products in catalog are added to the batch.
	for _, p := range c.Products {
		batch.IDs = append(batch.IDs, p.ID)
		batch.Titles = append(batch.Titles, p.Title)
		batch.Descriptions = append(batch.Descriptions, p.Description)
		batch.Prices = append(batch.Prices, p.Price)

		batch.Images = append(batch.Images, p.Image)
		batch.Links = append(batch.Links, p.Link)

		batch.StyleTexts = append(batch.StyleTexts, p.StyleText)
		batch.UseCaseTexts = append(batch.UseCaseTexts, p.UseCaseText)

		batch.StyleVectors = append(batch.StyleVectors, p.StyleEmbedding)
	}

	// Columns are created for all attributes.
	idColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.ID, batch.IDs)
	titleColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.Title, batch.Titles)
	descriptionColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.Description, batch.Descriptions)
	priceColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.Price, batch.Prices)

	imageColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.Image, batch.Images)
	linkColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.Link, batch.Links)

	styleTextColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.StyleText, batch.StyleTexts)
	useCaseColumn := entity.NewColumnVarChar(conf.MilvusColumnNames.UseCaseText, batch.UseCaseTexts)

	vectorColumn := entity.NewColumnFloatVector(conf.MilvusColumnNames.StyleEmbedding, len(batch.StyleVectors[0]), batch.StyleVectors)

	// Columns are inserted to database.
	_, err := c.MilvusClient.Insert(
		context.Background(),
		conf.MilvusCollection,
		"",
		idColumn,
		titleColumn,
		descriptionColumn,
		priceColumn,

		imageColumn,
		linkColumn,

		styleTextColumn,
		useCaseColumn,

		vectorColumn,
	)
	return err
}

// CreateIndex creates index for style embedding column. Returns an error.
func (c *Catalog) CreateIndex(conf config.Init) error {
	// Dimensionality of the embeddings in StyleEmbedding field.
	dim := len(c.Products[0].StyleEmbedding)

	// Index object with L2 similarity is created.
	idx, err := entity.NewIndexIvfFlat(
		entity.IP,
		dim,
	)
	if err != nil {
		return err
	}

	// Index for collection is created.
	err = c.MilvusClient.CreateIndex(
		context.Background(),
		conf.MilvusCollection,
		conf.MilvusColumnNames.StyleEmbedding,
		idx,
		false,
	)
	return err
}

// createSchema creates a schema object for a new Milvus collection. Returns the schema created.
func (c *Catalog) createSchema() *entity.Schema {
	var dim int
	if e := c.Products[0].StyleEmbedding; e != nil {
		dim = len(e) // If embeddings already exist, schema dimensionality is matched to embeddings
	} else {
		dim = 1536 // This is the default dimension count of OpenAI's "text-embedding-3-small" model
	}

	schema := &entity.Schema{
		CollectionName: c.MilvusCollection,
		Fields: []*entity.Field{
			{
				Name:       c.ColumnNames.ID,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{
					"max_length": "100",
				},
			},
			{
				Name:       c.ColumnNames.Title,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "500",
				},
			},
			{
				Name:       c.ColumnNames.Description,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "5000",
				},
			},
			{
				Name:       c.ColumnNames.Price,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "20",
				},
			},
			{
				Name:       c.ColumnNames.Image,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "1000",
				},
			},
			{
				Name:       c.ColumnNames.Link,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "1000",
				},
			},
			{
				Name:       c.ColumnNames.StyleText,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "5000",
				},
			},
			{
				Name:       c.ColumnNames.UseCaseText,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: false,
				TypeParams: map[string]string{
					"max_length": "5000",
				},
			},
			{
				Name:     c.ColumnNames.StyleEmbedding,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprint(dim),
				},
			},
		},
	}

	return schema
}
