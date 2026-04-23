// Package fournf provides an entc extension that enforces 4NF.
//
// Entity schemas must not have edges that use .Field() (which would place a foreign key
// column on the entity table). Only schemas annotated with [JoinTable] may do so.
// Wire it into your entc.go:
//
//	ext, err := fournf.NewExtension()
//	if err != nil {
//	    log.Fatalf("creating fournf extension: %v", err)
//	}
//	entc.Extensions(ext)
package fournf

import (
	"fmt"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// Extension is an entc.Extension that fails code generation when an entity
// schema (one NOT annotated with [JoinTable]) contains an edge with .Field().
type Extension struct {
	entc.DefaultExtension
}

// NewExtension returns a new 4NF extension.
func NewExtension() (*Extension, error) {
	return &Extension{}, nil
}

// Hooks returns the generation hooks that perform the 4NF validation.
func (*Extension) Hooks() []gen.Hook {
	return []gen.Hook{validate}
}

func validate(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		for _, n := range g.Nodes {
			if isJoinTable(n) {
				if err := validateJoinTable(n); err != nil {
					return err
				}
				continue
			}
			for _, e := range n.Edges {
				if e.Field() != nil {
					return fmt.Errorf(
						"4NF violation: entity %q has edge %q with .Field(%q); "+
							"move this foreign key to a join table schema annotated with fournf.JoinTable()",
						n.Name, e.Name, e.Field().Name,
					)
				}
			}
		}
		return next.Generate(g)
	})
}

// validateJoinTable checks that a schema annotated as a join table actually
// looks like one: every field must be a foreign key and there must be at least
// two foreign key edges.
func validateJoinTable(n *gen.Type) error {
	var fkEdges int
	for _, e := range n.Edges {
		if e.Field() != nil {
			fkEdges++
		}
	}
	if fkEdges < 2 {
		return fmt.Errorf(
			"4NF violation: join table %q has %d foreign key edge(s), need at least 2",
			n.Name, fkEdges,
		)
	}
	for _, f := range n.Fields {
		if !f.IsEdgeField() {
			return fmt.Errorf(
				"4NF violation: join table %q has non-foreign-key field %q; "+
					"join tables may only contain foreign key fields",
				n.Name, f.Name,
			)
		}
	}
	return nil
}

func isJoinTable(n *gen.Type) bool {
	// Check typed annotation.
	for _, a := range n.Annotations {
		if a, ok := a.(Annotation); ok && a.IsJoinTable {
			return true
		}
	}
	// Also check the raw map representation (annotations loaded from schema).
	if m, ok := n.Annotations["FourNF"]; ok && m != nil {
		if raw, ok := m.(map[string]interface{}); ok {
			if v, ok := raw["is_join_table"]; ok {
				if b, ok := v.(bool); ok && b {
					return true
				}
			}
		}
	}
	return false
}
