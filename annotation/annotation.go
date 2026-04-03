// Package annotation provides schema annotations for 4NF enforcement.
package annotation

// JoinTable marks an Ent schema as a join table.
// Schemas annotated with JoinTable are permitted to have edges with .Field()
// (which place foreign key columns on the table). Entity schemas must NOT have such edges.
var JoinTable = joinTable{}

type joinTable struct{}

// Name implements the ent schema annotation interface.
func (joinTable) Name() string {
	return "JoinTable"
}
