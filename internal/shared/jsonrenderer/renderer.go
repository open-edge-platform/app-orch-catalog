// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package jsonrenderer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

// JSONFromProto contains the given interface object.
type JSONFromProto struct {
	Data proto.Message
}

var jsonContentType = []string{"application/json"}

// Render (JSONFromProto) marshals the given interface object and writes data with custom ContentType.
func (r JSONFromProto) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	bytes, err := protojson.Marshal(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

// WriteContentType (JSONFromProto) writes JSONFromProto ContentType.
func (r JSONFromProto) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
