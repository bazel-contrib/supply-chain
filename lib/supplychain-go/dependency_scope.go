package supplychain

// dependencyScope describes the nature of a dependency relationship between
// two targets in the target graph. See the `DependencyScope*` constants for
// the set of recognized scopes.
type dependencyScope string

const (
	// DependencyScopeTool denotes a tool dependency: the dependent target was
	// used to produce the depending target (e.g., a compiler or other toolchain
	// dependency) but is not itself part of the resulting artifact.
	DependencyScopeTool dependencyScope = "tool"

	// DependencyScopeRuntime denotes a runtime dependency that is linked into
	// the depending target (e.g., a library listed in the target's `deps`
	// attribute) and is required for the target to execute.
	DependencyScopeRuntime dependencyScope = "runtime"

	// DependencyScopeDynamic denotes a dynamic dependency that is not linked
	// into the depending target but is required at runtime (e.g., a system-wide
	// shared library resolved by the dynamic linker).
	DependencyScopeDynamic dependencyScope = "dynamic"

	// DependencyScopeBundled denotes a bundled dependency that is included
	// inside the depending target as a whole (e.g., another target embedded
	// in an archive or container image).
	DependencyScopeBundled dependencyScope = "bundled"
)
