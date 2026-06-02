package storage

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add persistence for uploaded file metadata.
