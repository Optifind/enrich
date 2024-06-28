package handler

import (
	"context"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/lattots/enrich/pkg/config"
)

type Handler struct {
	MilvusClient client.Client
	Config       config.API
}

func New(conf config.API) (*Handler, error) {
	// Milvus client configuration is created.
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

	// Pointer to handler is returned.
	return &Handler{
		MilvusClient: milvusClient,
		Config:       conf,
	}, nil
}
