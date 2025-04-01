// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dp

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"io"
)

// OutputError is an error that occurs while generating output files

type OutputError struct {
	Helm       helm.HelmInfo
	OutputDir  string
	OutputFile string
	Msg        string
	Err        error
}

func (e *OutputError) Error() string {
	msg := fmt.Sprintf("%s in path %s", e.Msg, e.OutputDir)
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

// InputError is an error that occurs while reading input files

type InputError struct {
	Helm      helm.HelmInfo
	InputFile string
	Msg       string
	Err       error
}

func (e *InputError) Error() string {
	msg := fmt.Sprintf("%s while reading %s", e.Msg, e.InputFile)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *InputError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .InputFile -}}
File:          {{.InputFile}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
Recommendation: Please check that the files that you have specified exist and
are readable by the current user.
`
	verboseerror.WriteErrorTemplate("InputError", errTemplate, wr, e)
}

func (e *InputError) Unwrap() error {
	return e.Err
}

// GenerationError is an error that occurs while generating yaml

type GenerationError struct {
	Helm helm.HelmInfo
	Msg  string
	Err  error
}

func (e *GenerationError) Error() string {
	msg := e.Msg
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *GenerationError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
These errors are caused internally by unexpected input that the tool was not
able to process correctly. Please check your input to ensure that it was
valid. Please also send an error report the Orchestrator staff letting them
know the arguments you supplied to the tool, so that we can investigate the
problem.
`
	verboseerror.WriteErrorTemplate("GenerationError", errTemplate, wr, e)
}

func (e *GenerationError) Unwrap() error {
	return e.Err
}
