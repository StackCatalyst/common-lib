package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/database"
	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/StackCatalyst/common-lib/pkg/module/storage"
	"github.com/jackc/pgx/v5"
)

// Storage implements the storage.Storage interface using PostgreSQL
type Storage struct {
	db      *database.Client
	metrics *metrics.Reporter
}

// Config represents PostgreSQL storage configuration
type Config struct {
	DBConfig      database.Config
	MetricsPrefix string
}

// New creates a new PostgreSQL storage instance
func New(config Config, metrics *metrics.Reporter) (*Storage, error) {
	db, err := database.New(config.DBConfig, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
	}

	return &Storage{
		db:      db,
		metrics: metrics,
	}, nil
}

// Store saves a module to PostgreSQL
func (s *Storage) Store(ctx context.Context, module *module.Module) error {
	query := `
		INSERT INTO modules (
			id, name, provider, version, description, source,
			variables, outputs, dependencies, tags,
			created_at, updated_at, metadata, content
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (id, version) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			source = EXCLUDED.source,
			variables = EXCLUDED.variables,
			outputs = EXCLUDED.outputs,
			dependencies = EXCLUDED.dependencies,
			tags = EXCLUDED.tags,
			updated_at = EXCLUDED.updated_at,
			metadata = EXCLUDED.metadata,
			content = EXCLUDED.content
		WHERE NOT modules.locked
	`

	variables, err := json.Marshal(module.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	outputs, err := json.Marshal(module.Outputs)
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}

	dependencies, err := json.Marshal(module.Dependencies)
	if err != nil {
		return fmt.Errorf("failed to marshal dependencies: %w", err)
	}

	_, err = s.db.Exec(ctx, query,
		module.ID,
		module.Name,
		module.Provider,
		module.Version,
		module.Description,
		module.Source,
		variables,
		outputs,
		dependencies,
		module.Tags,
		module.CreatedAt,
		module.UpdatedAt,
		module.Metadata,
		nil, // content is stored separately
	)

	if err != nil {
		return fmt.Errorf("failed to store module: %w", err)
	}

	return nil
}

