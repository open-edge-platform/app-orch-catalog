// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"database/sql"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	"regexp"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/open-edge-platform/orch-library/go/pkg/northbound"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
	"sync"
)

var log = dazl.GetPackageLogger()
var utilsLog = dazl.GetPackageLogger().WithSkipCalls(1)

// NewService returns a new catalog Service
func NewService(databaseClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) northbound.Service {
	return &Service{
		DatabaseClient: databaseClient,
		OpaClient:      opaClient,
	}
}

// Service is a Service implementation for administration.
type Service struct {
	DatabaseClient *ent.Client
	OpaClient      openpolicyagent.ClientWithResponsesInterface
}

// Register registers the Service with the gRPC server.
func (s Service) Register(r *grpc.Server) {

	server := &Server{
		databaseClient: s.DatabaseClient,
		opaClient:      s.OpaClient,
		uploadSessions: make(map[string]*uploadSession, 0),
		listeners:      NewEventListeners(),
	}

	catalogv3.RegisterCatalogServiceServer(r, server)
}

const (
	ActiveProjectID = "activeprojectid"
	AdminProjectID  = "default"
)

// GetActiveProjectID extracts ActiveProjectID metadata from the incoming context
func GetActiveProjectID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.NewInvalidArgument(errors.WithMessage("incomplete request: unable to fetch request metadata"))
	}

	values := md.Get(ActiveProjectID)
	// FIXME: Remove this comment when ready to enforce empty activeprojectid metadata
	//if len(values) == 0 || values[0] == "" {
	//	return "", errors.NewInvalidArgument(errors.WithMessage("incomplete request: missing activeprojectid metadata"))
	//}

	// FIXME: Remove this if clause when ready to enforce empty activeprojectid metadata
	if len(values) == 0 || values[0] == "" {
		return AdminProjectID, nil
	}
	return values[0], nil
}

// GetActiveProjectIDAllowAdmin extracts ActiveProjectID metadata from the incoming context and if it's an admin project, it will return the
// fallback project, if one is specified.
func GetActiveProjectIDAllowAdmin(ctx context.Context, fallbackProjectID string) (string, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return "", err
	}

	// If the context project is administrative one, use the project from the request, if specified.
	if projectUUID == AdminProjectID && fallbackProjectID != "" {
		return fallbackProjectID, nil
	}
	return projectUUID, nil
}

// Server implements the gRPC service for administrative facilities.
type Server struct {
	catalogv3.UnimplementedCatalogServiceServer
	databaseClient *ent.Client
	opaClient      openpolicyagent.ClientWithResponsesInterface

	lock           sync.RWMutex
	uploadSessions map[string]*uploadSession

	listeners *EventListeners
}

// NewServer creates a new server with the specified database client and OPA client entities.
func NewServer(dbClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) *Server {
	return &Server{
		UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
		databaseClient:                    dbClient,
		opaClient:                         opaClient,
	}
}

