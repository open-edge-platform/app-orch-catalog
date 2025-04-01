// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// ApplicationDependency table
type ApplicationDependency struct {
	ent.Schema
}

// Edges application relations
func (ApplicationDependency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("application_dependencies").
			Unique().
			Required().
			Comment("Application Dependency must belong to a Deployment Package"),
		edge.To("source_fk", Application.Type).
			Unique().
			Required().
			Comment("Source of application dependency"),
		edge.To("target_fk", Application.Type).
			Unique().
			Required().
			Comment("Target of application dependency"),
	}
}
