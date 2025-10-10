package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"semki/pkg/telemetry"
	"time"
)

type IEmbedderService interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
	EmbedBatchWithIDs(texts []TextWithID) ([]EmbeddingWithID, error)
}

type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

type EmbeddingWithIDResponse struct {
	Embeddings []struct {
		ID        string    `json:"id"`
		Embedding []float32 `json:"embedding"`
	} `json:"embeddings"`
}

type TextWithID struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// EmbeddingWithID represents embedding result with ID
type EmbeddingWithID struct {
	ID        string
	Embedding []float32
}

// embedderService handles text embedding operations
type embedderService struct {
	embedderURL string
	httpClient  *http.Client
}

// NewEmbedderService creates a new embedder service instance
func NewEmbedderService(embedderURL string) IEmbedderService {
	return &embedderService{
		embedderURL: embedderURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Default timeout, will be adjusted per request
		},
	}
}

// Embed embeds a single text
func (s *embedderService) Embed(text string) ([]float32, error) {
	embeddings, err := s.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// EmbedBatch embeds multiple texts
func (s *embedderService) EmbedBatch(texts []string) ([][]float32, error) {
	telemetry.Log.Info(fmt.Sprintf("working with texts %d", len(texts)))

	// Prepare request body
	requestBody := map[string][]string{
		"texts": texts,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with dynamic timeout
	req, err := http.NewRequest("POST", s.embedderURL+"/embed", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Create client with dynamic timeout based on text count
	client := &http.Client{
		Timeout: time.Duration(len(texts)*5) * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var response EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	telemetry.Log.Info(fmt.Sprintf("✅ embeddings done! %d", len(response.Embeddings)))

	return response.Embeddings, nil
}

// EmbedBatchWithIDs embeds multiple texts with their IDs
func (s *embedderService) EmbedBatchWithIDs(texts []TextWithID) ([]EmbeddingWithID, error) {
	telemetry.Log.Info(fmt.Sprintf("working with texts %d", len(texts)))

	// Prepare request body
	requestBody := map[string][]TextWithID{
		"texts": texts,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with dynamic timeout
	req, err := http.NewRequest("POST", s.embedderURL+"/embed_with_ids", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Create client with dynamic timeout based on text count
	client := &http.Client{
		Timeout: time.Duration(len(texts)*5) * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var response EmbeddingWithIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to result format
	result := make([]EmbeddingWithID, len(response.Embeddings))
	for i, item := range response.Embeddings {
		result[i] = EmbeddingWithID{
			ID:        item.ID,
			Embedding: item.Embedding,
		}
	}

	telemetry.Log.Info(fmt.Sprintf("✅ embeddings done! %d", len(result)))

	return result, nil
}
