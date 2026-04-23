// Package annotation provides schema annotations for 4NF enforcement.
package annotation

// JoinTable marks an Ent schema as a join table.
// Annotated schemas must have at least two foreign key edges and every field
// must be a foreign key. Entity schemas must NOT have edges with .Field().
type JoinTable struct{}

// Name implements the ent schema annotation interface.
func (JoinTable) Name() string {
	return "JoinTable"
}
