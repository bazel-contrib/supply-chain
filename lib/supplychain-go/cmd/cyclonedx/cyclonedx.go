package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func main() {
	var outPath, graphPath, classificationsPath, format, validatorPath string
	flag.StringVar(&outPath, "out", "", "The path to write the generated CycloneDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the CycloneDX SBOM (json or xml).")
	flag.StringVar(&validatorPath, "validator", "", "Optional runfiles path or executable path to the CycloneDX CLI validator.")
	flag.Parse()

	if outPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --out flag is required")
		os.Exit(1)
	}

	if graphPath == "" || classificationsPath == "" {
		fmt.Fprintln(os.Stderr, "Error: both --graph and --classifications flags are required")
		os.Exit(1)
	}

	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading graph: %v\n", err)
		os.Exit(1)
	}

	var graph sbom.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing graph: %v\n", err)
		os.Exit(1)
	}

	classBytes, err := os.ReadFile(classificationsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading classifications: %v\n", err)
		os.Exit(1)
	}

	var classifications sbom.Classifications
	if err := json.Unmarshal(classBytes, &classifications); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing classifications: %v\n", err)
		os.Exit(1)
	}

	bom, err := GenerateBOM(graph, classifications)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating BOM: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	switch format {
	case "json":
		encoder = cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	case "xml":
		encoder = cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatXML)
	default:
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a supported format. Use 'json' or 'xml'\n", format)
		os.Exit(1)
	}

	if err := encoder.EncodeVersion(bom, cdx.SpecVersion1_6); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding BOM: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if validatorPath != "" {
		if err := sbom.RunValidator(validatorPath, "validate", "--input-file", outPath, "--input-format", format, "--input-version", "v1_6", "--fail-on-errors"); err != nil {
			fmt.Fprintf(os.Stderr, "Error validating BOM: %v\n", err)
			os.Exit(1)
		}
	}
}

