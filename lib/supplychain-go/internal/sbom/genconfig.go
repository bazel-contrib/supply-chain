package sbom

type GenConfig struct {
	Deps []DepConfig `json:"deps"`
}

type DepConfig struct {
	Metadata string `json:"metadata"`
}
