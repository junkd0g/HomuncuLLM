#!/bin/sh

# Start Ollama in the background
ollama serve &

# Wait for Ollama to start
echo "Waiting for Ollama to start..."
until $(curl --output /dev/null --silent --head --fail http://localhost:11434/api/tags); do
  printf '.'
  sleep 1
done
echo "Ollama started!"

# Pull the default model (this only happens once)
DEFAULT_MODEL=${DEFAULT_MODEL:-"llama2"}
echo "Pulling model: $DEFAULT_MODEL"
ollama pull $DEFAULT_MODEL

# Start the Go service
echo "Starting Go LLM service..."
./llm-service