// Starts a new transaction or returns ready to punt error
func (g *Server) startTransaction(ctx context.Context) (*generated.Tx, error) {
	tx, err := g.databaseClient.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Commits the specified transaction or returns ready to punt error
func (g *Server) commitTransaction(tx *generated.Tx) error {
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Rolls back the specified transaction, absorbing any error
func (g *Server) rollbackTransaction(tx *generated.Tx) {
	_ = tx.Rollback()
}

func (g *Server) checkApplication(ctx context.Context, tx *generated.Tx, name, version, projectUUID string) error {
	ok, err := tx.Application.Query().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(name),
			application.Version(version)).
		Exist(ctx)
	if err != nil {
		return err
	} else if !ok {
		return errors.NewNotFound(errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(name),
			errors.WithResourceVersion(version))
	}
	return nil
}

//func (g *Server) checkDeploymentPackage(ctx context.Context, tx *generated.Tx, name, version, projectUUID string) error {
//	ok, err := tx.DeploymentPackage.Query().
//		Where(
//			deploymentpackage.ProjectUUID(projectUUID),
//			deploymentpackage.Name(name),
//			deploymentpackage.Version(version)).
//		Exist(ctx)
//	if err != nil {
//		return err
//	} else if !ok {
//		return errors.NewNotFound(errors.WithResourceType(errors.DeploymentPackageType),
//			errors.WithResourceName(name),
//			errors.WithResourceVersion(version))
//	}
//	return nil
//}

// Processes deletion results
func (g *Server) checkDeleteResult(ctx context.Context, tx *generated.Tx, err error, thing string, publisher string) (*emptypb.Empty, error) {
	if err != nil {
		if tx != nil {
			g.rollbackTransaction(tx)
		}
		return nil, err
	}

	if tx != nil {
		err = g.commitTransaction(tx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
	}

	logActivity(ctx, "deleted", thing, publisher)
	return &emptypb.Empty{}, nil
}

func logActivity(ctx context.Context, verb string, thing string, publisher string, args ...string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md.Get("name")) > 0 {
		utilsLog.Infof("User '%v' %s %s %s (publisher: %s) from %v",
			md.Get("name"), verb, thing, strings.Join(args, "/"), publisher, md.Get("client"))
	} else {
		utilsLog.Infof("Someone %s %s %s (publisher: %s) from %v",
			verb, thing, strings.Join(args, "/"), publisher, md.Get("client"))
	}
}

func findColumnName(attributeName string, columns map[string]string, resourceType errors.ResourceType, operation string) (string, error) {
	columnName, ok := columns[attributeName]
	if !ok {
		return "", errors.NewInvalidArgument(
			errors.WithResourceType(resourceType),
			errors.WithMessage("%s: no such attribute: %s", operation, attributeName))
	}
	if columnName == "" {
		return "", errors.NewInvalidArgument(
			errors.WithResourceType(resourceType),
			errors.WithMessage("%s: cannot %s on attribute: %s", operation, operation, attributeName))
	}
	return columnName, nil
}

type orderBy struct {
	name   string
	isDesc bool
}

func parseOrderBy(orderByParameter string, resourceType errors.ResourceType) ([]*orderBy, error) {
	if orderByParameter == "" {
		return nil, nil
	}
	elements := strings.Split(orderByParameter, ",")
	var orderBys []*orderBy
	for _, element := range elements {
		descending := false
		direction := strings.Split(strings.Trim(element, " "), " ")
		if len(direction) > 2 {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(resourceType),
				errors.WithMessage("invalid order by: %s", element))
		}
		if len(direction) == 2 {
			if direction[1] == "asc" {
				descending = false
			} else if direction[1] == "desc" {
				descending = true
			} else {
				return nil, errors.NewInvalidArgument(
					errors.WithResourceType(resourceType),
					errors.WithMessage("invalid order by: %s", element))
			}
		}
		orderBys = append(orderBys, &orderBy{
			name:   direction[0],
			isDesc: descending,
		})
	}
	return orderBys, nil
}

func (o *orderBy) orderByDirection() entsql.OrderTermOption {
	orderMap := map[bool]entsql.OrderTermOption{
		true:  entsql.OrderDesc(),
		false: entsql.OrderAsc(),
	}
	return orderMap[o.isDesc]
}

func orderByOptions(orderBys []*orderBy, columns map[string]string, resourceType errors.ResourceType) ([]func(selector *entsql.Selector), error) {
	var options []func(s *entsql.Selector)

	if len(orderBys) != 0 {
		for _, o := range orderBys {
			orderTermOption := o.orderByDirection()
			columnName, err := findColumnName(o.name, columns, resourceType, "orderBy")
			if err != nil {
				return nil, err
			}
			options = append(options, entsql.OrderByField(columnName, orderTermOption).ToFunc())
		}
	}
	return options, nil
}

type filter struct {
	name  string
	value string
}

