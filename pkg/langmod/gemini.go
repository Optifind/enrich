package langmod

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/lattots/enrich/pkg/config"
)

// Gemini client. Implements LangMod interface. Can be used to create chat completions or text embeddings.
//
// Create new instance by calling langmod.NewGemini
type Gemini struct {
	client          *genai.Client
	embeddingsModel string
	generativeModel string
	images          map[string][]byte // Mapping from external url to image bytes
}

// NewGemini is used to create Gemini client using specified generative and embeddings models.
// Returns pointer to client and an error.
func NewGemini(token, embeddingsModel, generativeModel string) (*Gemini, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(token))
	if err != nil {
		return nil, fmt.Errorf("error creating gemini client: %w", err)
	}

	return &Gemini{
		client:          client,
		embeddingsModel: embeddingsModel,
		generativeModel: generativeModel,
		images:          make(map[string][]byte),
	}, nil
}

// NewGeminiFromOption is used to create Gemini client with just config.ProviderOption.
// Returns pointer to client and an error.
func NewGeminiFromOption(token string, options config.ProviderOption) (*Gemini, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(token))
	if err != nil {
		return nil, fmt.Errorf("error creating gemini client: %w", err)
	}

	return &Gemini{
		client:          client,
		embeddingsModel: options.EmbeddingsModel,
		generativeModel: options.GenerativeModel,
		images:          make(map[string][]byte),
	}, nil
}

// message represents a single message with role and content
// It is used for arranging parts correctly in Gemini API call
type message struct {
	role    string
	content genai.Part
}

// CreateChatResponse completes the chat conversation with Gemini's chat completions API.
// Returns result string and an error.
func (m *Gemini) CreateChatResponse(args ...argument) (string, error) {
	model := m.client.GenerativeModel(m.generativeModel)

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	if args[0].getRole() == "system" {
		model.SystemInstruction = genai.NewUserContent(genai.Text(args[0].getText()))
		args = args[1:]
	}

	// Using message as temporary format for data makes it possible
	// to have images automatically appear first in Gemini API call
	var messages []message
	for _, arg := range args {
		switch arg.getType() {
		case "text":
			t := genai.Text(arg.getText())
			messages = append(messages, message{role: arg.getRole(), content: t})
		case "image":
			img, err := m.getImageData(arg.getImage())
			if err != nil {
				return "", fmt.Errorf("error getting image: %w", err)
			}
			format, err := detectImageFormat(img)
			if err != nil {
				return "", fmt.Errorf("error detecting image format: %w", err)
			}
			i := genai.ImageData(format, img)
			messages = insertImage(messages, message{role: arg.getRole(), content: i})
		}
	}

	prompt := make([]genai.Part, 0, len(messages))
	for _, msg := range messages {
		prompt = append(prompt, msg.content)
	}

	res, err := model.GenerateContent(context.Background(), prompt...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	result := string(res.Candidates[0].Content.Parts[0].(genai.Text))

	return result, nil
}

// getImageData downloads image from external source, saves it to local cache and returns image and an error.
func (m *Gemini) getImageData(url string) ([]byte, error) {
	var img []byte
	var exists bool
	if img, exists = m.images[url]; !exists {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error getting image: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error getting image: %s", resp.Status)
		}

		img, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading image: %w", err)
		}

		m.images[url] = img
	}

	return img, nil
}

// detectImageFormat returns images format and an error.
func detectImageFormat(img []byte) (format string, err error) {
	reader := bytes.NewReader(img)
	_, format, err = image.DecodeConfig(reader)

	// Function ignores jpeg feature error
	// See https://github.com/golang/go/issues/62421 for explanation
	// This should not affect the use in this case
	if err != nil && !errors.Is(err, jpeg.UnsupportedError("luma/chroma subsampling ratio")) {
		return "", err
	}
	return format, nil
}

// insertImage inserts given image to messages so that it is the first consecutive message from the same role
func insertImage(messages []message, img message) []message {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].role != img.role {
			// Insert img after position i
			return append(messages[:i+1], append([]message{img}, messages[i+1:]...)...)
		}
	}
	// If img is not preceded by other messages from the same role, img is appended to messages
	return append(messages, img)
}

// CreateEmbeddings creates text embeddings for input arguments.
// Currently only text embeddings are supported but images might be added in the future.
func (m *Gemini) CreateEmbeddings(args ...argument) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(args))

	model := m.client.EmbeddingModel(m.embeddingsModel)

	for _, arg := range args {
		if arg.getType() != "text" {
			return nil, fmt.Errorf("invalid argument type: %s", arg.getType())
		}

		res, err := model.EmbedContent(context.Background(), genai.Text(arg.getText()))
		if err != nil {
			return nil, fmt.Errorf("error creating embedding: %w", err)
		}

		emb := res.Embedding.Values
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}

// Close closes the Gemini client
func (m *Gemini) Close() error {
	return m.client.Close()
}

// GetEmbeddingDimensions returns the dimension count of the embeddings created by this language model
func (m *Gemini) GetEmbeddingDimensions() (int, error) {
	testString := "Text to be embedded"
	embeddings, err := m.CreateEmbeddings(NewTextInput(testString, "user"))
	if err != nil {
		return 0, err
	}

	return len(embeddings[0]), nil
}
