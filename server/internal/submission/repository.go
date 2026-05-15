package submission

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add database access for submissions and audit history.
