package langmod

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integration test for CreateChatResponse using the real OpenAI API
func TestCreateChatResponse_OpenAI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_TOKEN")
	if apiKey == "" {
		t.Error("OPENAI_TOKEN environment variable not set")
	}

	// Initialize OpenAI instance
	langMod, err := NewOpenAI(apiKey, "text-embedding-3-small", "gpt-4o-mini")
	assert.NoError(t, err)

	args := []argument{
		NewTextInput("You help users describe objects based on text descriptions and images.", "system"),
		NewTextInput("This is an image I found online", "user"),
		NewImageInput("https://picsum.photos/200", "user"),
	}

	resp, err := langMod.CreateChatResponse(args...)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

// Integration test for CreateChatResponse using the real Gemini API
func TestCreateChatResponse_Gemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_TOKEN")
	if apiKey == "" {
		t.Error("GEMINI_TOKEN environment variable not set")
	}

	// Initialize Gemini instance
	langMod, err := NewGemini(apiKey, "text-embedding-004", "gemini-1.5-flash")
	assert.NoError(t, err)

	args := []argument{
		NewTextInput("You help users describe objects based on text descriptions and images.", "system"),
		NewTextInput("This is an image I found online", "user"),
		NewImageInput("https://picsum.photos/200", "user"),
	}

	resp, err := langMod.CreateChatResponse(args...)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp)
}

// Integration test for CreateEmbeddings using the real API
func TestCreateEmbeddings_OpenAI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_TOKEN")
	if apiKey == "" {
		t.Error("OPENAI_TOKEN environment variable not set")
	}

	// Initialize OpenAI instance
	langMod, err := NewOpenAI(apiKey, "text-embedding-3-small", "gpt-4o-mini")
	assert.NoError(t, err)

	args := []argument{
		NewTextInput("Hello, world!", ""),
		NewTextInput("This is some lorem ipsum...", ""),
	}

	embeddings, err := langMod.CreateEmbeddings(args...)

	assert.NoError(t, err)
	assert.NotNil(t, embeddings)
	assert.NotEmpty(t, embeddings[0])
	assert.Equal(t, len(embeddings), len(args))
}

// Integration test for CreateEmbeddings using the real API
func TestCreateEmbeddings_Gemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_TOKEN")
	if apiKey == "" {
		t.Error("GEMINI_TOKEN environment variable not set")
	}

	// Initialize Gemini instance
	langMod, err := NewGemini(apiKey, "text-embedding-004", "gemini-1.5-flash")
	assert.NoError(t, err)

	args := []argument{
		NewTextInput("Hello, world!", ""),
		NewTextInput("This is some lorem ipsum...", ""),
	}

	embeddings, err := langMod.CreateEmbeddings(args...)

	assert.NoError(t, err)
	assert.NotNil(t, embeddings)
	assert.NotEmpty(t, embeddings[0])
	assert.Equal(t, len(embeddings), len(args))
}

// Test for handling invalid input types in CreateEmbeddings
func TestCreateEmbeddings_OpenAI_InvalidInput(t *testing.T) {
	apiKey := os.Getenv("OPENAI_TOKEN")
	if apiKey == "" {
		t.Error("OPENAI_TOKEN environment variable not set")
	}

	// Initialize OpenAI instance
	langMod, err := NewOpenAI(apiKey, "text-embedding-3-small", "gpt-4o-mini")
	assert.NoError(t, err)

	args := []argument{
		NewImageInput("fake_image_data", "user"), // Passing an image input where text is expected
	}

	embeddings, err := langMod.CreateEmbeddings(args...)

	assert.Error(t, err)
	assert.Nil(t, embeddings)
}

// Test for handling invalid input types in CreateEmbeddings
func TestCreateEmbeddings_Gemini_InvalidInput(t *testing.T) {
	apiKey := os.Getenv("GEMINI_TOKEN")
	if apiKey == "" {
		t.Error("GEMINI_TOKEN environment variable not set")
	}

	// Initialize Gemini instance
	langMod, err := NewGemini(apiKey, "text-embedding-004", "gemini-1.5-flash")
	assert.NoError(t, err)

	args := []argument{
		NewImageInput("fake_image_data", "user"), // Passing an image input where text is expected
	}

	embeddings, err := langMod.CreateEmbeddings(args...)

	assert.Error(t, err)
	assert.Nil(t, embeddings)
}
