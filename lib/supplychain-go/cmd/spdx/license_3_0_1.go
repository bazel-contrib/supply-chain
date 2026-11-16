package main

import (
	"fmt"
	"os"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

// licenseCollector creates the license elements of the packages.
//
// A license on the SPDX license list is a listed license. Any other license is
// a custom license that carries the text of its license file, if it has one.
// A custom id can be declared with different texts (e.g. different copyright
// lines); each text gets its own element, and the ids after the first are
// suffixed to keep them unique. Exceptions work alike.
type licenseCollector struct {
	elements []spdx.AnyElement
	byKey    map[licenseKey]spdx.AnyElement
	variants map[string]int
	// Counters of the nodes without an identifier of their own. The ids are
	// assigned here, in order of appearance, because the library would number
	// them in an order that changes between runs.
	nodes map[string]int
}

type licenseKey struct {
	id, text  string
	exception bool
}

// declaredLicenses creates the license elements of the packages and the
// hasDeclaredLicense relationships to them. Relationship ids start at
// firstRelationship.
func declaredLicenses(packages []*spdx.Package, licenseOf map[*spdx.Package]*supplychain.LicenseAttribute, firstRelationship int) ([]spdx.AnyElement, []*spdx.Relationship, error) {
	c := &licenseCollector{
		byKey:    make(map[licenseKey]spdx.AnyElement),
		variants: make(map[string]int),
		nodes:    make(map[string]int),
	}
	var relationships []*spdx.Relationship
	for _, pkg := range packages {
		attribute, ok := licenseOf[pkg]
		if !ok {
			continue
		}
		root := attribute.Root()
		text, err := licenseText(attribute)
		if err != nil {
			return nil, nil, fmt.Errorf("reading the license text of %s: %w", pkg.Name, err)
		}
		node, err := c.node(root, text)
		if err != nil {
			return nil, nil, err
		}
		relationships = append(relationships, &spdx.Relationship{
			ID:   fmt.Sprintf("rel-%d", firstRelationship+len(relationships)),
			From: pkg,
			To:   spdx.ElementList{node},
			Type: spdx.RelationshipType_HasDeclaredLicense,
		})
	}
	return c.elements, relationships, nil
}

// licenseText returns the content of the license file of the package, which
// is the text of every custom license and exception of its expression.
func licenseText(attribute *supplychain.LicenseAttribute) (string, error) {
	if attribute.Text == "" {
		return "", nil
	}
	content, err := os.ReadFile(attribute.Text)
	return string(content), err
}

func (c *licenseCollector) node(e supplychain.LicenseExpression, text string) (spdx.AnyLicenseInfo, error) {
	switch {
	case e.IsSingle():
		license := c.license(*e.Kind, text)
		if e.OrLater {
			return &spdx.OrLaterOperator{ID: c.nodeID("OrLaterOperator"), SubjectLicense: license}, nil
		}
		return license, nil
	case e.Operator == supplychain.LicenseWith:
		leaf := e.Operands[0]
		license := c.license(*leaf.Kind, text)
		var subject spdx.AnyExtendableLicense = license
		if leaf.OrLater {
			subject = &spdx.OrLaterOperator{ID: c.nodeID("OrLaterOperator"), SubjectLicense: license}
		}
		return &spdx.WithAdditionOperator{
			ID:                       c.nodeID("WithAdditionOperator"),
			SubjectExtendableLicense: subject,
			SubjectAddition:          c.addition(*e.Exception, text),
		}, nil
	}

	// The id is taken before the operands so that it follows the expression.
	kind := "DisjunctiveLicenseSet"
	if e.Operator == supplychain.LicenseAnd {
		kind = "ConjunctiveLicenseSet"
	}
	id := c.nodeID(kind)
	var members spdx.LicenseInfoList
	for _, operand := range e.Operands {
		member, err := c.node(operand, text)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if e.Operator == supplychain.LicenseAnd {
		return &spdx.ConjunctiveLicenseSet{ID: id, Members: members}, nil
	}
	return &spdx.DisjunctiveLicenseSet{ID: id, Members: members}, nil
}

func (c *licenseCollector) license(kind supplychain.LicenseKind, text string) spdx.AnyLicense {
	if !kind.IsCustom() {
		key := licenseKey{id: kind.Identifier}
		return c.define(key, func(string) spdx.AnyElement {
			return &spdx.ListedLicense{ID: "https://spdx.org/licenses/" + kind.Identifier, Name: kind.Identifier}
		}).(spdx.AnyLicense)
	}
	return c.define(licenseKey{id: kind.RefID(false), text: text}, func(id string) spdx.AnyElement {
		return &spdx.CustomLicense{ID: id, Name: nameOf(kind), Text: text}
	}).(spdx.AnyLicense)
}

func (c *licenseCollector) addition(kind supplychain.LicenseKind, text string) spdx.AnyLicenseAddition {
	if !kind.IsCustom() {
		key := licenseKey{id: kind.Identifier, exception: true}
		return c.define(key, func(string) spdx.AnyElement {
			return &spdx.ListedLicenseException{ID: "https://spdx.org/licenses/" + kind.Identifier, Name: kind.Identifier}
		}).(spdx.AnyLicenseAddition)
	}
	return c.define(licenseKey{id: kind.RefID(true), text: text, exception: true}, func(id string) spdx.AnyElement {
		return &spdx.CustomLicenseAddition{ID: id, Name: nameOf(kind), AdditionText: text}
	}).(spdx.AnyLicenseAddition)
}

// define returns the element of key, creating it with make if it is new. make
// gets the id of the element: the identifier, suffixed if another element
// already uses it with a different text.
func (c *licenseCollector) define(key licenseKey, make func(id string) spdx.AnyElement) spdx.AnyElement {
	if element, ok := c.byKey[key]; ok {
		return element
	}
	id := key.id
	idKey := fmt.Sprintf("%v:%s", key.exception, key.id)
	if n := c.variants[idKey]; n > 0 {
		id = fmt.Sprintf("%s-%d", id, n)
	}
	c.variants[idKey]++
	element := make(id)
	c.byKey[key] = element
	c.elements = append(c.elements, element)
	return element
}

func (c *licenseCollector) nodeID(kind string) string {
	c.nodes[kind]++
	return fmt.Sprintf("%s-%d", kind, c.nodes[kind])
}

func nameOf(kind supplychain.LicenseKind) string {
	if kind.Name == "" {
		return kind.Identifier
	}
	return kind.Name
}
