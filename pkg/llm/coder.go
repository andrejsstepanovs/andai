package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/teilomillet/gollm"
)

func (l *LLM) Simple(input string) (string, error) {
	ctx := context.Background()

	basicPrompt := gollm.NewPrompt(input)
	fmt.Printf("Basic prompt created: %+v\n", basicPrompt)

	response, err := l.Coder.Generate(ctx, basicPrompt)
	if err != nil {
		log.Fatalf("Failed to generate text: %v", err)
		return "", err
	}
	fmt.Printf("Basic prompt response:\n%s\n", response)

	return response, nil
}
