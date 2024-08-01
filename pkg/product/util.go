package product

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/pgvector/pgvector-go"
)

// FromSQLRows converts SQL query results to a slice of Product
func FromSQLRows(rows *sql.Rows) ([]*Product, error) {
	defer rows.Close()
	products := make([]*Product, 0)
	for rows.Next() {
		p := new(Product)
		err := scanProduct(rows, p)
		if err != nil {
			return nil, fmt.Errorf("error parsing product from rows: %w", err)
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return products, nil
}

// FromSQLRow a single SQL query result to a Product
func FromSQLRow(row *sql.Row) (*Product, error) {
	p := new(Product)
	err := scanProduct(row, p)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Product not found
	}
	if err != nil {
		return nil, fmt.Errorf("error parsing product from row: %w", err)
	}
	return p, nil
}

// Scanner for scanning from sql.Row or sql.Rows to Product
// sql.Row and sql.Rows both implement scanner interface
type scanner interface {
	Scan(dest ...any) error
}

// scanProduct is a helper function for scanning SQL query results to Product
func scanProduct(scanner scanner, p *Product) error {
	var styleEmbedding, useEmbedding pgvector.Vector // Embeddings need to be temporarily scanned to pgvector.Vector type
	err := scanner.Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Price,
		&p.Link,
		&p.Image,
		&p.StyleText,
		&p.UseCaseText,
		&styleEmbedding,
		&useEmbedding,
	)
	if err != nil {
		return fmt.Errorf("error scanning product: %w", err)
	}
	// Embeddings are set from temporary pgvector.Vector values
	p.StyleEmbedding = styleEmbedding.Slice()
	p.UseCaseEmbedding = useEmbedding.Slice()

	return nil
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
