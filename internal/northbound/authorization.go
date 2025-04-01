// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"google.golang.org/grpc/metadata"
)

func (g *Server) authCheckAllowed(ctx context.Context, request interface{}, customRequestName ...string) error {
	if g.opaClient == nil {
		log.Debugf("ignoring Authorization")
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.NewInvalidArgument(errors.WithMessage("authentication failed"))
	}
	opaInputStruct := openpolicyagent.OpaInput{
		Input: map[string]interface{}{
			"request":  request,
			"metadata": md,
		},
	}

	// can safely ignore the JSON error - will not happen with OPA data
	completeInputJSON, _ := json.Marshal(opaInputStruct)

	bodyReader := bytes.NewReader(completeInputJSON)

	// The name of the protobuf request is an easy way of linking to REGO rules e.g. "*catalogv3.CreatePublisherRequest"
	requestType := reflect.TypeOf(request).String()
	requestPackage := requestType[1:strings.LastIndex(requestType, ".")]
	requestName := requestType[strings.LastIndex(requestType, ".")+1:]

	if len(customRequestName) > 0 {
		requestName = customRequestName[0]
	}

	trueBool := true
	resp, err := g.opaClient.PostV1DataPackageRuleWithBodyWithResponse(
		ctx,
		requestPackage,
		requestName,
		&openpolicyagent.PostV1DataPackageRuleParams{
			Pretty:  &trueBool,
			Metrics: &trueBool,
		},
		"application/json",
		bodyReader)
	if err != nil {
		return err
	}

	resultBool, boolErr := resp.JSON200.Result.AsOpaResponseResult1()
	if boolErr != nil {
		resultObj, objErr := resp.JSON200.Result.AsOpaResponseResult0()
		if objErr != nil {
			log.Debugf("(#1) access denied by OPA rule %s: %v", requestName, objErr)
			return errors.NewPermissionDenied()
		}
		log.Debugf("(#2) access denied by OPA rule %s: %v", requestName, resultObj)
		return errors.NewPermissionDenied()

	}
	if resultBool {
		log.Debugf("%s Authorized", requestName)
		return nil
	}

	log.Debugf("access denied by OPA rule %s. OPA response %d %v", requestName, resp.StatusCode(), resp.HTTPResponse)
	return errors.NewPermissionDenied()
}
