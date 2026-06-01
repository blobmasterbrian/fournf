package fournf

import (
	"errors"
	"testing"

	"entgo.io/ent/entc/gen"
)

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

func makeNode(name string) *gen.Type {
	return &gen.Type{Name: name}
}

func withTypedAnnotation(n *gen.Type, a Annotation) *gen.Type {
	if n.Annotations == nil {
		n.Annotations = make(gen.Annotations)
	}
	n.Annotations.Set(a.Name(), a)
	return n
}

func withRawAnnotation(n *gen.Type, raw map[string]interface{}) *gen.Type {
	if n.Annotations == nil {
		n.Annotations = make(gen.Annotations)
	}
	n.Annotations["FourNF"] = raw
	return n
}

func withFields(n *gen.Type, fields ...*gen.Field) *gen.Type {
	n.Fields = append(n.Fields, fields...)
	return n
}

func makeField(name string) *gen.Field {
	return &gen.Field{Name: name}
}

func makeTimestampField(name string, updateDefault bool) *gen.Field {
	return &gen.Field{Name: name, Default: true, UpdateDefault: updateDefault}
}

func noopGenerator() (gen.Generator, *bool) {
	called := new(bool)
	g := gen.GenerateFunc(func(*gen.Graph) error {
		*called = true
		return nil
	})
	return g, called
}

// ---------------------------------------------------------------------------
// TestNewExtension
// ---------------------------------------------------------------------------

func TestNewExtension(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		ext, err := NewExtension()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ext.requireTimestamps {
			t.Fatal("requireTimestamps = true, want false")
		}
	})
	t.Run("WithTimestamps", func(t *testing.T) {
		ext, err := NewExtension(WithTimestamps())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ext.requireTimestamps {
			t.Fatal("requireTimestamps = false, want true")
		}
	})
}

// ---------------------------------------------------------------------------
// TestExtensionHooks
// ---------------------------------------------------------------------------

func TestExtensionHooks(t *testing.T) {
	ext, _ := NewExtension()
	hooks := ext.Hooks()
	if len(hooks) != 1 {
		t.Fatalf("len(Hooks()) = %d, want 1", len(hooks))
	}
}

// ---------------------------------------------------------------------------
// TestFindField
// ---------------------------------------------------------------------------

func TestFindField(t *testing.T) {
	n := makeNode("User")
	withFields(n, makeField("name"), makeField("email"))

	tests := []struct {
		name    string
		search  string
		wantNil bool
	}{
		{"found", "name", false},
		{"found second", "email", false},
		{"not found", "age", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := findField(n, tt.search)
			if (f == nil) != tt.wantNil {
				t.Fatalf("findField(%q) nil=%v, want nil=%v", tt.search, f == nil, tt.wantNil)
			}
		})
	}
	t.Run("empty fields", func(t *testing.T) {
		empty := makeNode("Empty")
		if f := findField(empty, "x"); f != nil {
			t.Fatal("expected nil for empty fields")
		}
	})
}

// ---------------------------------------------------------------------------
// TestIsJoinTable
// ---------------------------------------------------------------------------

