package fournf

import (
	"fmt"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// ValidateOption configures [ValidateGraph].
type ValidateOption func(*validateConfig)

type validateConfig struct {
	requireTimestamps bool
}

// WithTimestampValidation returns a [ValidateOption] that checks every entity
// schema for created_at and updated_at fields. Join tables and schemas
// annotated with [SkipTimestamps] are exempt.
func WithTimestampValidation() ValidateOption {
	return func(c *validateConfig) {
		c.requireTimestamps = true
	}
}

// ValidateGraph loads the schema graph from schemaDir and checks that no entity
// schema has edges with .Field(). Returns a list of violations (empty = pass).
func ValidateGraph(schemaDir, pkg string, opts ...ValidateOption) ([]string, error) {
	var cfg validateConfig
	for _, o := range opts {
		o(&cfg)
	}

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

	if cfg.requireTimestamps {
		for _, n := range graph.Nodes {
			if isJoinTable(n) || skipTimestamps(n) {
				continue
			}
			violations = append(violations, checkTimestampFieldGraph(n, "created_at", false)...)
			violations = append(violations, checkTimestampFieldGraph(n, "updated_at", true)...)
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

func checkTimestampFieldGraph(n *gen.Type, name string, needUpdateDefault bool) []string {
	var violations []string
	f := findField(n, name)
	if f == nil {
		violations = append(violations, fmt.Sprintf(
			"entity %q is missing required field %q",
			n.Name, name,
		))
		return violations
	}
	if !f.Default {
		violations = append(violations, fmt.Sprintf(
			"entity %q field %q must have a Default value",
			n.Name, name,
		))
	}
	if needUpdateDefault && !f.UpdateDefault {
		violations = append(violations, fmt.Sprintf(
			"entity %q field %q must have an UpdateDefault value",
			n.Name, name,
		))
	}
	return violations
}
