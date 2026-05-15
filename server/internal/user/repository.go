package user

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add database access for student accounts and profile data.
