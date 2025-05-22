// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dptohelm

import (
	"fmt"
	"io"

	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
)

// OutputError is an error that occurs while generating output files

type UsageError struct {
	Msg   string
	Input string
	Err   error
}

func (e *UsageError) Error() string {
	msg := e.Msg
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *UsageError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .Input-}}
Input:         {{.Input}}
{{end -}}
{{- if .OutputFile -}}
File:          {{.OutputFile}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
Recommendation: Use -h to ensure that you are using the correct syntax.
`
	verboseerror.WriteErrorTemplate("UsageError", errTemplate, wr, e)
}

func (e *UsageError) Unwrap() error {
	return e.Err
}

// InputError is an error that occurs while reading input files

type NotFoundError struct {
	Msg             string
	ObjectKind      string
	ObjectName      string
	ObjectVersion   string
	ApplicationName string
	Err             error
}

func (e *NotFoundError) Error() string {
	msg := fmt.Sprintf("%s on object %s value %s: %s", e.Msg, e.ObjectKind, e.ObjectName, e.Err)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *NotFoundError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .ObjectKind -}}
Object:        {{.ObjectKind}}
{{end -}}
{{- if .ObjectName -}}
Name:          {{.ObjectName}}
{{end -}}
{{- if .ObjectVersion -}}
Version:       {{.ObjectVersion}}
{{end -}}
{{- if .ApplicationName -}}
Name:          {{.ApplicationName}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
This could indicate an internal inconsistency in the deployment package. Make sure the
applications named inside your Deployment Package match the applications in your Application
objects. Make sure versions are correct. Make sure profile names are correct. Make sure that
any dependent objects, such as registries, are present in the DP.
`
	verboseerror.WriteErrorTemplate("NotFoundError", errTemplate, wr, e)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

// OutputError is an error that occurs while generating output files

type OutputError struct {
	Msg        string
	OutputFile string
	OutputDir  string
	Err        error
}

func (e *OutputError) Error() string {
	msg := e.Msg
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *OutputError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .OutputDir -}}
Path:          {{.OutputDir}}
{{end -}}
{{- if .OutputFile -}}
File:          {{.OutputFile}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
Recommendation: Please check that the output path exists, is writable, has
enough space and that you have sufficient permission to write files in that
directory.
`
	verboseerror.WriteErrorTemplate("outputError", errTemplate, wr, e)
}

func (e *OutputError) Unwrap() error {
	return e.Err
}

// DirectoryError is an error that occurs while reading input files

type DirectoryError struct {
	InputDir string
	Msg      string
	Err      error
}

func (e *DirectoryError) Error() string {
	msg := fmt.Sprintf("%s while processing directory %s", e.Msg, e.InputDir)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *DirectoryError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .InputDir -}}
Dir:          {{.InputDir}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
Recommendation: Please ensure that the directory exists and has a valid Deployment
Package in it. There should be only one Deployment Package in the directory.
`
	verboseerror.WriteErrorTemplate("DirectoryError", errTemplate, wr, e)
}

func (e *DirectoryError) Unwrap() error {
	return e.Err
}
