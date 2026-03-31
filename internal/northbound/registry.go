// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/predicate"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/registry"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/nberrors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	helmType  = "HELM"
	imageType = "IMAGE"

	// Registries with this value for CA certs trigger use of a dynamically loaded CA
	dynamicCACertsName = "use-dynamic-cacert"

	// System registries that are managed by app-orch-tenant-controller
	// NOTE: Ensure app-orch-tenant-controller and app-orch-catalog are updated at the same time,
	//       if these registry names change.
	SystemRegistryHarborHelm   = "harbor-helm-oci"
	SystemRegistryHarborDocker = "harbor-docker-oci"
)

var (
	systemRegistries = []string{
		SystemRegistryHarborHelm,
		SystemRegistryHarborDocker,
	}
)

// validateRegistryName validates the registry name according to the specified rules
func validateRegistryName(name string) error {
	if len(name) < 1 {
		return fmt.Errorf("registry name must be at least 1 character long")
	}
	if len(name) > 40 {
		return fmt.Errorf("registry name must be at most 40 characters long")
	}

	// Pattern: ^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$
	// This means: start with alphanumeric, middle can have hyphens, end with alphanumeric (if more than 1 char)
	matched, err := regexp.MatchString("^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$", name)
	if err != nil {
		return fmt.Errorf("error validating registry name pattern: %v", err)
	}
	if !matched {
		return fmt.Errorf("registry name must match pattern: start with a-z0-9, can contain hyphens, end with a-z0-9")
	}

	return nil
}

type RegistrySecretData interface {
	RootURL() string
	InventoryURL() string
	Username() string
	AuthToken() string
	Cacerts() string
}

type registrySecretData struct {
	RootURL      string
	InventoryURL string
	Username     string
	AuthToken    string
	Cacerts      string
}

type Base64Strings interface {
	EncodeBase64(r registrySecretData) string
	DecodeBase64(r *registrySecretData, encodedData string) error
}

type base64Strings struct{}

func (b *base64Strings) EncodeBase64(r registrySecretData) string {
	dataBlob, _ := json.Marshal(r) //nolint:gosec // G117: intentional marshaling of credential data for base64 encoding
	return base64.URLEncoding.EncodeToString(dataBlob)
}

func (b *base64Strings) DecodeBase64(r *registrySecretData, encodedData string) error {
	decodedBytes, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(decodedBytes)).Decode(r)
}

func newBase64() Base64Strings {
	bs := &base64Strings{}
	return bs
}

func readTLSCert() (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	value, err := clientSet.CoreV1().Secrets("orch-gateway").Get(context.Background(), os.Getenv("TLS_CERT_NAME"), metav1.GetOptions{})
	if err != nil {
		log.Errorf("Can't read secret %v", err)
		return "", nil
	}

	CABytes, ok := value.Data["tls.crt"]
	if !ok {
		return "", fmt.Errorf("unable to find TLS CA certificate")
	}
	return string(CABytes), nil
}

var UseSecretService = false
var VaultServer = VaultServerAddress
var Base64Factory = newBase64

func MakeSecretPath(projectUUID string, registryName string) string {
	return "cat-" + projectUUID + "_" + registryName
}

// CreateRegistry creates a Registry from gRPC request
func (g *Server) CreateRegistry(ctx context.Context, req *catalogv3.CreateRegistryRequest) (*catalogv3.CreateRegistryResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Registry == nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	} else if err := protovalidate.Validate(req); err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage(err.Error()))
	}

	// Additional validation for registry name
	if err := validateRegistryName(req.Registry.Name); err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage(err.Error()))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	events := &RegistryEvents{}
	created, err := g.createRegistry(ctx, tx, projectUUID, req.Registry, events)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	logActivity(ctx, "created", "registry", projectUUID, req.Registry.Name)
	events.sendToAll(g.listeners)

	return &catalogv3.CreateRegistryResponse{
		Registry: &catalogv3.Registry{
			Name:         created.Name,
			Description:  created.Description,
			DisplayName:  created.DisplayName,
			RootUrl:      req.Registry.RootUrl,
			InventoryUrl: req.Registry.InventoryUrl,
			Username:     req.Registry.Username,
			AuthToken:    req.Registry.AuthToken,
			Cacerts:      req.Registry.Cacerts,
			Type:         created.Type,
			ApiType:      req.Registry.ApiType,
			CreateTime:   timestamppb.New(created.CreateTime),
		},
	}, nil
}

