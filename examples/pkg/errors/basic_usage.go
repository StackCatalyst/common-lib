package main

import (
	"fmt"
	"log"

	"github.com/StackCatalyst/common-lib/pkg/errors"
)

func main() {
	// Example 1: Creating a new error
	if err := createUser(""); err != nil {
		if errors.Is(err, errors.ErrValidation) {
			log.Printf("Validation error: %v\n", err)
		} else {
			log.Printf("Unexpected error: %v\n", err)
		}
	}

	// Example 2: Wrapping an error
	if err := processUser("123"); err != nil {
		log.Printf("Processing error: %v\n", err)
	}
}

func createUser(username string) error {
	if username == "" {
		return errors.New(errors.ErrValidation, "username cannot be empty")
	}
	return nil
}

func processUser(userID string) error {
	// Simulate a database error
	dbErr := fmt.Errorf("connection refused")

	// Wrap the low-level error with more context
	return errors.Wrap(dbErr, errors.ErrInternal, "failed to process user")
}
