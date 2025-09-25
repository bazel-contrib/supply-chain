package label

import (
	"fmt"
	"log"
)

// Label represents the [Label](https://bazel.build/rules/lib/builtins/Label) of a `Bazel` target.
type Label interface {
	fmt.Stringer

	// labelPrivate acts as marker to prevent other packages to implement the interface.
	//
	// See also https://medium.com/@johnsiilver/writing-an-interface-that-only-sub-packages-can-implement-fe36e7511449
	labelPrivate()
}

// Parse parses a `Label`.
func Parse(label string) (Label, error) {
	return &simple{label}, nil
}

// MustParse parses a `Label`.
func MustParse(label string) Label {
	l, err := Parse(label)
	if err != nil {
		log.Fatalf("Error parsing label %q: %v", label, err)
	}
	return l
}