func GenerateBOM(graph sbom.GraphConfig, classifications sbom.Classifications) (*cdx.BOM, error) {
	graph = sbom.NormalizeGraph(graph)
	components := make([]cdx.Component, 0)
	labelToBOMRef := make(map[string]string)
	var rootComponent *cdx.Component

	// Helper function to create component from node
	createComponent := func(node *sbom.NodeConfig) (cdx.Component, error) {
		if node.MetadataFile == "" {
			return cdx.Component{}, fmt.Errorf("node %s has no metadata file", node.Label)
		}

		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading metadata file %s: %w", node.MetadataFile, err)
		}

		purl := pkgMetadata.GetPURL()

		fullName := purl.Name
		if purl.Namespace != "" {
			fullName = purl.Namespace + "/" + fullName
		}

		bomRef := purl.String()
		component := cdx.Component{
			BOMRef:     bomRef,
			Type:       cdx.ComponentTypeLibrary,
			Name:       fullName,
			PackageURL: bomRef,
		}
		if node.Involvement != "" {
			setInvolvementProperty(&component, node.Involvement)
		} else if node.Scope == "build" {
			setInvolvementProperty(&component, "build")
		}
		if node.Scope == "build" {
			component.Scope = cdx.ScopeExcluded
		}

		downloadLocation := ""
		qualifiers := purl.Qualifiers.Map()
		if qualifier := strings.TrimSpace(qualifiers["download_url"]); qualifier != "" {
			downloadLocation = qualifier
		} else if qualifier := strings.TrimSpace(qualifiers["url_download"]); qualifier != "" {
			downloadLocation = qualifier
		}
		if downloadLocation != "" {
			component.ExternalReferences = &[]cdx.ExternalReference{
				{
					Type: cdx.ERTypeDistribution,
					URL:  downloadLocation,
				},
			}
		}

		// Add version if available
		if purl.Version != "" {
			component.Version = purl.Version
		}

		licenseAttribute, err := supplychain.GetLicenseAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading license of %s: %w", node.Label, err)
		}
		component.Licenses = licensesOf(licenseAttribute)

		labelToBOMRef[node.Label] = bomRef
		return component, nil
	}

	// Handle root component
	if classifications.RootComponent != nil {
		comp, err := createComponent(classifications.RootComponent)
		if err != nil {
			return nil, err
		}
		rootComponent = &comp
	}

	// Add direct and transitive dependencies, skipping any component whose
	// BOMRef was already added (classification should already be unique per
	// component, but a duplicate BOMRef is invalid CycloneDX, so guard here
	// too).
	componentIndexes := make(map[string]int)
	addComponent := func(node *sbom.NodeConfig) error {
		comp, err := createComponent(node)
		if err != nil {
			return err
		}
		if index, exists := componentIndexes[comp.BOMRef]; exists {
			mergeComponentClassifications(&components[index], comp)
			return nil
		}
		componentIndexes[comp.BOMRef] = len(components)
		components = append(components, comp)
		return nil
	}

	for i := range classifications.Dependencies.Direct {
		if err := addComponent(&classifications.Dependencies.Direct[i]); err != nil {
			return nil, err
		}
	}

	for i := range classifications.Dependencies.Transitive {
		if err := addComponent(&classifications.Dependencies.Transitive[i]); err != nil {
			return nil, err
		}
	}

	// Collapse target edges to component edges. Containment is represented by
	// compositions; only dependency edges participate in cycle removal.
	componentEdges := componentDependencyEdges(graph.Edges, labelToBOMRef)
	assembliesByParent := make(map[string]map[string]struct{})
	for _, edge := range componentEdges {
		relationship := effectiveRelationship(edge)
		if relationship == "contains" {
			assemblies, ok := assembliesByParent[edge.from]
			if !ok {
				assemblies = make(map[string]struct{})
				assembliesByParent[edge.from] = assemblies
			}
			assemblies[edge.to] = struct{}{}
		} else if relationship != "" && relationship != "depends_on" {
			component := componentByRef(edge.from, rootComponent, components, componentIndexes)
			if component != nil {
				addRelationshipProperty(component, relationship, edge.to)
			}
		}
	}
	compositions := make([]cdx.Composition, 0, len(assembliesByParent))
	for parentRef, assemblyRefs := range assembliesByParent {
		assemblies := make([]cdx.BOMReference, 0, len(assemblyRefs))
		for assemblyRef := range assemblyRefs {
			assemblies = append(assemblies, cdx.BOMReference(assemblyRef))
		}
		sort.Slice(assemblies, func(i, j int) bool {
			return assemblies[i] < assemblies[j]
		})
		compositions = append(compositions, cdx.Composition{
			BOMRef:     parentRef,
			Aggregate:  cdx.CompositionAggregateUnknown,
			Assemblies: &assemblies,
		})
	}
	sort.Slice(compositions, func(i, j int) bool {
		return compositions[i].BOMRef < compositions[j].BOMRef
	})
	rootRef := ""
	if rootComponent != nil {
		rootRef = rootComponent.BOMRef
	}
	componentEdges = makeAcyclicComponentDependencies(componentEdges, rootRef)
	depMap := make(map[string][]string) // parent BOMRef -> []child BOMRefs
	for _, edge := range componentEdges {
		depMap[edge.from] = append(depMap[edge.from], edge.to)
	}

	// Convert to CycloneDX Dependencies format
	var dependencies *[]cdx.Dependency
	if len(depMap) > 0 {
		deps := make([]cdx.Dependency, 0, len(depMap))
		for parentRef, childRefs := range depMap {
			sort.Strings(childRefs)
			deps = append(deps, cdx.Dependency{
				Ref:          parentRef,
				Dependencies: &childRefs,
			})
		}
		// Map iteration order is random; keep the output reproducible.
		sort.Slice(deps, func(i, j int) bool { return deps[i].Ref < deps[j].Ref })
		dependencies = &deps
	}

	bom := cdx.NewBOM()
	bom.Version = 1

	if len(components) > 0 {
		bom.Components = &components
	}

	bom.Dependencies = dependencies
	if len(compositions) > 0 {
		bom.Compositions = &compositions
	}

	// Add metadata with tool information and root component
	metadata := &cdx.Metadata{
		Tools: &cdx.ToolsChoice{
			Components: &[]cdx.Component{
				{
					Type: cdx.ComponentTypeApplication,
					Name: "Bazel Supply Chain Tools CycloneDX generator",
				},
			},
		},
	}

	// Set the root component in metadata if we have one
	if rootComponent != nil {
		metadata.Component = rootComponent
	}

	bom.Metadata = metadata

	return bom, nil
}

