package search

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add SQL and pgvector access for retrieval data.
