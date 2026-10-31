package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/spdx/tools-golang/spdx"
)

// licenseCollector gathers the custom licenses used by packages, which SPDX
// requires to be defined in the document as `LicenseRef-` licenses.
type licenseCollector struct {
	byID map[string]*spdx.OtherLicense
	// keyByID is the identity of the terms defined under each id, to tell
	// apart different licenses that would get the same reference.
	keyByID map[string]string
	// fallback holds the ids whose text is only the license name, so that a
	// later package providing the real text can replace it.
	fallback map[string]bool
}

func newLicenseCollector() *licenseCollector {
	return &licenseCollector{
		byID:     map[string]*spdx.OtherLicense{},
		keyByID:  map[string]string{},
		fallback: map[string]bool{},
	}
}

// declared returns the SPDX license expression for the package and records
// the custom licenses it references.
//
// SPDX 2.3 only allows exceptions from the SPDX list after `WITH`, so a
// license combined with a custom exception is defined as a `LicenseRef-` for
// the combined terms.
func (c *licenseCollector) declared(attribute *supplychain.LicenseAttribute) string {
	if attribute == nil {
		return "NOASSERTION"
	}
	root := attribute.Root()
	return c.render(attribute, root, root)
}

func (c *licenseCollector) render(attribute *supplychain.LicenseAttribute, root, node supplychain.LicenseExpression) string {
	switch {
	case node.Kind != nil:
		return c.leaf(attribute, root, node)
	case node.Operator == supplychain.LicenseWith:
		leaf := node.Operands[0]
		if !node.Exception.IsCustom() {
			return c.leaf(attribute, root, leaf) + " WITH " + node.Exception.Identifier
		}
		key := "WITH\x00" + leaf.Kind.Identifier + "\x00" + node.Exception.Identifier
		if leaf.OrLater {
			key += "\x00+"
		}
		id := "LicenseRef-" + refSuffix(leaf.LeafID(), "LicenseRef-") + "-WITH-" + refSuffix(node.Exception.RefID(true), "AdditionRef-")
		return c.define(id, key, leaf.Kind.Name+" WITH "+node.Exception.Name, extractedText(attribute, root, node.Exception.Name))
	default:
		parts := make([]string, len(node.Operands))
		for i, o := range node.Operands {
			parts[i] = c.render(attribute, root, o)
			if (o.Operator == supplychain.LicenseAnd || o.Operator == supplychain.LicenseOr) && o.Operator != node.Operator {
				parts[i] = "(" + parts[i] + ")"
			}
		}
		return strings.Join(parts, " "+string(node.Operator)+" ")
	}
}

// leaf renders a license. A custom license cannot be followed by `+` in SPDX,
// so "or later" is part of the defined reference.
func (c *licenseCollector) leaf(attribute *supplychain.LicenseAttribute, root, node supplychain.LicenseExpression) string {
	kind := *node.Kind
	if !kind.IsCustom() {
		return node.String()
	}
	key := "LICENSE\x00" + kind.Identifier
	name := kind.Name
	if node.OrLater {
		key += "\x00+"
		name += " or later"
	}
	return c.define(node.LeafID(), key, name, extractedText(attribute, root, name))
}

// refSuffix is the part of a custom reference after its prefix.
func refSuffix(ref, prefix string) string {
	return strings.TrimPrefix(ref, prefix)
}

// define records a custom license under id, replacing a name-only text by a
// real one, and returns the id it was recorded under. If another license
// already uses id (ignoring case, as SPDX does), the id is disambiguated.
func (c *licenseCollector) define(id, key, name string, text licenseText) string {
	base := id
	for i := 0; ; i++ {
		other, taken := c.keyByID[strings.ToLower(id)]
		if !taken || other == key {
			break
		}
		id = base + "-" + supplychain.ShortHash(key, fmt.Sprint(i))
	}
	if existing, ok := c.byID[strings.ToLower(id)]; ok {
		if c.fallback[strings.ToLower(id)] && text.real {
			existing.ExtractedText = text.text
			delete(c.fallback, strings.ToLower(id))
		}
		return id
	}
	c.keyByID[strings.ToLower(id)] = key
	c.byID[strings.ToLower(id)] = &spdx.OtherLicense{
		LicenseIdentifier: id,
		LicenseName:       name,
		ExtractedText:     text.text,
	}
	if !text.real {
		c.fallback[strings.ToLower(id)] = true
	}
	return id
}

func (c *licenseCollector) otherLicenses() []*spdx.OtherLicense {
	licenses := make([]*spdx.OtherLicense, 0, len(c.byID))
	for _, l := range c.byID {
		licenses = append(licenses, l)
	}
	sort.Slice(licenses, func(i, j int) bool {
		return licenses[i].LicenseIdentifier < licenses[j].LicenseIdentifier
	})
	return licenses
}

type licenseText struct {
	text string
	// real is whether text was read from a license file.
	real bool
}

// extractedText is the license text file of the package when the license
// consists of a single set of terms (a license, optionally with an
// exception), otherwise the fallback name.
func extractedText(attribute *supplychain.LicenseAttribute, root supplychain.LicenseExpression, fallback string) licenseText {
	if attribute.Text != "" && (root.IsSingle() || root.Operator == supplychain.LicenseWith) {
		if text, err := os.ReadFile(attribute.Text); err == nil {
			return licenseText{text: string(text), real: true}
		}
	}
	return licenseText{text: fallback}
}
