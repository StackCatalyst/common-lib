package testing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockProvider is a mock implementation of Provider for testing
type MockProvider struct {
	mu        sync.RWMutex
	resources map[string]*Resource
}

// NewMockProvider creates a new mock provider
func NewMockProvider() Provider {
	return &MockProvider{
		resources: make(map[string]*Resource),
	}
}

// CreateResource simulates creating a cloud resource
func (p *MockProvider) CreateResource(ctx context.Context, resource *Resource) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if resource.ID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if _, exists := p.resources[resource.ID]; exists {
		return fmt.Errorf("resource %s already exists", resource.ID)
	}

	if err := p.ValidateResource(resource); err != nil {
		return err
	}

	resource.CreatedAt = time.Now()
	p.resources[resource.ID] = resource

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	return nil
}

// DeleteResource simulates deleting a cloud resource
func (p *MockProvider) DeleteResource(ctx context.Context, resource *Resource) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if resource.ID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if _, exists := p.resources[resource.ID]; !exists {
		return fmt.Errorf("resource %s not found", resource.ID)
	}

	delete(p.resources, resource.ID)

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetResource simulates getting information about a cloud resource
func (p *MockProvider) GetResource(ctx context.Context, id string) (*Resource, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if id == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	resource, exists := p.resources[id]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", id)
	}

	// Simulate API delay
	time.Sleep(50 * time.Millisecond)

	return resource, nil
}

// ListResources simulates listing cloud resources
func (p *MockProvider) ListResources(ctx context.Context, filter map[string]string) ([]*Resource, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var resources []*Resource

	for _, resource := range p.resources {
		matches := true
		for k, v := range filter {
			if tagValue, ok := resource.Tags[k]; !ok || tagValue != v {
				matches = false
				break
			}
		}
		if matches {
			resources = append(resources, resource)
		}
	}

	// Simulate API delay
	time.Sleep(50 * time.Millisecond)

	return resources, nil
}

// ValidateResource validates a resource definition
func (p *MockProvider) ValidateResource(resource *Resource) error {
	if resource.Type == "" {
		return fmt.Errorf("resource type is required")
	}

	if resource.Provider == "" {
		return fmt.Errorf("resource provider is required")
	}

	if resource.Region == "" {
		return fmt.Errorf("resource region is required")
	}

	return nil
}