func (g *Server) createRegistry(ctx context.Context, tx *generated.Tx, projectUUID string, reg *catalogv3.Registry, events *RegistryEvents) (*generated.Registry, error) {
	displayName, ok := validateDisplayName(reg.Name, reg.DisplayName)
	if !ok {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("display name cannot contain leading or trailing spaces"))
	}
	reg.DisplayName = displayName

	// Make sure that the display name, if specified is unique
	if err := g.checkRegistryUniqueness(ctx, tx, projectUUID, reg); err != nil {
		return nil, err
	}

	create := tx.Registry.Create().
		SetProjectUUID(projectUUID).
		SetName(reg.Name).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(reg.Description).
		SetType(reg.Type)

	registrySecret := &registrySecretData{
		RootURL:      reg.RootUrl,
		InventoryURL: reg.InventoryUrl,
		Username:     reg.Username,
		AuthToken:    reg.AuthToken,
		Cacerts:      reg.Cacerts,
	}

	registrySecretData := Base64Factory().EncodeBase64(*registrySecret)
	if !UseSecretService {
		create.SetAuthToken(registrySecretData)
	}

	created, err := create.Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, nberrors.NewInvalidArgument(
				nberrors.WithResourceType(nberrors.RegistryType),
				nberrors.WithResourceName(reg.Name),
				nberrors.WithMessage(`registry %s already exists`, reg.Name))
		}
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}
	if UseSecretService {
		secretService, err := SecretServiceFactory(ctx)
		if err != nil {
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
		defer secretService.Logout(ctx)

		registryKey := MakeSecretPath(projectUUID, reg.Name)
		err = secretService.WriteSecret(ctx, registryKey, registrySecretData)
		if err != nil {
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
	}

	events.append(CreatedEvent, projectUUID, reg)
	return created, nil
}

// Returns an error if the registry display name is not unique
func (g *Server) checkRegistryUniqueness(ctx context.Context, tx *generated.Tx, projectUUID string, r *catalogv3.Registry) error {
	if len(r.DisplayName) > 0 {
		existing, err := tx.Registry.Query().
			Where(
				registry.ProjectUUID(projectUUID),
				registry.DisplayNameLc(strings.ToLower(r.DisplayName)),
				registry.Not(registry.Name(r.Name))).
			Count(ctx)
		if err = checkUniqueness(existing, err, "registry", r.Name, r.DisplayName, nberrors.RegistryType); err != nil {
			return err
		}
	}
	return nil
}

func (g *Server) IsSystemRegistry(name string) bool {
	for _, r := range systemRegistries {
		if r == name {
			return true
		}
	}
	return false
}

// ListRegistries gets a list of all registries through gRPC
func (g *Server) ListRegistries(ctx context.Context, req *catalogv3.ListRegistriesRequest) (*catalogv3.ListRegistriesResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	}

	requestName := "ListRegistriesRequest"
	if req.ShowSensitiveInfo {
		requestName = "ListRegistriesWithSensitiveInfoRequest"
	}
	if err := g.authCheckAllowed(ctx, req, requestName); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	orderBys, err := parseOrderBy(req.OrderBy, nberrors.RegistryType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}
	filters, err := parseFilter(req.Filter, nberrors.RegistryType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}
	registries, _, count, err := g.getRegistries(ctx, tx, projectUUID, req.ShowSensitiveInfo, orderBys, filters, req.PageSize, req.Offset)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}
	logActivity(ctx, "listed", "registries", projectUUID, "Total => "+fmt.Sprintf("%d", count))
	return &catalogv3.ListRegistriesResponse{
			Registries:    registries,
			TotalElements: count},
		nil
}

var registryColumns = map[string]string{
	"project_uuid": "",
	"name":         "name",
	"displayName":  "display_name",
	"description":  "description",
	"createTime":   "create_time",
	"updateTime":   "update_time",
	"type":         "type",
	"rootUrl":      "",
	"inventoryUrl": "",
	"username":     "",
	"authToken":    "",
	"cacerts":      "",
}

