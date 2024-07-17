package handler

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/lattots/enrich/pkg/config"
)

type Handler struct {
	DB     *sql.DB
	Config config.Config
}

// New creates a new instance of Handler with given configuration file. Returns a pointer to Handler and an error.
func New(conf config.Config) (*Handler, error) {
	// Database connection is opened.
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database connection: %w", err)
	}
	return &Handler{DB: db, Config: conf}, nil
}