func parseFilter(filterParameter string, resourceType errors.ResourceType) ([]*filter, error) {
	if filterParameter == "" {
		return nil, nil
	}
	normalizeEqualsRe := regexp.MustCompile("[ \t]*=[ \t]*")
	normalizedFilterParameter := normalizeEqualsRe.ReplaceAllString(filterParameter, "=")

	elements := strings.Split(normalizedFilterParameter, " ")
	var filters []*filter
	var currentFilter *filter

	for index, element := range elements {
		if strings.Contains(element, "=") {
			selectors := strings.Split(element, "=")
			if currentFilter != nil || len(selectors) != 2 || selectors[0] == "" || selectors[1] == "" {
				// Error condition - too many equals
				return nil, errors.NewInvalidArgument(
					errors.WithResourceType(resourceType),
					errors.WithMessage("filter: invalid filter request: %s", elements))
			}
			currentFilter = &filter{}
			// This is the start of a selector. Grab the name and the value
			currentFilter.name = selectors[0]
			currentFilter.value = selectors[1]
		} else if element == "OR" {
			if currentFilter == nil || index == len(elements)-1 {
				//  Error condition - OR with no other term
				return nil, errors.NewInvalidArgument(
					errors.WithResourceType(resourceType),
					errors.WithMessage("filter: invalid filter request: %s", elements))
			}
			filters = append(filters, currentFilter)
			currentFilter = nil
			continue
		} else {
			if currentFilter == nil {
				// Error condition - missing an =
				return nil, errors.NewInvalidArgument(
					errors.WithResourceType(resourceType),
					errors.WithMessage("filter: invalid filter request: %s", elements))
			}
			currentFilter.value = currentFilter.value + " " + element
		}
	}
	if currentFilter != nil {
		filters = append(filters, currentFilter)
	}

	return filters, nil
}

func filterPredicates(filters []*filter, columns map[string]string, resourceType errors.ResourceType) ([]func(s *entsql.Selector), error) {
	var preds []func(s *entsql.Selector)
	if len(filters) != 0 {
		for _, f := range filters {
			column, err := findColumnName(f.name, columns, resourceType, "filter")
			if err != nil {
				return nil, err
			}

			likeValue := "%" + strings.ToLower(strings.ReplaceAll(f.value, "*", "%")) + "%"
			likePred := func(s *entsql.Selector) {
				s.Where(entsql.Like(entsql.Lower(column), likeValue))
			}
			preds = append(preds, likePred)
		}
	}
	return preds, nil
}

func kindPredicate(kinds []catalogv3.Kind) func(s *entsql.Selector) {
	if len(kinds) > 0 {
		hasNormalKind := false
		kindDBs := make([]string, 0, len(kinds))
		for _, kind := range kinds {
			kindDBs = append(kindDBs, kindToDB(kind))
			hasNormalKind = hasNormalKind || kind == catalogv3.Kind_KIND_NORMAL
		}
		if hasNormalKind {
			return entsql.OrPredicates(entsql.FieldIsNull("kind"), entsql.FieldIn("kind", kindDBs...))
		}
		return entsql.FieldIn("kind", kindDBs...)
	}
	return nil
}

const MaxPageSize = 500
const DefaultPageSize = 20

func computePageRange(pageSize int32, offset int32, totalCount int) (int, int, int32, error) {
	if offset < 0 {
		return 0, 0, 0, errors.NewInvalidArgument(errors.WithMessage("invalid pagination: offset must not be negative"))
	}
	if pageSize < 0 {
		return 0, 0, 0, errors.NewInvalidArgument(errors.WithMessage("invalid pagination: pageSize must not be negative"))
	}
	if pageSize > MaxPageSize {
		return 0, 0, 0, errors.NewInvalidArgument(errors.WithMessage("invalid pagination: pageSize must not exceed %d", MaxPageSize))
	}

	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	startIndex := 0
	endIndex := 0

	if totalCount == 0 {
		return startIndex, endIndex, 0, nil
	}
	startIndex = int(offset)
	if pageSize == 0 {
		endIndex = totalCount - 1
	} else {
		if startIndex+int(offset) > totalCount || int(offset+pageSize) > totalCount {
			endIndex = totalCount - 1
		} else {
			endIndex = (startIndex + int(pageSize)) - 1
		}
	}
	return startIndex, endIndex, int32(totalCount), nil
}

// Validates the display name format, using name as a fallback
func validateDisplayName(name string, displayName string) (string, bool) {
	dn := name
	if displayName != "" {
		dn = displayName
		if displayName != strings.TrimSpace(displayName) {
			return dn, false
		}
	}
	return dn, true
}

// Checks whether the given results of query support uniqueness of display name.
func checkUniqueness(count int, err error, thing string, name string, displayName string, resourceType errors.ResourceType) error {
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	if count > 0 {
		return errors.NewAlreadyExists(
			errors.WithResourceType(resourceType),
			errors.WithResourceName(name),
			errors.WithMessage("%s %s display name %s is not unique", thing, name, displayName))
	}
	return nil
}
