package langmod

// LangMod is a language model interface. It is used for creating text embeddings and chat completions.
// The interface makes switching from one language model to another easy.
type LangMod interface {
	CreateChatResponse(args ...argument) (string, error)
	CreateEmbeddings(args ...argument) ([][]float32, error)
	GetEmbeddingDimensions() (int, error)
	Close() error
}

type argument interface {
	getType() string
	getImage() string
	getText() string
	getRole() string
}

// TextInput implements argument interface and can be used as an input in language model method calls.
type TextInput struct {
	text      string
	role      string
	inputType string
}

func NewTextInput(text, role string) *TextInput {
	return &TextInput{text: text, role: role, inputType: "text"}
}

func (t *TextInput) getType() string {
	return t.inputType
}

func (t *TextInput) getText() string {
	return t.text
}

func (t *TextInput) getImage() string {
	return ""
}

func (t *TextInput) getRole() string {
	return t.role
}

// ImageInput implements argument interface and can be used as an input in language model method calls.
type ImageInput struct {
	image     string
	inputType string
	role      string
}

func NewImageInput(image, role string) *ImageInput {
	return &ImageInput{image: image, role: role, inputType: "image"}
}

func (i *ImageInput) getType() string {
	return i.inputType
}

func (i *ImageInput) getImage() string {
	return i.image
}

func (i *ImageInput) getText() string {
	return ""
}

func (i *ImageInput) getRole() string {
	return i.role
}
