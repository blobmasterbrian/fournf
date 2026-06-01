package fournf

import (
	"testing"

	"entgo.io/ent/schema"
)

func TestAnnotationName(t *testing.T) {
	if got := (Annotation{}).Name(); got != "FourNF" {
		t.Fatalf("Name() = %q, want %q", got, "FourNF")
	}
}

func TestJoinTableConstructor(t *testing.T) {
	a := JoinTable()
	if !a.IsJoinTable {
		t.Fatal("JoinTable().IsJoinTable = false, want true")
	}
	if a.SkipTimestamps {
		t.Fatal("JoinTable().SkipTimestamps = true, want false")
	}
}

func TestSkipTimestampsConstructor(t *testing.T) {
	a := SkipTimestamps()
	if !a.SkipTimestamps {
		t.Fatal("SkipTimestamps().SkipTimestamps = false, want true")
	}
	if a.IsJoinTable {
		t.Fatal("SkipTimestamps().IsJoinTable = true, want false")
	}
}

func TestAnnotationMerge(t *testing.T) {
	tests := []struct {
		name           string
		base           Annotation
		other          schema.Annotation
		wantJoinTable  bool
		wantSkipTS     bool
	}{
		{
			name:  "both false unchanged",
			base:  Annotation{},
			other: Annotation{},
		},
		{
			name:          "merge IsJoinTable into false",
			base:          Annotation{},
			other:         Annotation{IsJoinTable: true},
			wantJoinTable: true,
		},
		{
			name:       "merge SkipTimestamps into false",
			base:       Annotation{},
			other:      Annotation{SkipTimestamps: true},
			wantSkipTS: true,
		},
		{
			name:          "merge both true",
			base:          Annotation{},
			other:         Annotation{IsJoinTable: true, SkipTimestamps: true},
			wantJoinTable: true,
			wantSkipTS:    true,
		},
		{
			name:          "OR semantics: true + false stays true",
			base:          Annotation{IsJoinTable: true, SkipTimestamps: true},
			other:         Annotation{},
			wantJoinTable: true,
			wantSkipTS:    true,
		},
		{
			name:  "non-Annotation type returns base unchanged",
			base:  Annotation{IsJoinTable: true},
			other: nonAnnotation{},
			wantJoinTable: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.Merge(tt.other).(Annotation)
			if got.IsJoinTable != tt.wantJoinTable {
				t.Errorf("IsJoinTable = %v, want %v", got.IsJoinTable, tt.wantJoinTable)
			}
			if got.SkipTimestamps != tt.wantSkipTS {
				t.Errorf("SkipTimestamps = %v, want %v", got.SkipTimestamps, tt.wantSkipTS)
			}
		})
	}
}

// nonAnnotation is a dummy schema.Annotation for testing Merge with a wrong type.
type nonAnnotation struct{}

func (nonAnnotation) Name() string { return "Other" }