// Get retrieves a module by its ID and version
func (s *Storage) Get(ctx context.Context, id, version string) (*module.Module, error) {
	query := `
		SELECT
			id, name, provider, version, description, source,
			variables, outputs, dependencies, tags,
			created_at, updated_at, metadata
		FROM modules
		WHERE id = $1 AND version = $2
	`

	row := s.db.QueryRow(ctx, query, id, version)
	module := &module.Module{}

	var variables, outputs, dependencies []byte

	err := row.Scan(
		&module.ID,
		&module.Name,
		&module.Provider,
		&module.Version,
		&module.Description,
		&module.Source,
		&variables,
		&outputs,
		&dependencies,
		&module.Tags,
		&module.CreatedAt,
		&module.UpdatedAt,
		&module.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("module not found: %s@%s", id, version)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan module: %w", err)
	}

	if err := json.Unmarshal(variables, &module.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	if err := json.Unmarshal(outputs, &module.Outputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	if err := json.Unmarshal(dependencies, &module.Dependencies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
	}

	return module, nil
}

// List returns modules matching the given filter
func (s *Storage) List(ctx context.Context, filter storage.Filter) ([]*module.Module, error) {
	query := `
		SELECT
			id, name, provider, version, description, source,
			variables, outputs, dependencies, tags,
			created_at, updated_at, metadata
		FROM modules
		WHERE ($1::text IS NULL OR provider = $1)
		AND ($2::text[] IS NULL OR tags && $2)
		AND ($3::text IS NULL OR name LIKE $3)
		AND ($4::text IS NULL OR version = $4)
		ORDER BY created_at DESC
		OFFSET $5 LIMIT $6
	`

	rows, err := s.db.Query(ctx, query,
		filter.Provider,
		filter.Tags,
		filter.NamePattern,
		filter.Version,
		filter.Offset,
		filter.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modules []*module.Module

	for rows.Next() {
		module := &module.Module{}
		var variables, outputs, dependencies []byte

		err := rows.Scan(
			&module.ID,
			&module.Name,
			&module.Provider,
			&module.Version,
			&module.Description,
			&module.Source,
			&variables,
			&outputs,
			&dependencies,
			&module.Tags,
			&module.CreatedAt,
			&module.UpdatedAt,
			&module.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}

		if err := json.Unmarshal(variables, &module.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		if err := json.Unmarshal(outputs, &module.Outputs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
		}

		if err := json.Unmarshal(dependencies, &module.Dependencies); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
		}

		modules = append(modules, module)
	}

	return modules, nil
}

// Delete removes a module from storage
func (s *Storage) Delete(ctx context.Context, id, version string) error {
	query := `DELETE FROM modules WHERE id = $1 AND version = $2 AND NOT locked`
	result, err := s.db.Exec(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found or locked: %s@%s", id, version)
	}

	return nil
}

// GetVersions returns all versions of a module
func (s *Storage) GetVersions(ctx context.Context, id string) ([]string, error) {
	query := `SELECT version FROM modules WHERE id = $1 ORDER BY version DESC`
	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetLatestVersion returns the latest version of a module
func (s *Storage) GetLatestVersion(ctx context.Context, id string) (string, error) {
	query := `SELECT version FROM modules WHERE id = $1 ORDER BY version DESC LIMIT 1`
	var version string
	err := s.db.QueryRow(ctx, query, id).Scan(&version)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("module not found: %s", id)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get latest version: %w", err)
	}
	return version, nil
}

// Lock marks a version as immutable
func (s *Storage) Lock(ctx context.Context, id, version string) error {
	query := `UPDATE modules SET locked = true WHERE id = $1 AND version = $2`
	result, err := s.db.Exec(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to lock module: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found: %s@%s", id, version)
	}

	return nil
}

// GetMetadata retrieves module metadata without content
func (s *Storage) GetMetadata(ctx context.Context, id, version string) (*module.Module, error) {
	return s.Get(ctx, id, version)
}

// UpdateMetadata updates module metadata without changing content
func (s *Storage) UpdateMetadata(ctx context.Context, id, version string, metadata map[string]interface{}) error {
	query := `
		UPDATE modules
		SET metadata = $3, updated_at = $4
		WHERE id = $1 AND version = $2 AND NOT locked
	`
	result, err := s.db.Exec(ctx, query, id, version, metadata, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found or locked: %s@%s", id, version)
	}

	return nil
}

// StoreContent saves module content to storage
func (s *Storage) StoreContent(ctx context.Context, id, version string, content []byte) error {
	query := `
		UPDATE modules
		SET content = $3, updated_at = $4
		WHERE id = $1 AND version = $2 AND NOT locked
	`
	result, err := s.db.Exec(ctx, query, id, version, content, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store content: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found or locked: %s@%s", id, version)
	}

	return nil
}

// GetContent retrieves module content from storage
func (s *Storage) GetContent(ctx context.Context, id, version string) ([]byte, error) {
	query := `SELECT content FROM modules WHERE id = $1 AND version = $2`
	var content []byte
	err := s.db.QueryRow(ctx, query, id, version).Scan(&content)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("module not found: %s@%s", id, version)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	return content, nil
}

// Exists checks if a module version exists
func (s *Storage) Exists(ctx context.Context, id, version string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM modules WHERE id = $1 AND version = $2)`
	var exists bool
	err := s.db.QueryRow(ctx, query, id, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}

// GetDependencies returns all modules that depend on the given module
func (s *Storage) GetDependencies(ctx context.Context, id, version string) ([]*module.Module, error) {
	query := `
		SELECT
			id, name, provider, version, description, source,
			variables, outputs, dependencies, tags,
			created_at, updated_at, metadata
		FROM modules
		WHERE dependencies @> $1
	`
	dependency := fmt.Sprintf(`[{"source": "%s", "version": "%s"}]`, id, version)

	rows, err := s.db.Query(ctx, query, dependency)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	var modules []*module.Module
	for rows.Next() {
		module := &module.Module{}
		var variables, outputs, dependencies []byte

		err := rows.Scan(
			&module.ID,
			&module.Name,
			&module.Provider,
			&module.Version,
			&module.Description,
			&module.Source,
			&variables,
			&outputs,
			&dependencies,
			&module.Tags,
			&module.CreatedAt,
			&module.UpdatedAt,
			&module.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}

		if err := json.Unmarshal(variables, &module.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		if err := json.Unmarshal(outputs, &module.Outputs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
		}

		if err := json.Unmarshal(dependencies, &module.Dependencies); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
		}

		modules = append(modules, module)
	}

	return modules, nil
}

// Close releases any resources held by the storage
func (s *Storage) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}