type componentDependencyEdge struct {
	from         string
	to           string
	involvement  string
	relationship string
}

func effectiveRelationship(edge componentDependencyEdge) string {
	switch edge.involvement {
	case "build":
		return ""
	case "embedded":
		return "contains"
	default:
		return edge.relationship
	}
}

func componentDependencyEdges(edges []sbom.EdgeConfig, labelToBOMRef map[string]string) []componentDependencyEdge {
	type edgeKey struct {
		from string
		to   string
	}
	bestByPair := make(map[edgeKey]componentDependencyEdge)
	for _, edge := range edges {
		fromRef, fromOk := labelToBOMRef[edge.From]
		toRef, toOk := labelToBOMRef[edge.To]
		if !fromOk || !toOk || fromRef == toRef {
			// A self-reference can appear if distinct graph nodes resolved to
			// the same component (e.g. duplicate metadata).
			continue
		}

		key := edgeKey{from: fromRef, to: toRef}
		candidate := componentDependencyEdge{
			from:         fromRef,
			to:           toRef,
			involvement:  edge.Involvement,
			relationship: edge.Relationship,
		}
		current, found := bestByPair[key]
		if !found {
			bestByPair[key] = candidate
		} else {
			if involvementPriority(candidate.involvement) < involvementPriority(current.involvement) {
				current.involvement = candidate.involvement
			}
			if relationshipPriority(candidate.relationship) < relationshipPriority(current.relationship) {
				current.relationship = candidate.relationship
			}
			bestByPair[key] = current
		}
	}

	result := make([]componentDependencyEdge, 0, len(bestByPair))
	for _, edge := range bestByPair {
		result = append(result, edge)
	}
	sortComponentDependencyEdges(result)
	return result
}

func relationshipPriority(relationship string) int {
	switch relationship {
	case "contains":
		return 0
	case "static_link":
		return 1
	case "dynamic_link":
		return 2
	case "test_dependency":
		return 3
	case "dev_dependency":
		return 4
	default:
		return 5
	}
}

func componentByRef(ref string, root *cdx.Component, components []cdx.Component, indexes map[string]int) *cdx.Component {
	if root != nil && root.BOMRef == ref {
		return root
	}
	if index, ok := indexes[ref]; ok {
		return &components[index]
	}
	return nil
}

func addRelationshipProperty(component *cdx.Component, relationship, childRef string) {
	property := cdx.Property{Name: "bazel:relationship:" + relationship, Value: childRef}
	if component.Properties == nil {
		component.Properties = &[]cdx.Property{property}
	} else {
		for _, existing := range *component.Properties {
			if existing == property {
				return
			}
		}
		*component.Properties = append(*component.Properties, property)
	}
	sort.Slice(*component.Properties, func(i, j int) bool {
		if (*component.Properties)[i].Name != (*component.Properties)[j].Name {
			return (*component.Properties)[i].Name < (*component.Properties)[j].Name
		}
		return (*component.Properties)[i].Value < (*component.Properties)[j].Value
	})
}

