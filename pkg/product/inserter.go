package product

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lattots/enrich/pkg/config"
)

type Inserter struct {
	db             *sql.DB
	productChannel <-chan *Product
	config         config.Database
}

func NewInserter(db *sql.DB, productChannel <-chan *Product, conf config.Database) *Inserter {
	return &Inserter{
		db:             db,
		productChannel: productChannel,
		config:         conf,
	}
}

// ListenAndInsert will listen the channel for incoming products and insert them to database.
// Method exits when channel is closed. Method errors if product insertion errors.
func (i *Inserter) ListenAndInsert() error {
	for p := range i.productChannel {
		err := i.insertProductWithRetry(p)
		if err != nil {
			// Log the error and continue instead of returning
			log.Printf("Failed to insert product %s: %v", p.ID, err)
			continue
		}
	}
	return nil
}

func (i *Inserter) insertProductWithRetry(p *Product) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = p.Insert(i.db, i.config.ProductTable, i.config.ColumnNames)
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("product %s insert failed after %d attempts: %w", p.ID, maxRetries, err)
}
