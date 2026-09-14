package policy

import (
	"testing"
	"time"
)

func testAllowlist() Allowlist {
	return Allowlist{
		SchemaVersion:            AllowlistSchemaVersion,
		AllowedLicenseIDs:        []string{"Apache-2.0", "MIT"},
		AllowedLicenseExceptions: []string{"LLVM-exception"},
		TargetExceptions:         []TargetException{},
	}
}

func TestRejectsIncompleteInputs(t *testing.T) {
	t.Parallel()
	allowlist := testAllowlist()
	if _, err := Evaluate(Report{SchemaVersion: ReportSchemaVersion}, allowlist, time.Now()); err == nil {
		t.Fatal("expected missing report fields to be rejected")
	}
	if err := (Allowlist{SchemaVersion: AllowlistSchemaVersion}).Validate(); err == nil {
		t.Fatal("expected missing allowlist fields to be rejected")
	}
}

func TestEvaluateSPDXExpressions(t *testing.T) {
	t.Parallel()
	allowlist := testAllowlist()
	report := Report{
		SchemaVersion:       ReportSchemaVersion,
		MissingLicensePURLs: []string{},
		LicensedPackages: map[string][]string{
			"pkg:generic/one@1":   {"MIT"},
			"pkg:generic/two@1":   {"Apache-2.0 AND MIT"},
			"pkg:generic/three@1": {"GPL-3.0-only OR MIT"},
			"pkg:generic/four@1":  {"Apache-2.0 WITH LLVM-exception"},
		},
	}
	evaluation, err := Evaluate(report, allowlist, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Acceptable {
		t.Fatalf("expected acceptable report: %+v", evaluation)
	}
}

func TestEvaluateRejectsMissingAndUnacceptable(t *testing.T) {
	t.Parallel()
	allowlist := testAllowlist()
	report := Report{
		SchemaVersion:       ReportSchemaVersion,
		MissingLicensePURLs: []string{"pkg:generic/missing@1"},
		LicensedPackages: map[string][]string{
			"pkg:generic/bad@1": {"GPL-3.0-only"},
		},
	}
	evaluation, err := Evaluate(report, allowlist, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Acceptable || len(evaluation.MissingLicensePURLs) != 1 || len(evaluation.UnacceptablePackages) != 1 {
		t.Fatalf("unexpected evaluation: %+v", evaluation)
	}
}

func TestTargetExceptionRequiresExactRootAndExpiry(t *testing.T) {
	t.Parallel()
	allowlist := testAllowlist()
	allowlist.TargetExceptions = []TargetException{{
		RootTarget:        "//docs:site",
		PURL:              "pkg:generic/docs@1",
		LicenseExpression: "CC-BY-4.0",
		Scope:             "internal-only",
		Rationale:         "Documentation-only artifact.",
		ExpiresOn:         "2099-01-01",
	}}
	report := Report{
		SchemaVersion:       ReportSchemaVersion,
		MissingLicensePURLs: []string{},
		RootTarget:          "//docs:site",
		LicensedPackages: map[string][]string{
			"pkg:generic/docs@1": {"CC-BY-4.0"},
		},
	}
	evaluation, err := Evaluate(report, allowlist, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.Acceptable {
		t.Fatalf("expected target exception to apply: %+v", evaluation)
	}

	report.RootTarget = "//other:binary"
	evaluation, err = Evaluate(report, allowlist, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Acceptable {
		t.Fatal("target exception applied to the wrong root")
	}
}

func TestRejectsStandaloneSPDXException(t *testing.T) {
	t.Parallel()
	allowlist := testAllowlist()
	report := Report{
		SchemaVersion:       ReportSchemaVersion,
		MissingLicensePURLs: []string{},
		LicensedPackages: map[string][]string{
			"pkg:generic/standalone@1": {"LLVM-exception"},
		},
	}
	evaluation, err := Evaluate(report, allowlist, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Acceptable {
		t.Fatal("standalone SPDX exception was accepted")
	}
}
