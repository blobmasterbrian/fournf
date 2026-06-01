package fournf

import (
	"testing"

	"entgo.io/ent/entc/gen"
)

// ---------------------------------------------------------------------------
// TestCheckTimestampFieldGraph
// ---------------------------------------------------------------------------

func TestCheckTimestampFieldGraph(t *testing.T) {
	tests := []struct {
		name              string
		node              *gen.Type
		fieldName         string
		needUpdateDefault bool
		wantCount         int
		wantContains      []string
	}{
		{
			name:         "field missing",
			node:         makeNode("User"),
			fieldName:    "created_at",
			wantCount:    1,
			wantContains: []string{"missing required field"},
		},
		{
			name:      "no Default, UpdateDefault not needed",
			node:      withFields(makeNode("User"), makeField("created_at")),
			fieldName: "created_at",
			wantCount: 1,
			wantContains: []string{"must have a Default value"},
		},
		{
			name:      "Default set, UpdateDefault not needed",
			node:      withFields(makeNode("User"), makeTimestampField("created_at", false)),
			fieldName: "created_at",
			wantCount: 0,
		},
		{
			name:              "Default set, UpdateDefault needed but missing",
			node:              withFields(makeNode("User"), makeTimestampField("updated_at", false)),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantCount:         1,
			wantContains:      []string{"must have an UpdateDefault value"},
		},
		{
			name:              "both Default and UpdateDefault set",
			node:              withFields(makeNode("User"), makeTimestampField("updated_at", true)),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantCount:         0,
		},
		{
			name:              "no Default AND UpdateDefault needed — accumulates 2 violations",
			node:              withFields(makeNode("User"), makeField("updated_at")),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantCount:         2,
			wantContains:      []string{"must have a Default value", "must have an UpdateDefault value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := checkTimestampFieldGraph(tt.node, tt.fieldName, tt.needUpdateDefault)
			if len(violations) != tt.wantCount {
				t.Fatalf("got %d violation(s) %v, want %d", len(violations), violations, tt.wantCount)
			}
			for _, want := range tt.wantContains {
				found := false
				for _, v := range violations {
					if contains(v, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no violation contains %q; got %v", want, violations)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestValidateJoinTableGraph
// ---------------------------------------------------------------------------

func TestValidateJoinTableGraph(t *testing.T) {
	t.Run("0 edges", func(t *testing.T) {
		n := makeNode("UserRole")
		violations := validateJoinTableGraph(n)
		if len(violations) != 1 {
			t.Fatalf("got %d violations, want 1: %v", len(violations), violations)
		}
		if !contains(violations[0], "0 foreign key edge(s)") {
			t.Fatalf("unexpected violation: %s", violations[0])
		}
	})

	t.Run("0 FK edges with non-FK fields", func(t *testing.T) {
		n := makeNode("UserRole")
		withFields(n, makeField("extra1"), makeField("extra2"))
		violations := validateJoinTableGraph(n)
		// 1 for FK count + 2 for non-FK fields = 3
		if len(violations) != 3 {
			t.Fatalf("got %d violations, want 3: %v", len(violations), violations)
		}
		if !contains(violations[0], "0 foreign key edge(s)") {
			t.Errorf("violation[0] = %q, want FK count violation", violations[0])
		}
		if !contains(violations[1], `non-foreign-key field "extra1"`) {
			t.Errorf("violation[1] = %q, want extra1 violation", violations[1])
		}
		if !contains(violations[2], `non-foreign-key field "extra2"`) {
			t.Errorf("violation[2] = %q, want extra2 violation", violations[2])
		}
	})

	t.Run("struct-literal edges count as 0 FK", func(t *testing.T) {
		n := makeNode("UserRole")
		n.Edges = []*gen.Edge{{Name: "user"}, {Name: "role"}}
		violations := validateJoinTableGraph(n)
		if len(violations) < 1 {
			t.Fatal("expected at least 1 violation")
		}
		if !contains(violations[0], "0 foreign key edge(s)") {
			t.Fatalf("unexpected violation: %s", violations[0])
		}
	})
}
