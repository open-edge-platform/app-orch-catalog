// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"
)

// DeploymentProfile table
type DeploymentProfile struct {
	ent.Schema
}

// Mixin common columns
func (DeploymentProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields deployment profile columns
func (DeploymentProfile) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges deployment profile relations
func (DeploymentProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profiles", Profile.Type).
			Comment("A Deployment Package can have 0-many Deployment Profiles"),
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("deployment_profiles").
			Unique().
			Required().
			Comment("Deployment Profile must belong to a Deployment Package"),
	}
}

func (DeploymentProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Edges("deployment_package_fk").
			Unique(),
	}
}
