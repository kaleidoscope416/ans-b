package model

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add model usage persistence if audit or billing data is required.
