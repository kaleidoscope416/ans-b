package auth

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add password hashing, token generation, and auth policy methods.
