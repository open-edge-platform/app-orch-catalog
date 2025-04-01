// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Namespace table
type Namespace struct {
	ent.Schema
}

// Fields application namespace columns
func (Namespace) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Namespace name."),
	}
}

// Edges application namespace relations
func (Namespace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("namespaces").
			Unique().
			Required().
			Comment("Namespace must belong to a Deployment Package"),
		edge.To("adornments", NamespaceAdornment.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Namespace adornments"),
	}
}
