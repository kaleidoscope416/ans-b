package model

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add provider configuration and model request orchestration.
