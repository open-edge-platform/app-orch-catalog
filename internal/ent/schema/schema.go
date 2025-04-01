// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// File updated by protoc-gen-ent.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// CommonMixin adds common fields across entities.
type CommonMixin struct{ ent.Schema }

// Fields common columns
// Common fields used by several entities.
//
//	See https://entgo.io/docs/schema-mixin.
//	See https://cloud.google.com/apis/design/standard_fields.
func (CommonMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("A unique name. Used in HTTP network paths."),
		field.String("display_name").
			Comment("A display name.").
			Optional(),
		field.String("display_name_lc").
			Comment("Lowercase display name.").
			Optional(),
		field.String("description").
			Optional().
			Comment("A description. Displayed on user interfaces."),
		field.Time("create_time").
			Default(time.Now).
			Immutable().
			Comment("The creation timestamp."),
		field.Time("update_time").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("The last update timestamp."),
	}
}

// Ensure CommonMixin implements the `Mixin` interface.
var _ ent.Mixin = (*CommonMixin)(nil)
