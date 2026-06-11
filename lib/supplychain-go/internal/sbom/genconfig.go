package sbom

type GenConfig struct {
	SchemaVersion string       `json:"schema_version"`
	RootTarget    string       `json:"root_target"`
	Nodes         []NodeConfig `json:"nodes"`
	Edges         []EdgeConfig `json:"edges"`
}

type NodeConfig struct {
	Label        string `json:"label"`
	MetadataFile string `json:"metadata_file"`
}

type EdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}
