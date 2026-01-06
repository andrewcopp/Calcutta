package model_catalogs

type ModelDescriptor struct {
	ID            string `json:"id"`
	DisplayName   string `json:"display_name"`
	SchemaVersion string `json:"schema_version"`
	Deprecated    bool   `json:"deprecated"`
}

var AvailableAdvancementModels = []ModelDescriptor{
	{ID: "kenpom-v1-go", DisplayName: "KenPom V1 (Go)", SchemaVersion: "v1", Deprecated: false},
}

var AvailableMarketShareModels = []ModelDescriptor{
	{ID: "ridge", DisplayName: "Ridge", SchemaVersion: "v1", Deprecated: false},
	{ID: "naive-ev-baseline", DisplayName: "Naive EV Baseline", SchemaVersion: "v1", Deprecated: false},
}

var AvailableEntryOptimizers = []ModelDescriptor{
	{ID: "minlp_v1", DisplayName: "MINLP V1", SchemaVersion: "v1", Deprecated: false},
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) ListAdvancementModels() []ModelDescriptor {
	out := make([]ModelDescriptor, len(AvailableAdvancementModels))
	copy(out, AvailableAdvancementModels)
	return out
}

func (s *Service) ListMarketShareModels() []ModelDescriptor {
	out := make([]ModelDescriptor, len(AvailableMarketShareModels))
	copy(out, AvailableMarketShareModels)
	return out
}

func (s *Service) ListEntryOptimizers() []ModelDescriptor {
	out := make([]ModelDescriptor, len(AvailableEntryOptimizers))
	copy(out, AvailableEntryOptimizers)
	return out
}
