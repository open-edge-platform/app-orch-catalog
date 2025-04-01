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

// ParameterTemplate table
type ParameterTemplate struct {
	ent.Schema
}

// Fields defines extension columns
func (ParameterTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("A unique name. Used in HTTP network paths."),
		field.String("display_name").
			Comment("A display name.").
			Optional(),
		field.String("display_name_lc").
			Comment("Lowercase display name.").
			Optional(),
		field.String("default").
			Optional().
			Comment("Default value for the parameter."),
		field.String("type").
			Optional().
			Comment("Type of parameter."),
		field.String("validator").
			Optional().
			Comment("Validator."),
		field.JSON("suggested_values", []string{}).
			Optional(),
		field.Bool("mandatory").
			Optional().
			Comment("Indicates a mandatory parameter."),
		field.Bool("secret").
			Optional().
			Comment("Indicates a secret parameter."),
	}
}

// Edges defines extension relations
func (ParameterTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("profile_fk", Profile.Type).
			Ref("parameter_templates").
			Unique().
			Required().
			Comment("Many parameter templates can be referenced by a profile"),
		edge.To("profiles", Profile.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Profile can have 0 to many ParameterTemplates"),
	}
}

func (ParameterTemplate) Indexes() []ent.Index {
	return []ent.Index{}
}
