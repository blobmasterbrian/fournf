package fournf

import "entgo.io/ent/schema"

// Annotation holds 4NF metadata for an Ent schema.
type Annotation struct {
	// IsJoinTable marks the schema as a join table. Join tables must have at
	// least two foreign key edges and every field must be a foreign key.
	IsJoinTable bool `json:"is_join_table,omitempty"`
	// SkipTimestamps opts this schema out of the timestamp requirement
	// enforced by [WithTimestamps].
	SkipTimestamps bool `json:"skip_timestamps,omitempty"`
}

// Name implements the ent schema.Annotation interface.
func (Annotation) Name() string {
	return "FourNF"
}

// Merge implements the ent schema.Annotation interface.
func (a Annotation) Merge(other schema.Annotation) schema.Annotation {
	o, ok := other.(Annotation)
	if !ok {
		return a
	}
	if o.IsJoinTable {
		a.IsJoinTable = true
	}
	if o.SkipTimestamps {
		a.SkipTimestamps = true
	}
	return a
}

// JoinTable returns an Annotation that marks the schema as a join table.
func JoinTable() Annotation {
	return Annotation{IsJoinTable: true}
}

// SkipTimestamps returns an Annotation that opts the schema out of the
// timestamp requirement enforced by [WithTimestamps].
func SkipTimestamps() Annotation {
	return Annotation{SkipTimestamps: true}
}
