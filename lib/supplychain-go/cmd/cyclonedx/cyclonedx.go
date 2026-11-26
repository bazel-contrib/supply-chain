package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func main() {
	var outPath, graphPath, classificationsPath, format, buildStatusPath, serialNumberPath, validatorPath string
	var embedPatchFiles bool
	flag.StringVar(&outPath, "out", "", "The path to write the generated CycloneDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the CycloneDX SBOM (json or xml).")
	flag.StringVar(&buildStatusPath, "created_from_status_file", "", "Path to a Bazel volatile-status.txt file to read a BUILD_TIMESTAMP from for metadata.timestamp. If unset, or the file has no BUILD_TIMESTAMP key, a fixed deterministic timestamp (the Unix epoch) is used instead, matching Bazel's own --nostamp default.")
	flag.StringVar(&serialNumberPath, "serial_number_file", "", "Optional path to a file holding the BOM serialNumber, in urn:uuid form. It should be unique to each generated BOM. If unset, the BOM has no serialNumber.")
	flag.StringVar(&validatorPath, "validator", "", "Optional runfiles path or executable path to the CycloneDX CLI validator.")
	flag.BoolVar(&embedPatchFiles, "embed_patch_files", true, "Embed the content of the patch files in the pedigree of the packages they patch. Disable it to keep local patches out of the report.")
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

	var opts []Option
	if !embedPatchFiles {
		opts = append(opts, WithoutPatchFiles())
	}
	bom, err := GenerateBOM(graph, classifications, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating BOM: %v\n", err)
		os.Exit(1)
	}

	created, err := sbom.ReadBuildTimestamp(buildStatusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading build status file: %v\n", err)
		os.Exit(1)
	}
	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}
	bom.Metadata.Timestamp = created.Format(time.RFC3339)

	serialNumber, err := readSerialNumber(serialNumberPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading serial number: %v\n", err)
		os.Exit(1)
	}
	if serialNumber != "" {
		bom.SerialNumber = serialNumber
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

// Option customizes GenerateBOM.
type Option func(*options)

type options struct {
	omitPatchFiles bool
}

// WithoutPatchFiles leaves the content of patch files out of the BOM. The
// patches are still reported, with their type and the issues they resolve, and
// the path of a patch file stays in the notes of the pedigree.
func WithoutPatchFiles() Option {
	return func(o *options) { o.omitPatchFiles = true }
}

func GenerateBOM(graph sbom.GraphConfig, classifications sbom.Classifications, opts ...Option) (*cdx.BOM, error) {
	var config options
	for _, opt := range opts {
		opt(&config)
	}
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

		integrity, err := supplychain.GetIntegrityAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
		}
		if integrity != nil {
			hashes, err := hashesOf(*integrity)
			if err != nil {
				return cdx.Component{}, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
			}
			component.Hashes = hashes
		}

		cpeAttribute, err := supplychain.GetCPEAttribute(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
		}
		if cpeAttribute != nil && len(*cpeAttribute) > 0 {
			// A CycloneDX component has room for a single CPE. It is the first
			// one, and the others are kept as properties.
			component.CPE = (*cpeAttribute)[0]
			for _, cpe := range (*cpeAttribute)[1:] {
				properties := append(valueOrEmpty(component.Properties), cdx.Property{Name: "bazel:cpe", Value: cpe})
				component.Properties = &properties
			}
		}

		patches, err := supplychain.GetPatchAttributes(pkgMetadata)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reading patches of %s: %w", node.Label, err)
		}
		pedigree, err := pedigreeOf(patches, !config.omitPatchFiles)
		if err != nil {
			return cdx.Component{}, fmt.Errorf("error reporting patches of %s: %w", node.Label, err)
		}
		component.Pedigree = pedigree

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

	// Collapse target edges to component edges. Typed relationships other than
	// plain dependencies are recorded as properties of the parent component;
	// containment is not a dependency, so only dependency edges participate in
	// cycle removal.
	componentEdges := componentDependencyEdges(graph.Edges, labelToBOMRef)
	for _, edge := range componentEdges {
		relationship := effectiveRelationship(edge)
		if relationship != "" && relationship != "depends_on" {
			component := componentByRef(edge.from, rootComponent, components, componentIndexes)
			if component != nil {
				addRelationshipProperty(component, relationship, edge.to)
			}
		}
	}
	// Cycles are broken from the entry components: the root, or, when the
	// root has no metadata, its direct dependencies.
	var entryRefs []string
	if rootComponent != nil {
		entryRefs = append(entryRefs, rootComponent.BOMRef)
	} else {
		for i := range classifications.Dependencies.Direct {
			if ref, ok := labelToBOMRef[classifications.Dependencies.Direct[i].Label]; ok {
				entryRefs = append(entryRefs, ref)
			}
		}
	}
	componentEdges = makeAcyclicComponentDependencies(componentEdges, entryRefs)
	depMap := make(map[string][]string) // parent BOMRef -> []child BOMRefs
	for _, edge := range componentEdges {
		depMap[edge.from] = append(depMap[edge.from], edge.to)
	}

	// CycloneDX treats a component that is absent from the dependency graph as
	// having unknown dependencies, and one with an empty dependsOn as having none.
	// Every component of the graph was resolved above, so declare the ones without
	// outgoing edges explicitly instead of leaving them out.
	declareLeaf := func(ref string) {
		if _, ok := depMap[ref]; !ok {
			depMap[ref] = []string{}
		}
	}
	if rootComponent != nil {
		declareLeaf(rootComponent.BOMRef)
	}
	for _, component := range components {
		declareLeaf(component.BOMRef)
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
		// Involvement and relationship must come from the same edge, so that
		// the emitted pair is one that an edge declares.
		current, found := bestByPair[key]
		if !found || edgeIsBetter(candidate, current) {
			bestByPair[key] = candidate
		}
	}

	result := make([]componentDependencyEdge, 0, len(bestByPair))
	for _, edge := range bestByPair {
		result = append(result, edge)
	}
	sortComponentDependencyEdges(result)
	return result
}

