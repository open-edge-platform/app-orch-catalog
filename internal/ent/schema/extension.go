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

// Extension table
type Extension struct {
	ent.Schema
}

// Fields defines extension columns
func (Extension) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("A unique name. Used in HTTP network paths."),
		field.String("version").
			Comment("Application API version."),
		field.String("display_name").
			Comment("A display name.").
			Optional(),
		field.String("display_name_lc").
			Comment("Lowercase display name.").
			Optional(),
		field.String("description").
			Optional().
			Comment("A description. Displayed on user interfaces."),
		field.String("ui_label").
			Optional().
			Comment("UI display label."),
		field.String("ui_service_name").
			Optional().
			Comment("UI service name."),
		field.String("ui_description").
			Optional().
			Comment("UI description."),
		field.String("ui_file_name").
			Optional().
			Comment("UI file name."),
		field.String("ui_app_name").
			Optional().
			Comment("UI app name."),
		field.String("ui_module_name").
			Optional().
			Comment("UI module name."),
	}
}

// Edges defines extension relations
func (Extension) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("endpoints", Endpoint.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Extension can have 0 to many Endpoints"),
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("extensions").
			Unique().
			Required().
			Comment("Many Extensions can referenced by 0-many Deployment Packages"),
	}
}

func (Extension) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "version").
			Edges("deployment_package_fk").
			Unique(),
	}
}
