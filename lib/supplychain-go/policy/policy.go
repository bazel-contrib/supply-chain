// Package policy evaluates a license coverage report against an organization
// supplied policy. It contains no organization-specific license decisions.
package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	AllowlistSchemaVersion = 1
	ReportSchemaVersion    = 1
)

// Allowlist is intentionally data-only. Organizations provide their own
// identifiers, exceptions, and reviewed target-scoped decisions.
type Allowlist struct {
	SchemaVersion            int               `json:"schema_version"`
	AllowedLicenseIDs        []string          `json:"allowed_license_identifiers"`
	AllowedLicenseExceptions []string          `json:"allowed_license_exceptions,omitempty"`
	AllowedMissingPURLs      []string          `json:"allowed_missing_license_purls,omitempty"`
	TargetExceptions         []TargetException `json:"target_exceptions"`
}

// TargetException allows one exact package/license expression for one exact
// artifact root. The scope is intentionally explicit to make exceptions
// reviewable and unsuitable for broad target patterns.
type TargetException struct {
	RootTarget        string `json:"root_target"`
	PURL              string `json:"purl"`
	LicenseExpression string `json:"license_expression"`
	Scope             string `json:"scope"`
	Rationale         string `json:"rationale"`
	ExpiresOn         string `json:"expires_on"`
}

// Report is the input contract produced by a coverage implementation.
type Report struct {
	SchemaVersion       int                 `json:"schema_version"`
	RootTarget          string              `json:"root_target,omitempty"`
	MissingLicensePURLs []string            `json:"missing_license_purls"`
	LicensedPackages    map[string][]string `json:"licensed_packages"`
}

// Evaluation is deterministic and suitable for both CLI output and CI.
type Evaluation struct {
	Acceptable           bool                `json:"acceptable"`
	MissingLicensePURLs  []string            `json:"missing_license_purls,omitempty"`
	UnacceptablePackages map[string][]string `json:"unacceptable_packages,omitempty"`
}

// LoadAllowlist reads and validates a JSON policy file.
func LoadAllowlist(path string) (Allowlist, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Allowlist{}, err
	}
	var allowlist Allowlist
	if err := json.Unmarshal(contents, &allowlist); err != nil {
		return Allowlist{}, err
	}
	if err := allowlist.Validate(); err != nil {
		return Allowlist{}, err
	}
	return allowlist, nil
}

// Validate checks schema shape and rejects ambiguous target exceptions.
func (a Allowlist) Validate() error {
	if a.SchemaVersion != AllowlistSchemaVersion {
		return fmt.Errorf("unsupported license allowlist schema version: %d", a.SchemaVersion)
	}
	if a.AllowedLicenseIDs == nil {
		return fmt.Errorf("allowed_license_identifiers must be present")
	}
	if a.TargetExceptions == nil {
		return fmt.Errorf("target_exceptions must be present")
	}
	if hasDuplicate(a.AllowedLicenseIDs) || hasDuplicate(a.AllowedLicenseExceptions) || hasDuplicate(a.AllowedMissingPURLs) {
		return fmt.Errorf("allowlist entries must not contain duplicates")
	}
	for _, identifier := range append(append([]string{}, a.AllowedLicenseIDs...), a.AllowedLicenseExceptions...) {
		if strings.TrimSpace(identifier) == "" {
			return fmt.Errorf("license identifiers must be non-empty")
		}
	}
	for _, purl := range a.AllowedMissingPURLs {
		if strings.TrimSpace(purl) == "" {
			return fmt.Errorf("allowed_missing_license_purls entries must be non-empty")
		}
	}
	seenExceptions := make(map[string]bool, len(a.TargetExceptions))
	for _, exception := range a.TargetExceptions {
		if err := exception.validate(); err != nil {
			return err
		}
		key := strings.Join([]string{exception.RootTarget, exception.PURL, exception.LicenseExpression}, "\x00")
		if seenExceptions[key] {
			return fmt.Errorf("target exceptions must not duplicate root, PURL, and expression")
		}
		seenExceptions[key] = true
	}
	return nil
}

func (e TargetException) validate() error {
	if !validArtifactRoot(e.RootTarget) {
		return fmt.Errorf("invalid target exception root_target %q", e.RootTarget)
	}
	if e.PURL == "" || e.LicenseExpression == "" || e.Rationale == "" {
		return fmt.Errorf("target exception purl, license_expression, and rationale must be non-empty")
	}
	if strings.TrimSpace(e.Scope) == "" {
		return fmt.Errorf("target exception scope must be non-empty")
	}
	if _, err := time.Parse("2006-01-02", e.ExpiresOn); err != nil {
		return fmt.Errorf("target exception expires_on must be an ISO-8601 date: %w", err)
	}
	return nil
}

