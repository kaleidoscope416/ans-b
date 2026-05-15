package knowledge

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add knowledge validation, import orchestration, and search-index updates.
