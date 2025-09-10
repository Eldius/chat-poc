package docs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	_ tools.Tool = &Search{}
)

type Search struct {
	s vectorstores.VectorStore
}

func NewSearch(s vectorstores.VectorStore) *Search {
	return &Search{s: s}
}

func (s Search) Name() string {
	return "documentation_search"
}

func (s Search) Description() string {
	return "Useful for searching internal company documentation and some market definitions (like ABECS deny rules and retry rules configurations). The input should be a keyword or phrase describing the document content needed."
}

func (s Search) Call(ctx context.Context, input string) (string, error) {
	docs, err := s.s.SimilaritySearch(ctx, input, 100)
	if err != nil {
		return "", err
	}

	var output string
	for _, doc := range docs {
		b, _ := json.Marshal(doc)
		output += fmt.Sprintf("%s\n", b)
	}

	return output, nil
}
