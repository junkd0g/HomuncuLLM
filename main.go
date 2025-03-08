package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
}

// PromptRequest is our API's request structure
type PromptRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Model  string `json:"model"`
}

// PromptResponse is our API's response structure
type PromptResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
	Time     string `json:"time"`
}

// LLMService handles communication with the Ollama service
type LLMService struct {
	ollamaURL  string
	httpClient *http.Client
	defaultModel string
}

// NewLLMService creates a new service
func NewLLMService(ollamaURL string, defaultModel string) *LLMService {
	return &LLMService{
		ollamaURL:  ollamaURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		defaultModel: defaultModel,
	}
}

// GetCompletion sends a prompt to Ollama and returns the response
func (s *LLMService) GetCompletion(prompt string, model string) (string, error) {
	if model == "" {
		model = s.defaultModel
	}

	reqBody, err := json.Marshal(OllamaRequest{
		Model:  model,
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.httpClient.Post(s.ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

func main() {
	// Get configuration from environment variables
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	
	defaultModel := os.Getenv("DEFAULT_MODEL")
	if defaultModel == "" {
		defaultModel = "llama2"
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create LLM service
	llmService := NewLLMService(ollamaURL, defaultModel)

	// Setup Gin router
	router := gin.Default()
	
	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Define endpoint for prompt completion
	router.POST("/api/complete", func(c *gin.Context) {
		var req PromptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		startTime := time.Now()
		response, err := llmService.GetCompletion(req.Prompt, req.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, PromptResponse{
			Response: response,
			Model:    req.Model,
			Time:     time.Since(startTime).String(),
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Start the server
	log.Printf("Starting server on port %s", port)
	log.Printf("Using Ollama at %s with default model %s", ollamaURL, defaultModel)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
