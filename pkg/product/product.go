package product

import (
	"context"
	"errors"
	"fmt"

	openaisdk "github.com/lattots/openai-sdk"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/lattots/enrich/pkg/config"
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

// GetById fetches a product corresponding to ID from Milvus database. Returns a pointer to product and error.
func GetById(milvusClient client.Client, config config.Config, id string) (*Product, error) {
	ctx := context.Background()
	collection := config.API.MilvusCollection
	var partitionNames []string
	// Queried row is set. For now, we use VarChar as primary key type. This might change in the future.
	ids := entity.NewColumnVarChar(config.Init.MilvusColumnNames.ID, []string{id})
	outputFields := []string{"*"}

	// Database is queried.
	resultSet, err := milvusClient.QueryByPks(
		ctx,
		collection,
		partitionNames,
		ids,
		outputFields,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting product by id %s: %v", id, err)
	} else if resultSet.Len() != 1 {
		return nil, nil
	}

	// Result set is parsed to a slice of product pointers.
	products, err := ParseResultSet(config.Init.MilvusColumnNames, resultSet)
	if err != nil {
		return nil, fmt.Errorf("error parsing product by id %s: %v", id, err)
	}
	// Pointer to product is returned.
	return products[0], nil
}

// Process processes single product. This means creating style and
// use case texts and creating embeddings for some of the fields.
// Returns an error.
func (p *Product) Process(client openaisdk.APIClient, GPTModel, embeddingsModel string, prompts prompts.Prompts) error {
	// Style text is created and saved to product's StyleText field.
	err := p.CreateStyleText(client, GPTModel, prompts.StyleText)
	if err != nil {
		return err
	}

	// Use case text is created and saved to product's UseCaseText field.
	err = p.CreateUseCaseText(client, GPTModel, prompts.UseCaseText)
	if err != nil {
		return err
	}

	// Embeddings are created and saved to product's embeddings fields depending on which embeddings are specified.
	// See CreateEmbeddings for additional information.
	err = p.CreateEmbeddings(client, embeddingsModel)
	return err
}

// CreateStyleText creates style description with OpenAI's GPT model according to prompts.
// Resulting description is saved to product's StyleText field. Returns an error.
func (p *Product) CreateStyleText(client openaisdk.APIClient, GPTModel string, prompt []openaisdk.Message) error {
	// Product information is concatenated to a single string.
	productInfo := fmt.Sprintf("Title: %s\nDescription: %s", p.Title, p.Description)

	// Chat message object is created with the product's information as the content.
	productMsg := openaisdk.Message{
		Role: "user",
		Content: []openaisdk.Content{
			openaisdk.NewTextContent(productInfo),
			openaisdk.NewImageContent(p.Image),
		},
	}

	// This prompt includes the pre-defined "system" prompt and the product specific information.
	fullPrompt := append(prompt, productMsg)

	// Chat completion is created with the prompt messages and product information.
	resp, err := client.CreateChatCompletion(GPTModel, fullPrompt, 3000)
	if err != nil {
		return err
	}

	// Style text is saved to the product object.
	p.StyleText = resp.Choices[0].Message.Content
	return nil
}

// CreateUseCaseText creates use case description with OpenAI's GPT model according to prompts.
// Resulting description is saved to product's UseCaseText field. Returns an error.
func (p *Product) CreateUseCaseText(client openaisdk.APIClient, GPTModel string, prompt []openaisdk.Message) error {
	// Product information is concatenated to a single string.
	productInfo := fmt.Sprintf("Title: %s\nDescription: %s", p.Title, p.Description)

	// Chat message object is created with the product's information as the content.
	productMsg := openaisdk.Message{
		Role: "user",
		Content: []openaisdk.Content{
			openaisdk.NewTextContent(productInfo),
			openaisdk.NewImageContent(p.Image),
		},
	}

	// This prompt includes the pre-defined "system" prompt and the product specific information.
	fullPrompt := append(prompt, productMsg)

	// Chat completion is created with the prompt messages and product information.
	resp, err := client.CreateChatCompletion(GPTModel, fullPrompt, 3000)
	if err != nil {
		return err
	}

	// Use case text is saved to the product object.
	p.UseCaseText = resp.Choices[0].Message.Content
	return nil
}

// CreateEmbeddings creates embeddings some or all fields of the product.
// Currently, method only embeds the style description of the
// product and saves the embedding to product's StyleEmbedding field.
// Returns an error.
func (p *Product) CreateEmbeddings(client openaisdk.APIClient, embeddingsModel string) error {
	// Embedding for the StyleText is created.
	resp, err := client.CreateVectorEmbedding(embeddingsModel, p.StyleText)
	if err != nil {
		return err
	}

	// This is the float64 embedding for style text
	f64Vector := resp.Data[0].Embedding
	// Float64 must be converted to float32 for later use
	f32Vector := make([]float32, len(f64Vector))
	for i, f := range f64Vector {
		f32Vector[i] = float32(f)
	}

	// Created embedding is saved to product object's StyleEmbedding field.
	p.StyleEmbedding = f32Vector
	return nil
}

// CreateAttributes TODO: Method should create product attributes with GPT and parse them to Attributes object.
func (p *Product) CreateAttributes(client openaisdk.APIClient, GPTModel string, prompt []openaisdk.Message) error {
	return errors.New("NOT IMPLEMENTED")
}
