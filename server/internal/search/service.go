package search

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add question cleaning, intent recognition, retrieval, and ranking.