func (g *Server) getRegistries(ctx context.Context, tx *generated.Tx, projectUUID string, showSensitiveInfo bool,
	orderBys []*orderBy, filters []*filter,
	pageSize int32, offset int32) ([]*catalogv3.Registry, []string, int32, error) {
	var err error
	var registriesDB []*generated.Registry
	var orderOptions []registry.OrderOption

	var secretService SecretService
	if UseSecretService {
		secretService, err = SecretServiceFactory(ctx)
		if err != nil {
			return nil, nil, 0, nberrors.NewVaultError(nberrors.WithError(err))
		}
		defer secretService.Logout(ctx)
	}

	registriesQuery := tx.Registry.Query()
	options, err := orderByOptions(orderBys, registryColumns, nberrors.RegistryType)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, pred := range options {
		orderOptions = append(orderOptions, pred)
	}
	registriesQuery = registriesQuery.Order(orderOptions...)

	filterPreds, err := filterPredicates(filters, registryColumns, nberrors.RegistryType)
	if err != nil {
		return nil, nil, 0, err
	}
	var registryPreds []predicate.Registry
	for _, pred := range filterPreds {
		registryPreds = append(registryPreds, pred)
	}
	registriesQuery = registriesQuery.Where(registry.Or(registryPreds...))

	if projectUUID == "" {
		registriesDB, err = registriesQuery.All(ctx)
	} else {
		registriesDB, err = registriesQuery.Where(registry.ProjectUUID(projectUUID)).All(ctx)
	}
	if err != nil {
		return nil, nil, 0, nberrors.NewDBError(nberrors.WithError(err))
	}

	startIndex, endIndex, totalElements, err := computePageRange(pageSize, offset, len(registriesDB))
	if err != nil {
		return nil, nil, 0, err
	}
	if len(registriesDB) == 0 {
		return []*catalogv3.Registry{}, []string{}, 0, nil
	}

	registries := make([]*catalogv3.Registry, 0)
	projectUUIDs := make([]string, 0)
	for i := startIndex; i <= endIndex; i++ {
		registryDB := registriesDB[i]
		reg, err := g.extractRegistry(ctx, registryDB, secretService, showSensitiveInfo)
		if err != nil {
			return nil, nil, 0, err
		}
		registries = append(registries, reg)
		projectUUIDs = append(projectUUIDs, registryDB.ProjectUUID)
	}
	return registries, projectUUIDs, totalElements, nil
}

func (g *Server) extractRegistry(ctx context.Context, registryDB *generated.Registry, secretService SecretService, showSensitiveInfo bool) (*catalogv3.Registry, error) {
	rsd := registrySecretData{}
	var encodedSecretData string
	var err error

	// Transient cache of the dynamically loaded CA certs
	dynamicCACert := ""

	if UseSecretService {
		registryKey := MakeSecretPath(registryDB.ProjectUUID, registryDB.Name)

		// Fetch the stored secret
		encodedSecretData, err = secretService.ReadSecret(ctx, registryKey)
		if err != nil {
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
	} else {
		encodedSecretData = registryDB.AuthToken
	}
	err = Base64Factory().DecodeBase64(&rsd, encodedSecretData)
	if err != nil {
		return nil, nberrors.NewVaultError(nberrors.WithError(err))
	}
	reg := &catalogv3.Registry{
		Name:         registryDB.Name,
		DisplayName:  registryDB.DisplayName,
		Description:  registryDB.Description,
		RootUrl:      rsd.RootURL,
		InventoryUrl: rsd.InventoryURL,
		Type:         registryDB.Type,
		ApiType:      registryDB.APIType,
		CreateTime:   timestamppb.New(registryDB.CreateTime),
		UpdateTime:   timestamppb.New(registryDB.UpdateTime),
	}
	if showSensitiveInfo {
		reg.Username = rsd.Username
		reg.AuthToken = rsd.AuthToken
		reg.Cacerts = rsd.Cacerts

		if rsd.Cacerts == dynamicCACertsName {
			if len(dynamicCACert) == 0 {
				dynamicCACert, err = readTLSCert()
				if err != nil {
					return nil, nberrors.NewVaultError(nberrors.WithError(err))
				}
			}
			reg.Cacerts = dynamicCACert
		}
	}
	return reg, nil
}

// GetRegistry gets a single registry through gRPC
func (g *Server) GetRegistry(ctx context.Context, req *catalogv3.GetRegistryRequest) (*catalogv3.GetRegistryResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.RegistryName == "" {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	}

	requestName := "GetRegistryRequest"
	if req.ShowSensitiveInfo {
		requestName = "GetRegistryWithSensitiveInfoRequest"
	}
	if err := g.authCheckAllowed(ctx, req, requestName); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	registryDB, err := tx.Registry.Query().
		Where(registry.ProjectUUID(projectUUID), registry.Name(req.RegistryName)).First(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		if generated.IsNotFound(err) {
			return nil, nberrors.NewNotFound(
				nberrors.WithResourceType(nberrors.RegistryType),
				nberrors.WithResourceName(req.RegistryName))
		}
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	var secretService SecretService
	if UseSecretService {
		secretService, err = SecretServiceFactory(ctx)
		if err != nil {
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
		defer secretService.Logout(ctx)
	}

	reg, err := g.extractRegistry(ctx, registryDB, secretService, req.ShowSensitiveInfo)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}
	logActivity(ctx, "got", "registry", projectUUID, req.RegistryName)
	return &catalogv3.GetRegistryResponse{Registry: reg}, nil
}

// UpdateRegistry updates a registry through gRPC
func (g *Server) UpdateRegistry(ctx context.Context, req *catalogv3.UpdateRegistryRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Registry == nil || req.RegistryName == "" {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	}

	// Validate registry name
	if err := validateRegistryName(req.RegistryName); err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage(err.Error()))
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage(err.Error()))
	}

	if req.RegistryName != req.Registry.Name {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("name cannot be changed %s != %s", req.RegistryName, req.Registry.Name))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	events := &RegistryEvents{}
	if err = g.updateRegistry(ctx, tx, projectUUID, req.Registry, events); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	logActivity(ctx, "updated", "registry", projectUUID, req.RegistryName)
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

