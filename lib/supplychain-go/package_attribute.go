package supplychain

import (
	"io"
)

// PackageAttributeDescriptor describes an attribute of `PackageMetadata`.
type PackageAttributeDescriptor[T any] struct {
	// Kind is the identifier of the attribute.
	Kind string

	// Parser is a parser for the attribute.
	Parser func(r io.Reader) (*T, error)
}
