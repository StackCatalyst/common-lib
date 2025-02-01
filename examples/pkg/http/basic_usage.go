package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/http"
	"github.com/StackCatalyst/common-lib/pkg/metrics"
)

type Post struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID int    `json:"userId"`
}

func main() {
	// Example 1: Client setup with custom configuration
	config := http.DefaultConfig()
	config.Timeout = 10 * time.Second
	config.MaxRetries = 3
	config.RetryWaitMin = 1 * time.Second
	config.RetryWaitMax = 5 * time.Second
	config.RetryableStatusCodes = []int{408, 429, 500, 502, 503, 504}

	metricsReporter := metrics.New(metrics.DefaultOptions())
	client := http.New(config, metricsReporter)

	ctx := context.Background()

	// Example 2: GET request
	resp, err := client.Get(ctx, "https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		log.Printf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		log.Printf("Failed to decode response: %v", err)
	}
	fmt.Printf("Retrieved post: %+v\n", post)

	// Example 3: POST request with JSON body
	newPost := Post{
		Title:  "New Post",
		Body:   "This is a new post",
		UserID: 1,
	}
	postBody, err := json.Marshal(newPost)
	if err != nil {
		log.Printf("Failed to marshal post: %v", err)
	}

	resp, err = client.Post(ctx, "https://jsonplaceholder.typicode.com/posts",
		"application/json", strings.NewReader(string(postBody)))
	if err != nil {
		log.Printf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	var createdPost Post
	if err := json.NewDecoder(resp.Body).Decode(&createdPost); err != nil {
		log.Printf("Failed to decode response: %v", err)
	}
	fmt.Printf("Created post: %+v\n", createdPost)

	// Example 4: Custom request with headers
	req, err := stdhttp.NewRequest("PUT", "https://jsonplaceholder.typicode.com/posts/1", strings.NewReader(string(postBody)))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	resp, err = client.Do(req)
	if err != nil {
		log.Printf("PUT request failed: %v", err)
	}
	defer resp.Body.Close()

	var updatedPost Post
	if err := json.NewDecoder(resp.Body).Decode(&updatedPost); err != nil {
		log.Printf("Failed to decode response: %v", err)
	}
	fmt.Printf("Updated post: %+v\n", updatedPost)

	// Example 5: Request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err = client.Get(ctx, "https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		log.Printf("Request with timeout failed: %v", err)
	}
	defer resp.Body.Close()

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		log.Printf("Failed to decode response: %v", err)
	}
	fmt.Printf("Retrieved %d posts\n", len(posts))
}
