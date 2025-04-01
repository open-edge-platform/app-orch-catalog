// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ArtifactReference table
type ArtifactReference struct {
	ent.Schema
}

// Fields defines artifact reference
func (ArtifactReference) Fields() []ent.Field {
	return []ent.Field{
		field.String("purpose").
			Comment("Purpose for the artifact."),
	}
}

// Edges defines artifact reference relations
func (ArtifactReference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("artifact", Artifact.Type).
			Unique().
			Required().
			Comment("Artifact being referred to."),
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("artifacts").
			Unique().
			Required().
			Comment("Many artifacts can referenced by 0-many Deployment Packages"),
	}
}
