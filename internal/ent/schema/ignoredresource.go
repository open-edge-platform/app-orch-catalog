// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// IgnoredResource table
type IgnoredResource struct {
	ent.Schema
}

// Fields defines endpoint columns
func (IgnoredResource) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Ignored resource name."),
		field.String("kind").
			Comment("Ignored resource kind."),
		field.String("namespace").
			Comment("Ignored resource namespace."),
	}
}

// Edges defines ignored resource relations
func (IgnoredResource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application_fk", Application.Type).
			Ref("ignored_resources").
			Unique().
			Required().
			Comment("Application can have 0 to many IgnoredResources"),
	}
}

func (IgnoredResource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "kind", "namespace").
			Edges("application_fk").
			Unique(),
	}
}
