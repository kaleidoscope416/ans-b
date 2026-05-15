package qa

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// TODO: add persistence for QA-specific records when needed.
