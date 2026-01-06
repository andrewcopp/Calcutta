package model_catalogs

type ModelDescriptor struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Deprecated  bool   `json:"deprecated"`
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) ListAdvancementModels() []ModelDescriptor {
	return []ModelDescriptor{
		{ID: "kenpom-v1-go", DisplayName: "KenPom V1 (Go)", Deprecated: false},
	}
}

func (s *Service) ListMarketShareModels() []ModelDescriptor {
	return []ModelDescriptor{
		{ID: "ridge", DisplayName: "Ridge", Deprecated: false},
		{ID: "naive-ev-baseline", DisplayName: "Naive EV Baseline", Deprecated: false},
	}
}

func (s *Service) ListEntryOptimizers() []ModelDescriptor {
	return []ModelDescriptor{
		{ID: "minlp_v1", DisplayName: "MINLP V1", Deprecated: false},
	}
}
