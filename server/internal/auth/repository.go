package auth

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add database access for student and administrator credentials.
