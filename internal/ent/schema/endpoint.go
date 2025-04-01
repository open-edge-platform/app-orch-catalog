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

// Endpoint table
type Endpoint struct {
	ent.Schema
}

// Fields defines endpoint columns
func (Endpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("service_name").
			Comment("Endpoint service name."),
		field.String("external_path").
			Comment("External endpoint path."),
		field.String("internal_path").
			Comment("Internal endpoint path."),
		field.String("scheme").
			Comment("Internal endpoint protocol scheme."),
		field.String("auth_type").
			Comment("Authentication type."),
		field.String("app_name").
			Optional().
			Comment("Application name."),
	}
}

// Edges defines endpoint relations
func (Endpoint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("extension_fk", Extension.Type).
			Ref("endpoints").
			Unique().
			Required().
			Comment("Extension can have 0 to many Endpoints"),
	}
}

func (Endpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("service_name").
			Edges("extension_fk").
			Unique(),
	}
}
