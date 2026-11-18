package main

import (
	"testing"

	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

func TestGeneratorTool(t *testing.T) {
	want := "pkg:bazel/supply-chain-go@1.2.3"
	released := generatorTool(want)
	if len(released.ExternalIdentifiers) != 1 {
		t.Fatalf("released tool has %d external identifiers, want 1", len(released.ExternalIdentifiers))
	}
	if got := released.ExternalIdentifiers[0].(interface{ GetIdentifier() string }).GetIdentifier(); got != want {
		t.Errorf("purl = %q, want %q", got, want)
	}

	if got := generatorTool("").ExternalIdentifiers; len(got) != 0 {
		t.Errorf("tool without a purl has identifiers %v, want none", got)
	}
}

func TestOriginatedBy(t *testing.T) {
	got := originatedBy("github", "acmecorp")
	if len(got) != 1 || got[0].(*spdx.Organization).Name != "acmecorp" {
		t.Errorf("originatedBy(github, acmecorp) = %v, want organization acmecorp", got)
	}
	for _, tc := range [][2]string{{"github", ""}, {"maven", "org.acme"}, {"bazel", ""}} {
		if got := originatedBy(tc[0], tc[1]); got != nil {
			t.Errorf("originatedBy(%q, %q) = %v, want none", tc[0], tc[1], got)
		}
	}
}
