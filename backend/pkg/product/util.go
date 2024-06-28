package product

import (
	"errors"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// parseResultSet parses Milvus's client.ResultSet to a slice of pointers to products.
// Returns a slice of pointers and error.
func parseResultSet(res client.ResultSet) ([]*Product, error) {
	// Temporary slices for holding result information.
	ids := make([]string, 0)
	titles := make([]string, 0)
	descs := make([]string, 0)
	prices := make([]string, 0)
	links := make([]string, 0)
	images := make([]string, 0)

	// Data fields from all columns are added to result information slices.
	for _, col := range res {
		var err error
		switch col.Name() {
		case "id":
			ids, err = addValues(ids, col)
		case "title":
			titles, err = addValues(titles, col)
		case "description":
			descs, err = addValues(descs, col)
		case "price":
			prices, err = addValues(prices, col)
		case "link":
			links, err = addValues(links, col)
		case "image":
			images, err = addValues(images, col)
		default:
			return nil, fmt.Errorf("unknown column name: %s", col.Name())
		}
		if err != nil {
			return nil, err
		}
	}

	var products []*Product
	// Product objects are created from temporary product information slices.
	for i := range titles {
		p := &Product{
			ID:          ids[i],
			Title:       titles[i],
			Description: descs[i],
			Price:       prices[i],
			Link:        links[i],
			Image:       images[i],
		}
		products = append(products, p)
	}

	// Created slice of pointers to product objects is returned.
	return products, nil
}

// addValues adds all data fields from column to target slice. Returns the modified slice and error.
func addValues(target []string, column entity.Column) ([]string, error) {
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
