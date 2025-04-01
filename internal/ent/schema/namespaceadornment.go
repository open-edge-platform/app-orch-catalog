// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// NamespaceAdornment table
type NamespaceAdornment struct {
	ent.Schema
}

// Fields namespace adornment columns
func (NamespaceAdornment) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			Comment("Adornment type: label or annotation."),
		field.String("key").
			Comment("Adornment key."),
		field.String("value").
			Optional().
			Comment("Adornment value."),
	}
}

// Edges namespace adornment relations
func (NamespaceAdornment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace_fk", Namespace.Type).
			Ref("adornments").
			Unique().
			Required().
			Comment("Adornment must belong to a Namespace"),
	}
}
