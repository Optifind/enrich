package prompts

import (
	"encoding/json"
	"fmt"
	"github.com/lattots/openai-sdk"
	"io"
	"os"
)

// Prompts is a struct for holding prompts used in chat completion creation.
// Prompts are loaded from JSON at runtime, so they can be changed without compiling the entire program.
type Prompts struct {
	StyleText   []openaisdk.Message `json:"style-text"`
	UseCaseText []openaisdk.Message `json:"use-case-text"`
	Attributes  []openaisdk.Message `json:"attributes"`
}

// Load loads prompts from JSON. Returns Prompts object and an error.
func Load(filepath string) (Prompts, error) {
	fmt.Println("Loading prompts file...")
	// Prompts JSON file is opened to read.
	file, err := os.Open(filepath)
	if err != nil {
		return Prompts{}, err
	}

	// JSON files content is read to byte code.
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return Prompts{}, err
	}

	// Prompts json is read unmarshalled to prompts struct.
	var prompts Prompts
	err = json.Unmarshal(fileContent, &prompts)
	if err != nil {
		return Prompts{}, err
	}

	// Prompts and nil error is returned.
	return prompts, nil
}
