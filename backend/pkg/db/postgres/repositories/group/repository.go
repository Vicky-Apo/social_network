package group

import "database/sql"

// Repository implements the group repository using Postgres.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a new Postgres group repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