func (g *Server) updateRegistry(ctx context.Context, tx *generated.Tx, projectUUID string, reg *catalogv3.Registry, events *RegistryEvents) error {
	displayName, ok := validateDisplayName(reg.Name, reg.DisplayName)
	if !ok {
		return nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithResourceName(reg.Name),
			nberrors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkRegistryUniqueness(ctx, tx, projectUUID, reg); err != nil {
		return err
	}

	// Make sure the registry type cannot mutate once the registry is used by an application.
	if err := g.checkRegistryType(ctx, tx, projectUUID, reg); err != nil {
		return err
	}

	update := tx.Registry.Update().
		Where(registry.ProjectUUID(projectUUID), registry.Name(reg.Name)).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(reg.GetDescription()).
		SetType(reg.Type).
		SetAPIType(reg.ApiType)

	registrySecret := &registrySecretData{
		RootURL:      reg.RootUrl,
		InventoryURL: reg.InventoryUrl,
		Username:     reg.Username,
		AuthToken:    reg.AuthToken,
		Cacerts:      reg.Cacerts,
	}
	registrySecretData := Base64Factory().EncodeBase64(*registrySecret)
	if !UseSecretService {
		update.SetAuthToken(registrySecretData)
	}
	updateCount, err := update.Save(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nberrors.NewDBError(nberrors.WithError(err))
	} else if updateCount == 0 {
		g.rollbackTransaction(tx)
		return nberrors.NewNotFound(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithResourceName(reg.Name),
			nberrors.WithMessage(`registry not found`))
	}
	if UseSecretService {
		registryKey := MakeSecretPath(projectUUID, reg.Name)
		secretService, err := SecretServiceFactory(ctx)
		if err != nil {
			return nberrors.NewVaultError(nberrors.WithError(err))
		}
		defer secretService.Logout(ctx)

		err = secretService.WriteSecret(ctx, registryKey, registrySecretData)
		if err != nil {
			return nberrors.NewVaultError(nberrors.WithError(err))
		}
	}
	events.append(UpdatedEvent, projectUUID, reg)
	return nil
}

func (g *Server) checkRegistryType(ctx context.Context, tx *generated.Tx, projectUUID string, reg *catalogv3.Registry) error {
	regDB, err := tx.Registry.Query().
		Where(registry.ProjectUUID(projectUUID), registry.Name(reg.Name)).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nberrors.NewNotFound(
				nberrors.WithResourceType(nberrors.RegistryType),
				nberrors.WithResourceName(reg.Name))
		}
		return nberrors.NewDBError(nberrors.WithError(err))
	}
	if regDB.Type != reg.Type {
		count := 0
		switch reg.Type {
		case helmType:
			count, err = tx.Registry.QueryApplicationImages(regDB).Count(ctx)
		case imageType:
			count, err = tx.Registry.QueryApplications(regDB).Count(ctx)
		}
		if err != nil {
			return nberrors.NewDBError(nberrors.WithError(err))
		}
		if count > 0 {
			return nberrors.NewFailedPrecondition(
				nberrors.WithResourceType(nberrors.RegistryType),
				nberrors.WithResourceName(reg.Name),
				nberrors.WithMessage("cannot change registry type to %s", reg.Type))
		}
	}
	return nil
}

