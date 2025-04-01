// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"io"
)

// ParseError is an error that occurs when parsing OCI URLs

type ParseError struct {
	URL string
	Msg string

	Err error
}

func (e *ParseError) Error() string {
	msg := fmt.Sprintf("%s %s", e.Msg, e.URL)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *ParseError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .URL -}}
URL: {{.URL}}
{{end}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
There is a problem with the OCI URL that you entered. Please make sure the URL
starts with OCI:// and is a properly formed URL.

Example: oci://registry-1.docker.io/bitnamicharts/wordpress
`
	verboseerror.WriteErrorTemplate("ParseError", errTemplate, wr, e)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ExtractError is an error that occurs while extracting files

type ExtractError struct {
	Filename string
	Msg      string

	Err error
}

func (e *ExtractError) Error() string {
	msg := fmt.Sprintf("%s %s", e.Msg, e.Filename)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *ExtractError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .Filename -}}
Filename: {{.Filename}}
{{end}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
The tool attempted to exctract a file from a tarball, but encountered an error.
Please check to make sure that the OCI URL given points to a valid helm chart.
If you believe the chart to be valid and this error persists, then please contact
the Orchestrator team to report the issue.
`
	verboseerror.WriteErrorTemplate("ExtractError", errTemplate, wr, e)
}

func (e *ExtractError) Unwrap() error {
	return e.Err
}

// FetchError is an error that occurs while fetching OCI assets

type FetchError struct {
	URL      string
	Host     string
	Artifact string
	Msg      string

	Err error
}

func (e *FetchError) Error() string {
	msg := fmt.Sprintf("%s %s", e.Msg, e.URL)
	if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *FetchError) Verbose(wr io.Writer) {
	errTemplate := `------------------------------------------------------------
A critical error was encountered
------------------------------------------------------------
{{ if .Msg -}}
Message:       {{.Msg}}
{{end -}}
{{- if .URL -}}
URL:           {{.URL}}
{{end}}
{{- if .Host -}}
Host:          {{.Host}}
{{end}}
{{- if .Artifact -}}
Artifact:      {{.Artifact}}
{{end}}
{{- if .Err -}}
Wrapped Error: {{.Err}}
{{end}}
There is a problem fetching the OCI URL that you provided. Please verify that the URL
is correct and the artifact exists. If this is a private registry, then you may need
to provide authentication using the -u and -p command line options. If this is a public
registry such as dockerhub then you may get rate-limited if you do not use -u and -p.
`
	verboseerror.WriteErrorTemplate("FetchError", errTemplate, wr, e)
}

func (e *FetchError) Unwrap() error {
	return e.Err
}
