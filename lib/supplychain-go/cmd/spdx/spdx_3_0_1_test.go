package main

import (
	"testing"

	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

func TestGeneratorTool(t *testing.T) {
	released := generatorTool("1.2.3")
	if len(released.ExternalIdentifiers) != 1 {
		t.Fatalf("released tool has %d external identifiers, want 1", len(released.ExternalIdentifiers))
	}
	want := "pkg:github/bazel-contrib/supply-chain@v1.2.3"
	if got := released.ExternalIdentifiers[0].(interface{ GetIdentifier() string }).GetIdentifier(); got != want {
		t.Errorf("purl = %q, want %q", got, want)
	}

	for _, version := range []string{"", "HEAD"} {
		if got := generatorTool(version).ExternalIdentifiers; len(got) != 0 {
			t.Errorf("generatorTool(%q) has identifiers %v, want none", version, got)
		}
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
