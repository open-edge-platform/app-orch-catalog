// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ApplicationNamespace table
type ApplicationNamespace struct {
	ent.Schema
}

// Fields application namespace columns
func (ApplicationNamespace) Fields() []ent.Field {
	return []ent.Field{
		field.String("namespace").
			Comment("Application namespace."),
	}
}

// Edges application namespace relations
func (ApplicationNamespace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("application_namespaces").
			Unique().
			Required().
			Comment("Application Namespace must belong to a Deployment Package"),
		edge.To("source_fk", Application.Type).
			Unique().
			Required().
			Comment("Source application for which this namespace applies"),
	}
}