// edgeIsBetter ranks the involvement first and, on a tie, the relationship.
func edgeIsBetter(candidate, current componentDependencyEdge) bool {
	if involvementPriority(candidate.involvement) != involvementPriority(current.involvement) {
		return involvementPriority(candidate.involvement) < involvementPriority(current.involvement)
	}
	return relationshipPriority(candidate.relationship) < relationshipPriority(current.relationship)
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

func makeAcyclicComponentDependencies(edges []componentDependencyEdge, roots []string) []componentDependencyEdge {
	// Containment is not a dependency: it is never emitted nor cut. What a
	// component contains is still reached from it, so containment edges are
	// followed when looking for the entry's reachable components.
	dependencyEdges := make([]componentDependencyEdge, 0, len(edges))
	contained := make(map[string][]string)
	for _, edge := range edges {
		if effectiveRelationship(edge) == "contains" {
			contained[edge.from] = append(contained[edge.from], edge.to)
		} else {
			dependencyEdges = append(dependencyEdges, edge)
		}
	}
	for from := range contained {
		sort.Strings(contained[from])
	}
	edges = dependencyEdges
	componentByNode, cyclicComponents := dependencyStronglyConnectedComponents(edges)
	filtered := make([]componentDependencyEdge, 0, len(edges))
	var dropped []componentDependencyEdge
	for _, edge := range edges {
		fromComponent, fromFound := componentByNode[edge.from]
		toComponent, toFound := componentByNode[edge.to]
		if edge.involvement == "build" &&
			fromFound && toFound &&
			fromComponent == toComponent &&
			cyclicComponents[fromComponent] {
			dropped = append(dropped, edge)
			continue
		}
		filtered = append(filtered, edge)
	}

	// Dropping every build edge inside a cycle can disconnect a build tool
	// from the root. Restore a dropped edge when it is the only way the root
	// reaches its target; the DFS below then cuts a different edge of the
	// cycle instead.
	if len(roots) > 0 {
		reachable := reachableFrom(filtered, contained, roots)
		for changed := true; changed; {
			changed = false
			for _, edge := range dropped {
				if reachable[edge.from] && !reachable[edge.to] {
					filtered = append(filtered, edge)
					reachable = reachableFrom(filtered, contained, roots)
					changed = true
				}
			}
		}
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
	for from, children := range contained {
		nodes[from] = true
		for _, child := range children {
			nodes[child] = true
		}
	}

	removed := make([]bool, len(filtered))
	state := make(map[string]uint8)
	var deferred []string
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
		// What a component contains is visited once the current search has
		// returned, as a separate entry: the container, and the nodes above
		// it, are not ancestors of the contained components in the
		// dependency graph, and must not turn their dependencies into back
		// edges.
		deferred = append(deferred, contained[from]...)
		state[from] = 2
	}
	visitDeferred := func() {
		for len(deferred) > 0 {
			next := deferred[0]
			deferred = deferred[1:]
			if state[next] == 0 {
				visit(next)
			}
		}
	}

	for _, root := range roots {
		if nodes[root] && state[root] == 0 {
			visit(root)
			visitDeferred()
		}
	}
	orderedNodes := make([]string, 0, len(nodes))
	for node := range nodes {
		orderedNodes = append(orderedNodes, node)
	}
	sort.Strings(orderedNodes)
	for _, node := range orderedNodes {
		if state[node] == 0 {
			visit(node)
			visitDeferred()
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

func reachableFrom(edges []componentDependencyEdge, contained map[string][]string, roots []string) map[string]bool {
	adjacency := make(map[string][]string)
	for _, edge := range edges {
		adjacency[edge.from] = append(adjacency[edge.from], edge.to)
	}
	reachable := make(map[string]bool, len(roots))
	queue := make([]string, 0, len(roots))
	for _, root := range roots {
		if !reachable[root] {
			reachable[root] = true
			queue = append(queue, root)
		}
	}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for _, next := range append(adjacency[node], contained[node]...) {
			if !reachable[next] {
				reachable[next] = true
				queue = append(queue, next)
			}
		}
	}
	return reachable
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

var hashAlgorithms = map[string]cdx.HashAlgorithm{
	"sha256": cdx.HashAlgoSHA256,
	"sha384": cdx.HashAlgoSHA384,
	"sha512": cdx.HashAlgoSHA512,
}

// hashesOf converts integrity entries to CycloneDX hashes, or nil if none of
// the algorithms is supported.
func hashesOf(integrity supplychain.IntegrityAttribute) (*[]cdx.Hash, error) {
	checksums, err := integrity.Checksums()
	if err != nil {
		return nil, err
	}
	var hashes []cdx.Hash
	for _, checksum := range checksums {
		hashes = append(hashes, cdx.Hash{Algorithm: hashAlgorithms[checksum.Algorithm], Value: checksum.Hex})
	}
	if len(hashes) == 0 {
		return nil, nil
	}
	return &hashes, nil
}

// serialNumberPattern is the one of the CycloneDX schema, which only accepts
// lowercase hexadecimal digits, as RFC 4122 asks for.
var serialNumberPattern = regexp.MustCompile(`^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// readSerialNumber reads a CycloneDX BOM serialNumber from the file at path.
// An empty path returns an empty serial number, meaning "not provided".
func readSerialNumber(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	serialNumber := strings.TrimSpace(string(data))
	if !serialNumberPattern.MatchString(serialNumber) {
		return "", fmt.Errorf("%s must contain a serial number in lowercase urn:uuid form, got %q", path, serialNumber)
	}
	return serialNumber, nil
}

func valueOrEmpty[T any](list *[]T) []T {
	if list == nil {
		return nil
	}
	return *list
}
