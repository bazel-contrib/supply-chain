package supplychain

// Dependency represents a directed edge in the target graph: a dependency
// relationship from one target onto another.
//
// It pairs the dependent target's `TargetMetadata` with the scope under which
// the dependency exists (e.g., a tool dependency used during the build versus
// a runtime dependency linked into the target). See `TargetMetadata.GetDependencies`
// for how `Dependency` values are obtained.
type Dependency interface {
	// dependencyPrivate acts as marker to prevent other packages to implement the interface.
	//
	// See also https://medium.com/@johnsiilver/writing-an-interface-that-only-sub-packages-can-implement-fe36e7511449
	dependencyPrivate()

	// GetScope returns the scope describing the nature of this dependency relationship,
	// such as whether the dependent target is a build tool, a linked runtime dependency,
	// a dynamic dependency, or a bundled dependency.
	GetScope() dependencyScope

	// GetTargetMetadata returns the `TargetMetadata` of the dependent target.
	GetTargetMetadata() TargetMetadata
}

type dependency struct {
	Scope          dependencyScope
	TargetMetadata TargetMetadata
}

func (d *dependency) dependencyPrivate() {}

func (d *dependency) GetScope() dependencyScope {
	return d.Scope
}

func (d *dependency) GetTargetMetadata() TargetMetadata {
	return d.TargetMetadata
}

type rawDependency struct {
	Scope string `json:"scope"`
	Path  string `json:"path"`
}