// DeleteRegistry deletes a registry through gRPC
func (g *Server) DeleteRegistry(ctx context.Context, req *catalogv3.DeleteRegistryRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.RegistryName == "" {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}

	events := &RegistryEvents{}
	uses, err := tx.Application.Query().Where(
		application.ProjectUUID(projectUUID),
		application.Or(
			application.HasImageRegistryFkWith(registry.Name(req.RegistryName)),
			application.HasRegistryFkWith(registry.Name(req.RegistryName)))).
		Count(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	}
	if uses > 0 {
		g.rollbackTransaction(tx)
		return nil, nberrors.NewFailedPrecondition(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithResourceName(req.RegistryName),
			nberrors.WithMessage("cannot delete registry while in use"))
	}

	deleteCount, err := tx.Registry.Delete().
		Where(registry.ProjectUUID(projectUUID), registry.Name(req.RegistryName)).Exec(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, nberrors.NewDBError(nberrors.WithError(err))
	} else if deleteCount == 0 {
		g.rollbackTransaction(tx)
		return nil, nberrors.NewNotFound(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithResourceName(req.RegistryName))
	}

	var secretService SecretService
	if UseSecretService {
		registryKey := MakeSecretPath(projectUUID, req.RegistryName)
		secretService, err = SecretServiceFactory(ctx)
		if err != nil {
			g.rollbackTransaction(tx)
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
		defer secretService.Logout(ctx)

		err = secretService.DeleteSecret(ctx, registryKey)
		if err != nil {
			log.Warnf("failed to delete key %s from secret service: %v", registryKey, err)
			return nil, nberrors.NewVaultError(nberrors.WithError(err))
		}
	}
	if _, err = g.checkDeleteResult(ctx, tx, err, fmt.Sprintf("registry %s", req.RegistryName), projectUUID); err != nil {
		return nil, err
	}
	events.append(DeletedEvent, projectUUID, &catalogv3.Registry{Name: req.RegistryName})
	events.sendToAll(g.listeners)
	logActivity(ctx, "deleted", "registry", projectUUID, req.RegistryName)
	return &emptypb.Empty{}, nil
}

// WatchRegistries watches inventory of registries for changes.
func (g *Server) WatchRegistries(req *catalogv3.WatchRegistriesRequest, server catalogv3.CatalogService_WatchRegistriesServer) error {
	if server == nil {
		return nberrors.NewInvalidArgument(
			nberrors.WithMessage("incomplete request"))
	}
	projectUUID, err := GetActiveProjectIDAllowAdmin(server.Context(), req.ProjectId)
	if err != nil {
		return err
	}
	if req == nil {
		return nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.RegistryType),
			nberrors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(server.Context(), req); err != nil {
		return err
	}

	ch := make(chan *catalogv3.WatchRegistriesResponse)

	// If replay requested
	if !req.NoReplay {
		// Get list of registries
		ctx := server.Context()
		tx, err := g.startTransaction(ctx)
		if err != nil {
			return nberrors.NewDBError(nberrors.WithError(err))
		}

		registries, projectUUIDs, _, err := g.getRegistries(ctx, tx, projectUUID, req.ShowSensitiveInfo, nil, nil, 0, 0)
		if err != nil {
			g.rollbackTransaction(tx)
			return err
		}

		events := &RegistryEvents{}
		for i, reg := range registries {
			events.append(ReplayedEvent, projectUUIDs[i], reg)
		}

		// Send each replay event to the stream
		for _, e := range events.queue {
			if err = server.Send(e); err != nil {
				g.rollbackTransaction(tx)
				return err
			}
		}

		// Register the stream, so it can start receiving updates
		g.listeners.addRegistryListener(ch, req)

		err = g.commitTransaction(tx)
		if err != nil {
			return nberrors.NewDBError(nberrors.WithError(err))
		}
	} else {
		// Register the stream, so it can start receiving updates
		g.listeners.addRegistryListener(ch, req)
	}
	defer g.listeners.deleteRegistryListener(ch)
	logActivity(server.Context(), "watched", "registries", projectUUID, "")
	return g.watchRegistryEvents(server, ch)
}

func (g *Server) watchRegistryEvents(server catalogv3.CatalogService_WatchRegistriesServer, ch chan *catalogv3.WatchRegistriesResponse) error {
	for e := range ch {
		if err := server.Send(e); err != nil {
			return err
		}
	}
	return nil
}
