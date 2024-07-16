package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/lattots/enrich/pkg/config"
)

type Handler struct {
	MilvusClient client.Client
	Config       config.Config
}

// New creates a new instance of Handler with given configuration file. Returns a pointer to Handler and an error.
func New(conf config.Config) (*Handler, error) {
	// Milvus client configuration is created.
	milvusConf := client.Config{
		Address: conf.API.MilvusAddress,
		APIKey:  os.Getenv("MILVUS_TOKEN"),
	}

	// Milvus client is created.
	milvusClient, err := client.NewClient(
		context.Background(),
		milvusConf,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Milvus client: %s\n", err)
	}

	// Pointer to handler is returned.
	return &Handler{
		MilvusClient: milvusClient,
		Config:       conf,
	}, nil
}
