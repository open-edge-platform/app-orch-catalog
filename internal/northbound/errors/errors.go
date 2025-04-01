// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = dazl.GetPackageLogger().WithSkipCalls(2)

type Options struct {
	Code            codes.Code
	ResourceType    ResourceType
	ResourceName    string
	ResourceVersion string
	Message         string
	Error           error
	Args            []any
}

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*Options)

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

func WithCode(code codes.Code) Option {
	return func(opts *Options) {
		opts.Code = code
	}
}

type ResourceType string

const (
	ApplicationType          ResourceType = "application"
	ApplicationReferenceType ResourceType = "application-reference"
	ArtifactType             ResourceType = "artifact"
	DeploymentPackageType    ResourceType = "deployment-package"
	DeploymentProfileType    ResourceType = "deployment-profile"
	ProfileType              ResourceType = "profile"
	PublisherType            ResourceType = "publisher"
	RegistryType             ResourceType = "registry"
	UploadSession            ResourceType = "upload-session"
)

func WithResourceType(resourceType ResourceType) Option {
	return func(opts *Options) {
		opts.ResourceType = resourceType
	}
}

func WithResourceName(resourceName string) Option {
	return func(opts *Options) {
		opts.ResourceName = resourceName
	}
}

func WithResourceVersion(resourceVersion string) Option {
	return func(opts *Options) {
		opts.ResourceVersion = resourceVersion
	}
}

func WithMessage(message string, args ...any) Option {
	return func(opts *Options) {
		opts.Message = message
		opts.Args = args
	}
}

func WithError(err error) Option {
	return func(opts *Options) {
		opts.Error = err
	}
}

func NewInvalidArgument(opts ...Option) error {
	return newCodedError(codes.InvalidArgument, "invalid", opts...)
}

func NewNotFound(opts ...Option) error {
	return newCodedError(codes.NotFound, "not found", opts...)
}

func NewUnavailable(opts ...Option) error {
	return newCodedError(codes.Unavailable, "unavailable", opts...)
}

func NewAlreadyExists(opts ...Option) error {
	return newCodedError(codes.AlreadyExists, "already exists", opts...)
}

func NewFailedPrecondition(opts ...Option) error {
	return newCodedError(codes.FailedPrecondition, "failed precondition", opts...)
}

func NewPermissionDenied(opts ...Option) error {
	return newCodedError(codes.PermissionDenied, "access denied", opts...)
}

func NewInternal(opts ...Option) error {
	return newCodedError(codes.Internal, "error", opts...)
}

func NewDBError(opts ...Option) error {
	return newCodedError(codes.Internal, "an internal database error occurred", opts...)
}

func NewVaultError(opts ...Option) error {
	return newCodedError(codes.Internal, "failed to access secret service", opts...)
}

func newCodedError(code codes.Code, message string, opts ...Option) error {
	var options Options
	options.apply(opts...)
	options.Code = code
	if options.Message != "" {
		options.Message = message + ": " + options.Message
	} else {
		options.Message = message
	}
	if options.Error != nil {
		log.Error(options.Error)
	}
	return New(WithOptions(options))
}

func New(opts ...Option) error {
	var options Options
	options.apply(opts...)
	return newError(options)
}

func newError(options Options) error {
	builder := &strings.Builder{}
	var args []any
	if options.ResourceType != "" {
		builder.WriteString("%s ")
		args = append(args, options.ResourceType)
	}
	if options.ResourceName != "" {
		builder.WriteString("%s")
		args = append(args, options.ResourceName)
		if options.ResourceVersion != "" {
			builder.WriteString(":%s")
			args = append(args, options.ResourceVersion)
		}
		builder.WriteString(" ")
	}
	if options.Message != "" {
		builder.WriteString(options.Message)
		args = append(args, options.Args...)
	}
	return status.Errorf(options.Code, builder.String(), args...)
}
