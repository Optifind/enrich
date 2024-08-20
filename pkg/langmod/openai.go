package langmod

import (
	"fmt"

	openai "github.com/lattots/openai-sdk"

	"github.com/lattots/enrich/pkg/config"
	"github.com/lattots/enrich/pkg/vector"
)

// OpenAI client. Implements LangMod interface. Can be used to create chat completions or text embeddings.
//
// Create new instance by calling langmod.NewOpenAI
type OpenAI struct {
	client          *openai.APIClient
	embeddingsModel string
	generativeModel string
}

// NewOpenAI is used to create OpenAI client using specified generative and embeddings models.
// Returns pointer to client and an error.
func NewOpenAI(token, embeddingsModel, generativeModel string) (*OpenAI, error) {
	client := openai.NewAPIClient(token)

	return &OpenAI{
		client:          client,
		embeddingsModel: embeddingsModel,
		generativeModel: generativeModel,
	}, nil
}

// NewOpenAIFromOption is used to create OpenAI client with just config.ProviderOption.
// Returns pointer to client and an error.
func NewOpenAIFromOption(token string, options config.ProviderOption) (*OpenAI, error) {
	client := openai.NewAPIClient(token)

	return &OpenAI{
		client:          client,
		embeddingsModel: options.EmbeddingsModel,
		generativeModel: options.GenerativeModel,
	}, nil
}

// CreateChatResponse completes the chat conversation with OpenAI's chat completions API.
// Returns result string and an error.
func (m *OpenAI) CreateChatResponse(args ...argument) (string, error) {
	// map for "system" and "user" messages
	// Finished map must contain "user" message but "system" is optional
	var messages []*openai.Message

	for _, arg := range args {
		if len(messages) == 0 {
			// If no messages exist, new message is created
			messages = append(messages, newMessage(arg))
		} else if lastMsg := messages[len(messages)-1]; lastMsg.Role == arg.getRole() {
			// If last message is from the same role as current message,
			// this new arguments content is added to that messages content
			lastMsg.Content = append(lastMsg.Content, openai.NewTextContent(arg.getText()))
		} else {
			// If last message is from another role, new message is created
			messages = append(messages, newMessage(arg))
		}
	}

	var prompt []openai.Message
	for _, msg := range messages {
		prompt = append(prompt, *msg)
	}

	const maxTokens = 3000

	res, err := m.client.CreateChatCompletion(m.generativeModel, prompt, maxTokens)
	if err != nil {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}

// newMessage converts argument to *openai.Message
func newMessage(arg argument) *openai.Message {
	var msg *openai.Message
	switch arg.getType() {
	case "text":
		msg = &openai.Message{
			Role: arg.getRole(),
			Content: []openai.Content{
				openai.NewTextContent(arg.getText()),
			},
		}
	case "image":
		msg = &openai.Message{
			Role: arg.getRole(),
			Content: []openai.Content{
				openai.NewImageContent(arg.getImage()),
			},
		}
	}

	return msg
}

// CreateEmbeddings creates text embeddings for input arguments.
// Currently only text embeddings are supported but images might be added in the future.
func (m *OpenAI) CreateEmbeddings(args ...argument) ([][]float32, error) {
	var embeddings [][]float32

	for _, arg := range args {
		if arg.getType() != "text" {
			return nil, fmt.Errorf("invalid argument type: %s", arg.getType())
		}
		res, err := m.client.CreateVectorEmbedding(m.embeddingsModel, arg.getText())
		if err != nil {
			return nil, err
		}

		emb := res.Data[0].Embedding

		embeddings = append(embeddings, vector.ToFloat32(emb))
	}

	return embeddings, nil
}
