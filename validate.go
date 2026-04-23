package fournf

import (
	"fmt"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// ValidateGraph loads the schema graph from schemaDir and checks that no entity
// schema has edges with .Field(). Returns a list of violations (empty = pass).
func ValidateGraph(schemaDir, pkg string) ([]string, error) {
	graph, err := entc.LoadGraph(schemaDir, &gen.Config{
		Package: pkg,
	})
	if err != nil {
		return nil, fmt.Errorf("loading schema graph: %w", err)
	}

	var violations []string
	for _, n := range graph.Nodes {
		if isJoinTable(n) {
			violations = append(violations, validateJoinTableGraph(n)...)
			continue
		}
		for _, e := range n.Edges {
			if e.Field() != nil {
				violations = append(violations, fmt.Sprintf(
					"entity %q has edge %q with .Field(%q)",
					n.Name, e.Name, e.Field().Name,
				))
			}
		}
	}
	return violations, nil
}

func validateJoinTableGraph(n *gen.Type) []string {
	var violations []string
	var fkEdges int
	for _, e := range n.Edges {
		if e.Field() != nil {
			fkEdges++
		}
	}
	if fkEdges < 2 {
		violations = append(violations, fmt.Sprintf(
			"join table %q has %d foreign key edge(s), need at least 2",
			n.Name, fkEdges,
		))
	}
	for _, f := range n.Fields {
		if !f.IsEdgeField() {
			violations = append(violations, fmt.Sprintf(
				"join table %q has non-foreign-key field %q",
				n.Name, f.Name,
			))
		}
	}
	return violations
}
