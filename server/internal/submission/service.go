package submission

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// TODO: add submission workflow, audit rules, and knowledge publishing integration.
