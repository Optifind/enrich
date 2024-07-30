package product

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
)

// FromSQLRows converts SQL query results to a slice of Product
func FromSQLRows(rows *sql.Rows) ([]*Product, error) {
	var products []*Product
	for rows.Next() {
		p := &Product{}
		err := scanProduct(rows, p)
		if err != nil {
			return nil, fmt.Errorf("error parsing product from rows: %s", err)
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %s", err)
	}

	return products, nil
}

// FromSQLRow a single SQL query result to a Product
func FromSQLRow(row *sql.Row) (*Product, error) {
	var p *Product
	err := scanProduct(row, p)
	if err != nil {
		return nil, fmt.Errorf("error parsing product from row: %s", err)
	}
	return p, nil
}

// scanProduct is a helper function for scanning SQL query results to Product
func scanProduct(scanner interface {
	Scan(dest ...interface{}) error
}, p *Product) error {
	err := scanner.Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Price,
		&p.Link,
		&p.Image,
		&p.StyleText,
		&p.UseCaseText,
		&p.StyleEmbedding,
		&p.UseCaseEmbedding,
	)
	if err != nil {
		return fmt.Errorf("error scanning product: %w", err)
	}
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
