// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
)

// DeploymentRequirement table.
type DeploymentRequirement struct {
	ent.Schema
}

// Edges DeploymentRequirement relations
func (DeploymentRequirement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("profile_fk", Profile.Type).
			Ref("deployment_requirements").
			Unique().
			Required().
			Comment("Deployment Requirement must belong to one Application Profile"),
		edge.To("deployment_package_fk", DeploymentPackage.Type).
			Required().
			Unique().
			Comment("Deployment requirement refers to a Deployment Package"),
		edge.To("deployment_profile_fk", DeploymentProfile.Type).
			Unique().
			Comment("Deployment requirement may refer to a Deployment Profile"),
	}
}
