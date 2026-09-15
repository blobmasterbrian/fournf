// Package fournf provides an entc extension that enforces 4NF.
//
// Entity schemas must not have edges that use .Field() (which would place a foreign key
// column on the entity table). Only schemas annotated with [JoinTable] may do so.
//
// Because the extension knows which schemas are join tables, it can also draw the
// schema as a domain diagram: see [WithClassDiagram].
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
	requireTimestamps bool
	classDiagramPaths []string
}

// ExtensionOption configures an [Extension].
type ExtensionOption func(*Extension)

// WithTimestamps returns an [ExtensionOption] that requires every entity schema to have
// created_at and updated_at fields. Join tables are exempt, as are schemas
// annotated with [SkipTimestamps].
func WithTimestamps() ExtensionOption {
	return func(e *Extension) {
		e.requireTimestamps = true
	}
}

// WithClassDiagram returns an [ExtensionOption] that writes a Mermaid class
// diagram of the schema to each path after each successful code generation.
//
// The diagram shows the domain rather than the storage: each join table
// collapses into a single association between the two entities it connects,
// labelled with the join's name, and its multiplicity is read from the join's
// unique indexes.
//
// Each path's file extension selects its format. A ".md" path is written as
// Markdown with the diagram inside a fenced block, which GitHub renders
// inline; any other extension is written as plain Mermaid, which is what
// Mermaid tooling reads. Passing one of each gives a document to browse and a
// file to render:
//
//	fournf.WithClassDiagram("schema.mmd", "schema.md")
//
// A relative path resolves against the directory code generation runs in, and
// missing parent directories are created.
func WithClassDiagram(path string, more ...string) ExtensionOption {
	return func(e *Extension) {
		e.classDiagramPaths = append([]string{path}, more...)
	}
}

// NewExtension returns a new 4NF extension.
func NewExtension(opts ...ExtensionOption) (*Extension, error) {
	e := &Extension{}
	for _, o := range opts {
		o(e)
	}
	return e, nil
}

// Hooks returns the 4NF validation hook, followed by the class diagram hook
// when [WithClassDiagram] is set. Validation runs first, so a diagram is only
// ever written for a schema that satisfies 4NF.
func (e *Extension) Hooks() []gen.Hook {
	hooks := []gen.Hook{e.validate}
	if len(e.classDiagramPaths) > 0 {
		hooks = append(hooks, e.classDiagram)
	}
	return hooks
}

func (e *Extension) validate(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		for _, n := range g.Nodes {
			if isJoinTable(n) {
				if err := validateJoinTable(n); err != nil {
					return err
				}
				continue
			}
			for _, edge := range n.Edges {
				if edge.Field() != nil {
					return fmt.Errorf(
						"4NF violation: entity %q has edge %q with .Field(%q); "+
							"move this foreign key to a join table schema annotated with fournf.JoinTable()",
						n.Name, edge.Name, edge.Field().Name,
					)
				}
			}
		}
		if e.requireTimestamps {
			for _, n := range g.Nodes {
				if isJoinTable(n) || skipTimestamps(n) {
					continue
				}
				if err := checkTimestampField(n, "created_at", false); err != nil {
					return err
				}
				if err := checkTimestampField(n, "updated_at", true); err != nil {
					return err
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

func skipTimestamps(n *gen.Type) bool {
	// Check typed annotation.
	for _, a := range n.Annotations {
		if a, ok := a.(Annotation); ok && a.SkipTimestamps {
			return true
		}
	}
	// Also check the raw map representation (annotations loaded from schema).
	if m, ok := n.Annotations["FourNF"]; ok && m != nil {
		if raw, ok := m.(map[string]interface{}); ok {
			if v, ok := raw["skip_timestamps"]; ok {
				if b, ok := v.(bool); ok && b {
					return true
				}
			}
		}
	}
	return false
}

// checkTimestampField returns an error if the field is missing, has no Default,
// or (when needUpdateDefault is true) has no UpdateDefault.
func checkTimestampField(n *gen.Type, name string, needUpdateDefault bool) error {
	f := findField(n, name)
	if f == nil {
		return fmt.Errorf(
			"timestamp violation: entity %q is missing required field %q; "+
				"add it or annotate the schema with fournf.SkipTimestamps()",
			n.Name, name,
		)
	}
	if !f.Default {
		return fmt.Errorf(
			"timestamp violation: entity %q field %q must have a Default value",
			n.Name, name,
		)
	}
	if needUpdateDefault && !f.UpdateDefault {
		return fmt.Errorf(
			"timestamp violation: entity %q field %q must have an UpdateDefault value",
			n.Name, name,
		)
	}
	return nil
}

func findField(n *gen.Type, name string) *gen.Field {
	for _, f := range n.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}