// Evaluate checks missing metadata and every reported SPDX expression. An OR
// expression is acceptable when one branch is acceptable; every AND term must
// be acceptable. Exceptions are valid only as the right operand of WITH.
func Evaluate(report Report, allowlist Allowlist, today time.Time) (Evaluation, error) {
	if report.SchemaVersion != ReportSchemaVersion {
		return Evaluation{}, fmt.Errorf("unsupported coverage report schema version: %d", report.SchemaVersion)
	}
	if report.MissingLicensePURLs == nil {
		return Evaluation{}, fmt.Errorf("missing_license_purls must be present")
	}
	if report.LicensedPackages == nil {
		return Evaluation{}, fmt.Errorf("licensed_packages must be present")
	}
	if err := allowlist.Validate(); err != nil {
		return Evaluation{}, err
	}
	allowedIDs := toSet(allowlist.AllowedLicenseIDs)
	allowedExceptions := toSet(allowlist.AllowedLicenseExceptions)
	allowedMissing := toSet(allowlist.AllowedMissingPURLs)

	missing := make([]string, 0)
	for _, purl := range report.MissingLicensePURLs {
		if purl == "" {
			return Evaluation{}, fmt.Errorf("missing_license_purls entries must be non-empty")
		}
		if !allowedMissing[purl] {
			missing = append(missing, purl)
		}
	}
	sort.Strings(missing)

	unacceptable := make(map[string][]string)
	for purl, expressions := range report.LicensedPackages {
		if purl == "" || len(expressions) == 0 {
			return Evaluation{}, fmt.Errorf("licensed_packages must map non-empty PURLs to non-empty expression lists")
		}
		for _, expression := range expressions {
			if strings.TrimSpace(expression) == "" {
				return Evaluation{}, fmt.Errorf("licensed_packages expressions must be non-empty")
			}
			acceptable, err := acceptableExpression(expression, allowedIDs, allowedExceptions)
			if err != nil {
				return Evaluation{}, err
			}
			if acceptable || matchingException(report.RootTarget, purl, expression, allowlist.TargetExceptions, today) {
				continue
			}
			unacceptable[purl] = append(unacceptable[purl], expression)
		}
		sort.Strings(unacceptable[purl])
	}
	return Evaluation{
		Acceptable:           len(missing) == 0 && len(unacceptable) == 0,
		MissingLicensePURLs:  missing,
		UnacceptablePackages: unacceptable,
	}, nil
}

func matchingException(rootTarget, purl, expression string, exceptions []TargetException, today time.Time) bool {
	if rootTarget == "" {
		return false
	}
	for _, exception := range exceptions {
		if exception.RootTarget != rootTarget || exception.PURL != purl || exception.LicenseExpression != expression {
			continue
		}
		expiresOn, _ := time.Parse("2006-01-02", exception.ExpiresOn)
		if !today.After(expiresOn) {
			return true
		}
	}
	return false
}

var tokenPattern = regexp.MustCompile(`[A-Za-z0-9.+-]+|[()]`)

func acceptableExpression(expression string, allowedIDs, allowedExceptions map[string]bool) (bool, error) {
	normalized := strings.ReplaceAll(expression, "/", " OR ")
	rawTokens := tokenPattern.FindAllString(normalized, -1)
	if len(rawTokens) == 0 || strings.Join(rawTokens, "") != strings.Join(strings.Fields(normalized), "") {
		return false, fmt.Errorf("invalid SPDX license expression %q", expression)
	}
	position := 0
	parseExpression := func() (bool, error) { return false, nil }
	var parsePrimary func() (bool, error)
	var parseConjunction func() (bool, error)
	parsePrimary = func() (bool, error) {
		if position >= len(rawTokens) {
			return false, fmt.Errorf("invalid SPDX license expression %q", expression)
		}
		token := rawTokens[position]
		position++
		if token == "(" {
			result, err := parseExpression()
			if err != nil || position >= len(rawTokens) || rawTokens[position] != ")" {
				return false, fmt.Errorf("invalid SPDX license expression %q", expression)
			}
			position++
			return result, nil
		}
		if token == "AND" || token == "OR" || token == "WITH" || token == ")" {
			return false, fmt.Errorf("invalid SPDX license expression %q", expression)
		}
		return allowedIDs[token], nil
	}
	parseConjunction = func() (bool, error) {
		result, err := parsePrimary()
		if err != nil {
			return false, err
		}
		for position < len(rawTokens) && (rawTokens[position] == "AND" || rawTokens[position] == "WITH") {
			operator := rawTokens[position]
			position++
			if operator == "WITH" {
				if position >= len(rawTokens) || rawTokens[position] == "AND" || rawTokens[position] == "OR" || rawTokens[position] == "WITH" || rawTokens[position] == ")" {
					return false, fmt.Errorf("invalid SPDX license expression %q", expression)
				}
				result = result && allowedExceptions[rawTokens[position]]
				position++
				continue
			}
			term, err := parsePrimary()
			if err != nil {
				return false, err
			}
			result = result && term
		}
		return result, nil
	}
	parseExpression = func() (bool, error) {
		result, err := parseConjunction()
		if err != nil {
			return false, err
		}
		for position < len(rawTokens) && rawTokens[position] == "OR" {
			position++
			alternative, err := parseConjunction()
			if err != nil {
				return false, err
			}
			result = result || alternative
		}
		return result, nil
	}

	result, err := parseExpression()
	if err != nil {
		return false, err
	}
	if position != len(rawTokens) {
		return false, fmt.Errorf("invalid SPDX license expression %q", expression)
	}
	return result, nil
}

func validArtifactRoot(value string) bool {
	if !strings.HasPrefix(value, "//") || strings.ContainsAny(value, " \t\n") || strings.Contains(value, "...") || strings.Contains(value, "*") {
		return false
	}
	withoutPrefix := strings.TrimPrefix(value, "//")
	packageName, target, found := strings.Cut(withoutPrefix, ":")
	return withoutPrefix != "" && found && target != "" && target != "all" && !strings.Contains(target, ":") && !strings.HasSuffix(packageName, "/")
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func toSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
