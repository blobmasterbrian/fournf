// Package testschema is a small 4NF schema used by fournf's tests. It covers
// one join of each multiplicity, a join across three entities, and an edge that
// carries no foreign key.
package testschema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/blobmasterbrian/fournf"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_profile", UserProfile.Type).Ref("user"),
		edge.From("user_sessions", UserSession.Type).Ref("user"),
		// An edge with no .Field() is permitted on an entity, and is drawn
		// directly rather than through a join table.
		edge.To("favorites", Article.Type),
	}
}

type Profile struct{ ent.Schema }

func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("bio"),
	}
}

type Session struct{ ent.Schema }

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("token").NotEmpty(),
	}
}

type Article struct{ ent.Schema }

func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").NotEmpty(),
	}
}

func (Article) Edges() []ent.Edge {
	// The inverse of User.favorites. A bidirectional edge must be drawn once,
	// not once per side.
	return []ent.Edge{
		edge.From("fans", User.Type).Ref("favorites"),
	}
}

type Tag struct{ ent.Schema }

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("label").NotEmpty(),
	}
}

// UserProfile is one-to-one: both foreign keys are unique.
type UserProfile struct{ ent.Schema }

func (UserProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{fournf.JoinTable()}
}

func (UserProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("profile_id", uuid.UUID{}),
	}
}

func (UserProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Unique().Required().Field("user_id"),
		edge.To("profile", Profile.Type).Unique().Required().Field("profile_id"),
	}
}

func (UserProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
		index.Fields("profile_id").Unique(),
	}
}

// UserSession is one-to-many: only the session side is unique, so a user holds
// many sessions and a session belongs to one user.
type UserSession struct{ ent.Schema }

func (UserSession) Annotations() []schema.Annotation {
	return []schema.Annotation{fournf.JoinTable()}
}

func (UserSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("session_id", uuid.UUID{}),
	}
}

func (UserSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Unique().Required().Field("user_id"),
		edge.To("session", Session.Type).Unique().Required().Field("session_id"),
	}
}

func (UserSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("session_id").Unique(),
	}
}

// ArticleTag is many-to-many: neither foreign key is unique on its own.
type ArticleTag struct{ ent.Schema }

func (ArticleTag) Annotations() []schema.Annotation {
	return []schema.Annotation{fournf.JoinTable()}
}

func (ArticleTag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("article_id", uuid.UUID{}),
		field.UUID("tag_id", uuid.UUID{}),
	}
}

func (ArticleTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("article", Article.Type).Unique().Required().Field("article_id"),
		edge.To("tag", Tag.Type).Unique().Required().Field("tag_id"),
	}
}

func (ArticleTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("article_id", "tag_id").Unique(),
	}
}

// ArticleTagUser joins three entities, so there is no single association it
// could collapse into and it stays a class in the diagram.
type ArticleTagUser struct{ ent.Schema }

func (ArticleTagUser) Annotations() []schema.Annotation {
	return []schema.Annotation{fournf.JoinTable()}
}

func (ArticleTagUser) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("article_id", uuid.UUID{}),
		field.UUID("tag_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
	}
}

func (ArticleTagUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("article", Article.Type).Unique().Required().Field("article_id"),
		edge.To("tag", Tag.Type).Unique().Required().Field("tag_id"),
		edge.To("user", User.Type).Unique().Required().Field("user_id"),
	}
}
