// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Profile table
type Profile struct {
	ent.Schema
}

// Mixin common columns
func (Profile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields profile columns
func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.String("chart_values"),
	}
}

// Edges profile relations
func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application_fk", Application.Type).
			Ref("profiles").
			Unique().
			Required().
			Comment("Profile must belong to one Application"),
		edge.From("deployment_profiles", DeploymentProfile.Type).
			Ref("profiles").
			Comment("Many Deployment Profiles can refer to an Application Profile"),
		edge.To("parameter_templates", ParameterTemplate.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Profile may contain a list of parameter templates."),
		edge.To("deployment_requirements", DeploymentRequirement.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Profile may depend on a set of Deployment Requirements."),
	}
}

func (Profile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Edges("application_fk").
			Unique(),
	}
}
