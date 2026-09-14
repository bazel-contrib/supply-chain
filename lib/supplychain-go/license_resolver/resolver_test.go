package license_resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyDirectoryMIT(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	license := `MIT License

Copyright (c) 2026 Example

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
	if err := os.WriteFile(filepath.Join(root, "LICENSE"), []byte(license), 0o644); err != nil {
		t.Fatal(err)
	}

	classifier, err := NewClassifier(Options{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := classifier.ClassifyDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.LicenseExpression, "MIT"; got != want {
		t.Fatalf("got expression %q, want %q; evidence: %+v", got, want, result.Evidence)
	}
	if got, want := len(result.Evidence), 1; got != want {
		t.Fatalf("got %d evidence records, want %d", got, want)
	}
}

func TestConservativeExpressionRequiresCompleteEvidence(t *testing.T) {
	t.Parallel()
	if got := conservativeExpression([]Evidence{{
		File:            "LICENSE",
		LicenseIDs:      []string{"MIT"},
		CoveragePercent: DefaultMinimumCoveragePercent - 1,
	}}, DefaultMinimumCoveragePercent); got != "" {
		t.Fatalf("got expression %q, want unresolved", got)
	}
}

func TestRejectsInvalidOptions(t *testing.T) {
	t.Parallel()
	if _, err := NewClassifier(Options{MinimumConfidence: 2}); err == nil {
		t.Fatal("expected invalid confidence error")
	}
}
