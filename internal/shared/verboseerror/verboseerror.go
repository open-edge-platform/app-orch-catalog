// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// verboseerror handles printing verbose error messages and other related
// tasks.

package verboseerror

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

// Quiet will suppress informational messages

var Quiet bool

// Infof prints an informational message to stdout

func Infof(format string, args ...interface{}) {
	if !Quiet {
		fmt.Printf(format, args...)
	}
}

// Fatalf prints a fatal error to stderr and then exists

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// FatalErrCheck checks the error and prints a formatted message to stderr if the
// error is not nil. It then prints the error and exits.

func FatalErrCheck(err error) {
	if err != nil {
		PrintVerboseError(err)
		os.Exit(1)
	}
}

// VerboseError is an error with a Verbose method

type VerboseError interface {
	Verbose(io.Writer)
}

// PrintVerboseError uses the error's Verbose method to print a detailed error message
// to stderr. If the error does not implement Verbose, it prints a standard error
// message.

func PrintVerboseError(err error) {
	if err != nil {
		if e, ok := err.(VerboseError); ok {
			e.Verbose(os.Stderr)
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

// WriteErrorTemplate renders an error template to the writer, using the provided
// template name and contents. The template is rendered with the provided error
// interface. If the template fails to render, the function will exit with a fatal
// error.

func WriteErrorTemplate(templateName string, templateContents string, wr io.Writer, e interface{}) {
	err := template.Must(template.New(templateName).Parse(templateContents)).Execute(wr, e)

	/* if we failed to render the template, then fatal exit */
	FatalErrCheck(fmt.Errorf("Failed to render error template: %v", err))
}
