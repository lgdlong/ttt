package infrastructure

import (
	"context"
	"fmt"
	"os"

	"github.com/pgvector/pgvector-go"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient wraps OpenAI API client with embedding functionality
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client instance
// Sử dụng OPENAI_BASE_URL và OPENAI_API_KEY từ environment
func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1" // Default OpenAI endpoint
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
	}, nil
}

// GetEmbedding converts text to 1536-dimensional vector using text-embedding-3-small
// Uses small model which fits PostgreSQL pgvector natively without truncation
// Returns pgvector.Vector ready to store in PostgreSQL
func (c *OpenAIClient) GetEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	// Validate input
	if text == "" {
		return pgvector.Vector{}, fmt.Errorf("text cannot be empty")
	}

	// Call OpenAI Embedding API
	resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3, // text-embedding-3-small (1536 dims, good balance of speed/quality)
	})

	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("failed to create embedding: %w", err)
	}

	// Validate response
	if len(resp.Data) == 0 {
		return pgvector.Vector{}, fmt.Errorf("no embedding data returned")
	}

	// No truncation needed - 1536 dims fits within PostgreSQL pgvector max
	return pgvector.NewVector(resp.Data[0].Embedding), nil
}

// BatchGetEmbeddings gets embeddings for multiple texts in one API call
// More efficient when creating multiple tags at once
// Uses text-embedding-3-small which fits natively without truncation
func (c *OpenAIClient) BatchGetEmbeddings(ctx context.Context, texts []string) ([]pgvector.Vector, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: openai.SmallEmbedding3, // text-embedding-3-small (1536 dims, faster/cheaper)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	vectors := make([]pgvector.Vector, len(resp.Data))
	for i, data := range resp.Data {
		// No truncation needed - small model outputs 1536 dims
		vectors[i] = pgvector.NewVector(data.Embedding)
	}

	return vectors, nil
}

// TranslateToEnglish translates text to English using GPT-4o-mini
// Used in Translation Layer to normalize cross-lingual queries (e.g., "Tiền" -> "Money")
// Cost: ~$0.15 per 1M tokens (extremely cheap)
// Latency: ~200-500ms per call
func (c *OpenAIClient) TranslateToEnglish(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", nil
	}

	const TRANS_ENG_TO_VIE_SYS_PROMPT = "You are an expert terminologist. Translate the Vietnamese input into its most standard, professional, and academic English equivalent. \nRules:\n1. Return ONLY the English term.\n2. Prioritize established terminology (e.g., 'Pragmatism' instead of 'Practicalism').\n3. Preserve proper nouns.\n4. Do not add punctuation or explanations."

	// Use gpt-4o-mini: Fast, Cheap, Smart enough for simple translation
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: TRANS_ENG_TO_VIE_SYS_PROMPT,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
			MaxTokens:   20,  // Limit tokens to save cost and avoid verbosity
			Temperature: 0.0, // Deterministic: Always return the same result
		},
	)

	if err != nil {
		return "", fmt.Errorf("translation error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	return resp.Choices[0].Message.Content, nil
}
