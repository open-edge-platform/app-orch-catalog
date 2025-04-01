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

// Registry table
type Registry struct {
	ent.Schema
}

// Mixin common columns
func (Registry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields registry columns
func (Registry) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_uuid").
			Comment("UUID of the owner project.").
			Default("default"),
		field.String("auth_token").
			Comment("A login token for registry access.").
			Optional(),
		field.String("type").
			Comment("Registry type (helm or image)."),
		field.String("api_type").
			Comment("Registry API type.").
			Optional(),
	}
}

// Edges registry relations
func (Registry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("applications", Application.Type).
			Comment("Many Applications can refer to a HELM Registry"),
		edge.To("application_images", Application.Type).
			Comment("Many Applications can refer to an IMAGE Registry"),
	}
}

func (Registry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_uuid", "name").Unique(),
	}
}
