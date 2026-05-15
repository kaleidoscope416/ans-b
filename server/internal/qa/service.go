package qa

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add QA orchestration across search, analytics, and optional model modules.
