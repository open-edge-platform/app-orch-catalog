// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package jsonrenderer

import (
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http/httptest"
	"testing"
)

func TestJsonFromProto_Render(t *testing.T) {
	tests := []struct {
		name string
		data proto.Message
		want []string
	}{
		{
			name: "render",
			data: &catalogv3.UploadMultipleCatalogEntitiesResponse{Responses: []*catalogv3.UploadCatalogEntitiesResponse{
				{
					SessionId:     "123",
					ErrorMessages: nil,
				},
				{
					SessionId:     "123",
					ErrorMessages: []string{"some error"},
				},
			}},
			// camelCase keys are defined in the protobuf annotation,
			// while the JSON annotation contains a snake_case format
			// we are testing that the protobuf are converted to JSON using the protobuf annotation
			want: []string{"sessionId", "errorMessages"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := JSONFromProto{
				Data: tt.data,
			}
			w := httptest.NewRecorder()
			err := r.Render(w)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}
			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			for _, s := range tt.want {
				assert.Contains(t, string(body), s)
			}
		})
	}
}
