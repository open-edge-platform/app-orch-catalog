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

// Application table
type Application struct {
	ent.Schema
}

// Mixin common columns
func (Application) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields application columns
func (Application) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_uuid").
			Comment("UUID of the owner project.").
			Default("default"),
		field.String("version").
			Comment("Application version."),
		field.String("chart_name").
			Comment("A chart name."),
		field.String("chart_version").
			Comment("A chart version."),
		field.String("kind").
			Comment("Application kind; normal, addon, extension.").
			Optional(),
	}
}

// Edges application relations
func (Application) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profiles", Profile.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Application contains 0 to many Profiles"),
		edge.From("registry_fk", Registry.Type).
			Ref("applications").
			Unique().
			Required().
			Comment("Application must refer to a valid HELM Registry"),
		edge.From("image_registry_fk", Registry.Type).
			Ref("application_images").
			Unique().
			Comment("Application can also refer to a valid IMAGE Registry"),
		edge.From("deployment_package_fk", DeploymentPackage.Type).
			Ref("applications").
			Comment("Many Applications may be referenced by 0-many Deployment Packages"),
		edge.From("dependency_source_fk", ApplicationDependency.Type).
			Ref("source_fk").
			Comment("Application dependency source"),
		edge.From("dependency_target_fk", ApplicationDependency.Type).
			Ref("target_fk").
			Comment("Application dependency target"),
		edge.To("default_profile", Profile.Type).
			Unique().
			Comment("Default Profile to be used when deploying this Application"),
		edge.To("ignored_resources", IgnoredResource.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Resource to ignore when deploying this Application"),
	}
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_uuid", "name", "version").Unique(),
	}
}