func TestIsJoinTable(t *testing.T) {
	tests := []struct {
		name string
		node *gen.Type
		want bool
	}{
		{
			name: "no annotations",
			node: makeNode("User"),
			want: false,
		},
		{
			name: "typed annotation true",
			node: withTypedAnnotation(makeNode("UserRole"), Annotation{IsJoinTable: true}),
			want: true,
		},
		{
			name: "typed annotation false",
			node: withTypedAnnotation(makeNode("User"), Annotation{IsJoinTable: false}),
			want: false,
		},
		{
			name: "raw map true",
			node: withRawAnnotation(makeNode("UserRole"), map[string]interface{}{"is_join_table": true}),
			want: true,
		},
		{
			name: "raw map false",
			node: withRawAnnotation(makeNode("UserRole"), map[string]interface{}{"is_join_table": false}),
			want: false,
		},
		{
			name: "raw map missing key",
			node: withRawAnnotation(makeNode("UserRole"), map[string]interface{}{}),
			want: false,
		},
		{
			name: "raw map wrong value type",
			node: withRawAnnotation(makeNode("UserRole"), map[string]interface{}{"is_join_table": "yes"}),
			want: false,
		},
		{
			name: "raw map nil",
			node: func() *gen.Type {
				n := makeNode("X")
				n.Annotations = gen.Annotations{"FourNF": nil}
				return n
			}(),
			want: false,
		},
		{
			name: "SkipTimestamps only",
			node: withTypedAnnotation(makeNode("Audit"), Annotation{SkipTimestamps: true}),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJoinTable(tt.node); got != tt.want {
				t.Fatalf("isJoinTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestSkipTimestamps
// ---------------------------------------------------------------------------

func TestSkipTimestamps(t *testing.T) {
	tests := []struct {
		name string
		node *gen.Type
		want bool
	}{
		{
			name: "no annotations",
			node: makeNode("User"),
			want: false,
		},
		{
			name: "typed annotation true",
			node: withTypedAnnotation(makeNode("Audit"), Annotation{SkipTimestamps: true}),
			want: true,
		},
		{
			name: "typed annotation false",
			node: withTypedAnnotation(makeNode("Audit"), Annotation{SkipTimestamps: false}),
			want: false,
		},
		{
			name: "raw map true",
			node: withRawAnnotation(makeNode("Audit"), map[string]interface{}{"skip_timestamps": true}),
			want: true,
		},
		{
			name: "raw map false",
			node: withRawAnnotation(makeNode("Audit"), map[string]interface{}{"skip_timestamps": false}),
			want: false,
		},
		{
			name: "raw map missing key",
			node: withRawAnnotation(makeNode("Audit"), map[string]interface{}{}),
			want: false,
		},
		{
			name: "raw map wrong value type",
			node: withRawAnnotation(makeNode("Audit"), map[string]interface{}{"skip_timestamps": "yes"}),
			want: false,
		},
		{
			name: "raw map nil",
			node: func() *gen.Type {
				n := makeNode("X")
				n.Annotations = gen.Annotations{"FourNF": nil}
				return n
			}(),
			want: false,
		},
		{
			name: "IsJoinTable only",
			node: withTypedAnnotation(makeNode("JT"), Annotation{IsJoinTable: true}),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skipTimestamps(tt.node); got != tt.want {
				t.Fatalf("skipTimestamps() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestCheckTimestampField
// ---------------------------------------------------------------------------

func TestCheckTimestampField(t *testing.T) {
	tests := []struct {
		name              string
		node              *gen.Type
		fieldName         string
		needUpdateDefault bool
		wantErr           bool
		wantContains      string
	}{
		{
			name:         "field missing",
			node:         makeNode("User"),
			fieldName:    "created_at",
			wantErr:      true,
			wantContains: "missing required field",
		},
		{
			name:         "no Default",
			node:         withFields(makeNode("User"), makeField("created_at")),
			fieldName:    "created_at",
			wantErr:      true,
			wantContains: "must have a Default value",
		},
		{
			name:      "Default set, UpdateDefault not needed",
			node:      withFields(makeNode("User"), makeTimestampField("created_at", false)),
			fieldName: "created_at",
			wantErr:   false,
		},
		{
			name:              "Default set, UpdateDefault needed but missing",
			node:              withFields(makeNode("User"), makeTimestampField("updated_at", false)),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantErr:           true,
			wantContains:      "must have an UpdateDefault value",
		},
		{
			name:              "both Default and UpdateDefault set",
			node:              withFields(makeNode("User"), makeTimestampField("updated_at", true)),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantErr:           false,
		},
		{
			name:              "no Default, UpdateDefault needed — Default checked first",
			node:              withFields(makeNode("User"), makeField("updated_at")),
			fieldName:         "updated_at",
			needUpdateDefault: true,
			wantErr:           true,
			wantContains:      "must have a Default value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTimestampField(tt.node, tt.fieldName, tt.needUpdateDefault)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkTimestampField() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.wantContains != "" {
				if msg := err.Error(); !contains(msg, tt.wantContains) {
					t.Fatalf("error %q does not contain %q", msg, tt.wantContains)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestValidateJoinTable
// ---------------------------------------------------------------------------

func TestValidateJoinTable(t *testing.T) {
	t.Run("no edges", func(t *testing.T) {
		n := makeNode("UserRole")
		err := validateJoinTable(n)
		if err == nil {
			t.Fatal("expected error for join table with no edges")
		}
		if !contains(err.Error(), "0 foreign key edge(s)") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("struct-literal edges have nil Field", func(t *testing.T) {
		n := makeNode("UserRole")
		n.Edges = []*gen.Edge{{Name: "user"}, {Name: "role"}}
		err := validateJoinTable(n)
		if err == nil {
			t.Fatal("expected error: struct-literal edges return nil from Field()")
		}
		if !contains(err.Error(), "0 foreign key edge(s)") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// TestValidateHook
// ---------------------------------------------------------------------------

func TestValidateHook(t *testing.T) {
	errSentinel := errors.New("next error")

	tests := []struct {
		name       string
		ext        *Extension
		graph      *gen.Graph
		nextErr    error
		wantErr    bool
		wantCalled bool
		errContains string
	}{
		{
			name:       "empty graph passes",
			ext:        &Extension{},
			graph:      &gen.Graph{},
			wantCalled: true,
		},
		{
			name: "entity with no edges passes",
			ext:  &Extension{},
			graph: &gen.Graph{
				Nodes: []*gen.Type{makeNode("User")},
			},
			wantCalled: true,
		},
		{
			name: "join table with 0 FK edges errors",
			ext:  &Extension{},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withTypedAnnotation(makeNode("UserRole"), Annotation{IsJoinTable: true}),
				},
			},
			wantErr:     true,
			wantCalled:  false,
			errContains: "0 foreign key edge(s)",
		},
		{
			name: "timestamps required, missing created_at",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{makeNode("User")},
			},
			wantErr:     true,
			wantCalled:  false,
			errContains: `missing required field "created_at"`,
		},
		{
			name: "timestamps required, missing updated_at",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withFields(makeNode("User"),
						makeTimestampField("created_at", false),
					),
				},
			},
			wantErr:     true,
			wantCalled:  false,
			errContains: `missing required field "updated_at"`,
		},
		{
			name: "timestamps required, created_at no Default",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withFields(makeNode("User"),
						makeField("created_at"),
						makeTimestampField("updated_at", true),
					),
				},
			},
			wantErr:     true,
			wantCalled:  false,
			errContains: `"created_at" must have a Default`,
		},
		{
			name: "timestamps required, updated_at no UpdateDefault",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withFields(makeNode("User"),
						makeTimestampField("created_at", false),
						&gen.Field{Name: "updated_at", Default: true, UpdateDefault: false},
					),
				},
			},
			wantErr:     true,
			wantCalled:  false,
			errContains: `"updated_at" must have an UpdateDefault`,
		},
		{
			name: "timestamps required, all correct passes",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withFields(makeNode("User"),
						makeTimestampField("created_at", false),
						makeTimestampField("updated_at", true),
					),
				},
			},
			wantCalled: true,
		},
		{
			name: "SkipTimestamps annotation skips check",
			ext:  &Extension{requireTimestamps: true},
			graph: &gen.Graph{
				Nodes: []*gen.Type{
					withTypedAnnotation(makeNode("Audit"), Annotation{SkipTimestamps: true}),
				},
			},
			wantCalled: true,
		},
		{
			name: "next error propagated",
			ext:  &Extension{},
			graph: &gen.Graph{
				Nodes: []*gen.Type{makeNode("User")},
			},
			nextErr:    errSentinel,
			wantErr:    true,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, called := noopGenerator()
			if tt.nextErr != nil {
				next = gen.GenerateFunc(func(*gen.Graph) error {
					*called = true
					return tt.nextErr
				})
			}

			hook := tt.ext.validate(next)
			err := hook.Generate(tt.graph)

			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *called != tt.wantCalled {
				t.Fatalf("next called = %v, want %v", *called, tt.wantCalled)
			}
			if tt.wantErr && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