func makeAcyclicComponentDependencies(edges []componentDependencyEdge, root string) []componentDependencyEdge {
	dependencyEdges := make([]componentDependencyEdge, 0, len(edges))
	for _, edge := range edges {
		if effectiveRelationship(edge) != "contains" {
			dependencyEdges = append(dependencyEdges, edge)
		}
	}
	edges = dependencyEdges
	componentByNode, cyclicComponents := dependencyStronglyConnectedComponents(edges)
	filtered := make([]componentDependencyEdge, 0, len(edges))
	for _, edge := range edges {
		fromComponent, fromFound := componentByNode[edge.from]
		toComponent, toFound := componentByNode[edge.to]
		if edge.involvement == "build" &&
			fromFound && toFound &&
			fromComponent == toComponent &&
			cyclicComponents[fromComponent] {
			continue
		}
		filtered = append(filtered, edge)
	}

	// SCC filtering handles build-only back edges first. A sorted DFS then
	// removes any remaining back edges in the dependency graph.
	adjacency := make(map[string][]int)
	nodes := make(map[string]bool)
	for index, edge := range filtered {
		adjacency[edge.from] = append(adjacency[edge.from], index)
		nodes[edge.from] = true
		nodes[edge.to] = true
	}
	for from := range adjacency {
		sort.Slice(adjacency[from], func(i, j int) bool {
			return filtered[adjacency[from][i]].to < filtered[adjacency[from][j]].to
		})
	}

	removed := make([]bool, len(filtered))
	state := make(map[string]uint8)
	var visit func(string)
	visit = func(from string) {
		state[from] = 1
		for _, edgeIndex := range adjacency[from] {
			if removed[edgeIndex] {
				continue
			}
			to := filtered[edgeIndex].to
			switch state[to] {
			case 0:
				visit(to)
			case 1:
				removed[edgeIndex] = true
			}
		}
		state[from] = 2
	}

	if root != "" && nodes[root] {
		visit(root)
	}
	orderedNodes := make([]string, 0, len(nodes))
	for node := range nodes {
		orderedNodes = append(orderedNodes, node)
	}
	sort.Strings(orderedNodes)
	for _, node := range orderedNodes {
		if state[node] == 0 {
			visit(node)
		}
	}

	result := make([]componentDependencyEdge, 0, len(filtered))
	for index, edge := range filtered {
		if !removed[index] {
			result = append(result, edge)
		}
	}
	return result
}

func dependencyStronglyConnectedComponents(edges []componentDependencyEdge) (map[string]int, map[int]bool) {
	adjacency := make(map[string][]string)
	nodes := make(map[string]bool)
	for _, edge := range edges {
		adjacency[edge.from] = append(adjacency[edge.from], edge.to)
		nodes[edge.from] = true
		nodes[edge.to] = true
	}
	orderedNodes := make([]string, 0, len(nodes))
	for node := range nodes {
		orderedNodes = append(orderedNodes, node)
	}
	sort.Strings(orderedNodes)
	for from := range adjacency {
		sort.Strings(adjacency[from])
	}

	index := 0
	componentID := 0
	indices := make(map[string]int)
	lowLinks := make(map[string]int)
	onStack := make(map[string]bool)
	stack := make([]string, 0, len(nodes))
	componentByNode := make(map[string]int, len(nodes))
	componentSizes := make(map[int]int)
	selfLoops := make(map[int]bool)
	var connect func(string)
	connect = func(node string) {
		index++
		indices[node] = index
		lowLinks[node] = index
		stack = append(stack, node)
		onStack[node] = true

		for _, child := range adjacency[node] {
			if indices[child] == 0 {
				connect(child)
				if lowLinks[child] < lowLinks[node] {
					lowLinks[node] = lowLinks[child]
				}
			} else if onStack[child] && indices[child] < lowLinks[node] {
				lowLinks[node] = indices[child]
			}
		}

		if lowLinks[node] != indices[node] {
			return
		}
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			componentByNode[last] = componentID
			componentSizes[componentID]++
			if last == node {
				break
			}
		}
		componentID++
	}
	for _, node := range orderedNodes {
		if indices[node] == 0 {
			connect(node)
		}
	}

	for _, edge := range edges {
		if edge.from == edge.to {
			selfLoops[componentByNode[edge.from]] = true
		}
	}
	cyclicComponents := make(map[int]bool)
	for id, size := range componentSizes {
		if size > 1 || selfLoops[id] {
			cyclicComponents[id] = true
		}
	}
	return componentByNode, cyclicComponents
}

