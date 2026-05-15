package analytics

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add database access for query logs and analytics aggregates.
