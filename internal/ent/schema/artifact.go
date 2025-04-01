// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Artifact table
type Artifact struct {
	ent.Schema
}

// Mixin common columns
func (Artifact) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

// Fields arfifact columns
func (Artifact) Fields() []ent.Field {
	return []ent.Field{
		field.String("project_uuid").
			Comment("UUID of the owner project.").
			Default("default"),
		field.String("mime_type").
			Comment("MIME type of artifact."),
		field.Bytes("artifact").
			Comment("bytes containing an image or other digital media."),
	}
}

// Edges artifact relations
func (Artifact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ca_icon_fk", DeploymentPackage.Type).
			Ref("icon").
			Comment("An Artifact may be referenced as icon by 0-many Deployment Packages"),
		edge.From("ca_thumbnail_fk", DeploymentPackage.Type).
			Ref("thumbnail").
			Comment("An Artifact may be referenced as thumbnail by 0-many Deployment Packages"),
	}
}

func (Artifact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_uuid", "name").Unique(),
	}
}