func sortComponentDependencyEdges(edges []componentDependencyEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		if edges[i].to != edges[j].to {
			return edges[i].to < edges[j].to
		}
		return relationshipPriority(edges[i].relationship) < relationshipPriority(edges[j].relationship)
	})
}

// mergeComponentClassifications combines target occurrences of one PURL using
// the best root-path involvement. Runtime is represented by no property.
func mergeComponentClassifications(existing *cdx.Component, incoming cdx.Component) {
	best := bestComponentInvolvement(*existing, incoming)
	excluded := existing.Scope == cdx.ScopeExcluded && incoming.Scope == cdx.ScopeExcluded
	properties := mergedProperties(*existing, incoming)
	existing.Properties = properties
	if excluded {
		existing.Scope = cdx.ScopeExcluded
	} else {
		existing.Scope = ""
	}
	setInvolvementProperty(existing, best)
}

func bestComponentInvolvement(components ...cdx.Component) string {
	best := "build"
	bestRank := 2
	for _, component := range components {
		involvement := "runtime"
		if component.Scope == cdx.ScopeExcluded {
			involvement = "build"
		} else if component.Properties != nil {
			for _, property := range *component.Properties {
				if property.Name == "bazel:involvement" {
					involvement = property.Value
					break
				}
			}
		}
		rank := involvementPriority(involvement)
		if rank < bestRank {
			best = involvement
			bestRank = rank
		}
	}
	return best
}

func mergedProperties(components ...cdx.Component) *[]cdx.Property {
	properties := make([]cdx.Property, 0)
	seen := make(map[string]bool)
	for _, component := range components {
		if component.Properties == nil {
			continue
		}
		for _, property := range *component.Properties {
			if property.Name == "bazel:involvement" {
				continue
			}
			key := property.Name + "\x00" + property.Value
			if seen[key] {
				continue
			}
			seen[key] = true
			properties = append(properties, property)
		}
	}
	sort.Slice(properties, func(i, j int) bool {
		if properties[i].Name != properties[j].Name {
			return properties[i].Name < properties[j].Name
		}
		return properties[i].Value < properties[j].Value
	})
	if len(properties) == 0 {
		return nil
	}
	return &properties
}

func involvementPriority(involvement string) int {
	switch involvement {
	case "embedded":
		return 1
	case "build":
		return 2
	default:
		return 0
	}
}

func setInvolvementProperty(component *cdx.Component, involvement string) {
	if involvement == "" || involvement == "runtime" {
		return
	}
	property := cdx.Property{Name: "bazel:involvement", Value: involvement}
	if component.Properties == nil {
		component.Properties = &[]cdx.Property{property}
		return
	}
	*component.Properties = append(*component.Properties, property)
	sort.Slice(*component.Properties, func(i, j int) bool {
		if (*component.Properties)[i].Name != (*component.Properties)[j].Name {
			return (*component.Properties)[i].Name < (*component.Properties)[j].Name
		}
		return (*component.Properties)[i].Value < (*component.Properties)[j].Value
	})
}

// licensesOf converts the license attribute of a package to CycloneDX: a
// license identified by its SPDX identifier, or by its name if it is custom.
// Anything else is an SPDX license expression.
func licensesOf(attribute *supplychain.LicenseAttribute) *cdx.Licenses {
	if attribute == nil {
		return nil
	}

	root := attribute.Root()
	if !root.IsSingle() || root.OrLater {
		return &cdx.Licenses{{Expression: root.String()}}
	}
	kind := root.Kind
	license := cdx.License{ID: kind.Identifier}
	if kind.IsCustom() {
		license = cdx.License{Name: kind.Name}
		if license.Name == "" {
			license.Name = kind.Identifier
		}
	}
	return &cdx.Licenses{{License: &license}}
}
