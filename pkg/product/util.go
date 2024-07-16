package product

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/lattots/enrich/pkg/config"
)

// ParseResultSet parses Milvus's client.ResultSet to a slice of pointers to products.
// Returns a slice of pointers and error.
func ParseResultSet(columnNames config.ColumnNames, res client.ResultSet) ([]*Product, error) {
	// Temporary slices for holding result information.
	ids := make([]string, 0)
	titles := make([]string, 0)
	descs := make([]string, 0)
	prices := make([]float32, 0)
	links := make([]string, 0)
	images := make([]string, 0)
	embeddings := make([][]float32, 0)
	styleTexts := make([]string, 0)
	useCaseTexts := make([]string, 0)

	// Data fields from all columns are added to result information slices.
	for _, col := range res {
		var err error
		switch col.Name() {
		case columnNames.ID:
			ids, err = addStrings(ids, col)
		case columnNames.Title:
			titles, err = addStrings(titles, col)
		case columnNames.Description:
			descs, err = addStrings(descs, col)
		case columnNames.Price:
			prices, err = addFloat32s(prices, col)
		case columnNames.Link:
			links, err = addStrings(links, col)
		case columnNames.Image:
			images, err = addStrings(images, col)
		case columnNames.StyleEmbedding:
			embeddings, err = addEmbeddings(embeddings, col)
		case columnNames.StyleText:
			styleTexts, err = addStrings(styleTexts, col)
		case columnNames.UseCaseText:
			useCaseTexts, err = addStrings(useCaseTexts, col)
		default:
			log.Printf("unknown column name: %s", col.Name())
		}
		if err != nil {
			return nil, err
		}
	}

	var products []*Product
	// Product objects are created from temporary product information slices.
	for i := range titles {
		p := &Product{
			ID:             ids[i],
			Title:          titles[i],
			Description:    descs[i],
			Price:          prices[i],
			Link:           links[i],
			Image:          images[i],
			StyleEmbedding: embeddings[i],
			StyleText:      styleTexts[i],
			UseCaseText:    useCaseTexts[i],
		}
		products = append(products, p)
	}

	// Created slice of pointers to product objects is returned.
	return products, nil
}

// addStrings adds all data fields from column to target slice. Returns the modified slice and error.
func addStrings(target []string, column entity.Column) ([]string, error) {
	// Varchar / string column is retrieved from column.
	strCol, ok := column.(*entity.ColumnVarChar)
	// If column doesn't contain varchar / string column, function errors.
	if !ok {
		return nil, errors.New("column is not var char")
	}
	// All elements from data column are added to target slice.
	target = append(target, strCol.Data()...)
	return target, nil
}

// addFloat32s adds all data fields from column to target slice. Returns the modified slice and error.
func addFloat32s(target []float32, column entity.Column) ([]float32, error) {
	// Float column is retrieved from column.
	floatCol, ok := column.(*entity.ColumnFloat)
	if !ok {
		return nil, errors.New("column is not float")
	}
	// All elements from data column are added to target slice.
	target = append(target, floatCol.Data()...)
	return target, nil
}

func addEmbeddings(target [][]float32, column entity.Column) ([][]float32, error) {
	// Vector column is retrieved from column.
	strCol, ok := column.(*entity.ColumnFloatVector)
	// If column doesn't contain varchar / string column, function errors.
	if !ok {
		return nil, errors.New("column is not float32 vector")
	}
	// All elements from data column are added to target slice.
	target = append(target, strCol.Data()...)
	return target, nil
}

func extractPrice(s string) (float32, error) {
	// This pattern matches (n * number . 2 * number) which should match the price format in input CSV.
	pattern := `^\d+\.\d{2}`
	re := regexp.MustCompile(pattern)

	// Find the match
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no price found in string")
	}

	// Convert the matched string to float32
	price, err := strconv.ParseFloat(match, 32)
	if err != nil {
		return 0, err
	}

	return float32(price), nil
}
