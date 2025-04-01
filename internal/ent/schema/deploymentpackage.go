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

// DeploymentPackage table.
type DeploymentPackage struct {
	ent.Schema
}

// Mixin common columns
func (DeploymentPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields DeploymentPackage columns
func (DeploymentPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_uuid").
			Comment("UUID of the owner project.").
			Default("default"),
		field.String("version").
			Comment("Version of the Deployment Package. Used in combination with the name to identify a unique Deployment Package within the catalog."),
		field.Bool("is_deployed").
			Comment("Indicates whether Deployment Package is deployed and available. Cannot be deleted while true").
			Optional(),
		field.Bool("is_visible").
			Comment("Indicates whether Deployment Package should be seen by user. Should not be deployed while false").
			Optional(),
		field.Bool("allows_multiple_deployments").
			Comment("Indicates whether Deployment Package can be deployed multiple times in the same realm.").
			Optional(),
		field.String("kind").
			Comment("Deployment package kind; normal, addon, extension").
			Optional(),
	}
}

// Edges DeploymentPackage relations
func (DeploymentPackage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("deployment_profiles", DeploymentProfile.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Deployment Package contains 0 to many Deployment Profiles"),
		edge.To("applications", Application.Type).
			Comment("Deployment Package can refer to 0-many Applications"),
		edge.To("icon", Artifact.Type).
			Comment("Deployment Package can refer to an Artifact as Icon"),
		edge.To("thumbnail", Artifact.Type).
			Comment("Deployment Package can refer to an Artifact as Icon"),
		edge.To("default_profile", DeploymentProfile.Type).
			Unique().
			Comment("Default Deployment Profile to be used when deploying this Deployment Package"),
		edge.To("application_dependencies", ApplicationDependency.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Application Dependencies to use when deploying this Deployment Package"),
		edge.To("application_namespaces", ApplicationNamespace.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Application Namespaces to use/create for Applications when deploying this Deployment Package"),
		edge.To("namespaces", Namespace.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Namespaces to create when deploying this Deployment Package"),
		edge.To("extensions", Extension.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Extensions to use when deploying this Deployment Package"),
		edge.To("artifacts", ArtifactReference.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			Comment("Various artifacts for use as icon, thumbnail or extensions."),
	}
}

func (DeploymentPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_uuid", "name", "version").Unique(),
	}
}
