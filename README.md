# fournf

An [Ent](https://entgo.io) extension that enforces **fourth normal form (4NF)**: all relationships must go through explicit join table schemas. No foreign key columns on entity tables.

## Install

```sh
go get github.com/blobmasterbrian/fournf
```

## Usage

### 1. Wire the extension into `entc.go`

```go
//go:build ignore

package main

import (
    "log"

    "entgo.io/ent/entc"
    "entgo.io/ent/entc/gen"
    "github.com/blobmasterbrian/fournf"
)

func main() {
    err := entc.Generate("./schema",
        &gen.Config{},
        entc.Extensions(fournf.Extension{}),
    )
    if err != nil {
        log.Fatalf("running ent codegen: %v", err)
    }
}
```

`go generate` will now fail if any schema violates 4NF.

### 2. Define entity schemas and join table schemas

Entity schemas define your domain models. Join table schemas wire relationships
between them and are annotated with `annotation.JoinTable{}`.

```go
// schema/species.go
package schema

type Species struct { ent.Schema }

func (Species) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
    }
}

func (Species) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("habitats", Habitat.Type).
            Ref("species"),
    }
}
```

```go
// schema/habitat.go
package schema

type Habitat struct { ent.Schema }

func (Habitat) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"), // e.g. "Rainforest", "Tundra"
    }
}

func (Habitat) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("species", Species.Type).
            Ref("habitats"),
    }
}
```

```go
// schema/species_habitat.go — join table
package schema

import "github.com/blobmasterbrian/fournf/annotation"

type SpeciesHabitat struct { ent.Schema }

func (SpeciesHabitat) Annotations() []schema.Annotation {
    return []schema.Annotation{
        annotation.JoinTable{},
    }
}

func (SpeciesHabitat) Fields() []ent.Field {
    return []ent.Field{
        field.Int("species_id"),
        field.Int("habitat_id"),
    }
}

func (SpeciesHabitat) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("species", Species.Type).
            Unique().
            Required().
            Field("species_id"),
        edge.To("habitat", Habitat.Type).
            Unique().
            Required().
            Field("habitat_id"),
    }
}
```

### What gets flagged

**Foreign keys on entity tables.** If an entity schema places a foreign key
directly on its own table via `.Field()`, code generation will fail:

```go
// schema/species.go — BAD: foreign key on an entity table
package schema

type Species struct { ent.Schema }

func (Species) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.Int("habitat_id"), // foreign key lives on the entity table
    }
}

func (Species) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("habitat", Habitat.Type).
            Unique().
            Field("habitat_id"), // this triggers the violation
    }
}
```

```
4NF violation: entity "Species" has edge "habitat" with .Field("habitat_id");
move this foreign key to a join table schema annotated with annotation.JoinTable
```

**Non-foreign-key fields on join tables.** Join tables may only contain foreign
key fields. Adding extra columns defeats the purpose of the join table:

```go
// schema/species_habitat.go — BAD: extra field on a join table
package schema

type SpeciesHabitat struct { ent.Schema }

func (SpeciesHabitat) Annotations() []schema.Annotation {
    return []schema.Annotation{annotation.JoinTable{}}
}

func (SpeciesHabitat) Fields() []ent.Field {
    return []ent.Field{
        field.Int("species_id"),
        field.Int("habitat_id"),
        field.String("notes"), // not a foreign key
    }
}
```

```
4NF violation: join table "SpeciesHabitat" has non-foreign-key field "notes";
join tables may only contain foreign key fields
```

**Join tables with fewer than two foreign key edges.** A join table must link at
least two entities:

```go
// schema/species_habitat.go — BAD: only one foreign key edge
package schema

type SpeciesHabitat struct { ent.Schema }

func (SpeciesHabitat) Annotations() []schema.Annotation {
    return []schema.Annotation{annotation.JoinTable{}}
}

func (SpeciesHabitat) Fields() []ent.Field {
    return []ent.Field{
        field.Int("species_id"),
    }
}

func (SpeciesHabitat) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("species", Species.Type).
            Unique().
            Required().
            Field("species_id"),
    }
}
```

```
4NF violation: join table "SpeciesHabitat" has 1 foreign key edge(s), need at least 2
```

### 3. CI test (optional)

For a safety net independent of code generation:

```go
func TestFourNF(t *testing.T) {
    fournftest.ValidateGraph(t, "./schema", "mymodule/ent")
}
```

## How it works

In Ent, calling `.Field()` on an edge places a foreign key column on the schema's table. 4NF requires that multi-valued dependencies are factored into separate tables. This extension enforces that rule at two levels:

1. **Entity schemas** must not have edges with `.Field()`. All foreign keys must live in dedicated join table schemas.
2. **Join table schemas** (annotated with `annotation.JoinTable{}`) must have at least two foreign key edges, and every field must be a foreign key. This prevents misuse of the annotation to bypass the entity restriction.

## License

MIT
