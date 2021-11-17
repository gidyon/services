package project

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	project "github.com/gidyon/services/pkg/api/project/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type Options struct {
	AuthAPI auth.API
	SqlDb   *gorm.DB
	Logger  grpclog.LoggerV2
}

type projectAPIServer struct {
	*Options
	project.UnimplementedProjectAPIServer
}

func NewProjectAPI(ctx context.Context, opt *Options) (_ project.ProjectAPIServer, err error) {
	defer func() {
		if err != nil {
			err = errs.WrapErrorWithMsgFunc("Failed to start account service")(err)
		}
	}()

	// Validation
	switch {
	case ctx == nil:
		err = errors.New("missing context")
	case opt == nil:
		err = errors.New("missing options")
	case opt.AuthAPI == nil:
		err = errors.New("missing authentication API")
	case opt.SqlDb == nil:
		err = errors.New("missing sql writes db")
		err = errors.New("missing redis reads db")
	case opt.Logger == nil:
		err = errors.New("missing logger")
	}
	if err != nil {
		return nil, err
	}

	projectAPI := &projectAPIServer{
		Options: opt,
	}

	// Do automigration
	if !projectAPI.SqlDb.Migrator().HasTable((&Project{}).TableName()) {
		err = projectAPI.SqlDb.AutoMigrate(&Project{})
		if err != nil {
			return nil, fmt.Errorf("failed to automigrate %s table: %v", (&Project{}).TableName(), err)
		}
	}

	return projectAPI, nil
}

func ValidateProject(pb *project.Project) error {
	switch {
	case pb == nil:
		return errs.NilObject("project")
	case pb.OwnerId == "":
		return errs.MissingField("project name")
	case pb.ProjectName == "":
		return errs.MissingField("project name")
	}
	return nil
}

func (projectAPI *projectAPIServer) CreateProject(
	ctx context.Context, req *project.CreateProjectRequest,
) (*project.Project, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.NilObject("request")
	default:
		err = ValidateProject(req.Project)
		if err != nil {
			return nil, err
		}
	}

	db, err := ProjectModel(req.Project)
	if err != nil {
		return nil, err
	}

	// Create in the database
	err = projectAPI.SqlDb.Create(db).Error
	if err == nil {
		return nil, errs.FailedToSave("project", err)
	}

	return ProjectProto(db)
}

func (projectAPI *projectAPIServer) UpdateProject(
	ctx context.Context, req *project.UpdateProjectRequest,
) (*project.Project, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.NilObject("request")
	case req.Project == nil:
		return nil, errs.MissingField("project resource")
	case req.Project.ProjectId == "":
		return nil, errs.MissingField("project id")
	}

	db, err := ProjectModel(req.Project)
	if err != nil {
		return nil, err
	}

	// Create in the database
	err = projectAPI.SqlDb.Where("id=?", req.Project.ProjectId).Updates(db).Error
	if err == nil {
		return nil, errs.FailedToSave("project", err)
	}

	return ProjectProto(db)
}

func (projectAPI *projectAPIServer) GetProject(
	ctx context.Context, req *project.GetProjectRequest,
) (*project.Project, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.ProjectId == "":
		return nil, errs.MissingField("project id")
	}

	db := &Project{}

	err = projectAPI.SqlDb.First(db, "id=?", req.ProjectId).Error
	if err != nil {
		return nil, errs.FailedToFind("project", err)
	}

	return ProjectProto(db)
}

func (projectAPI *projectAPIServer) DeleteProject(
	ctx context.Context, req *project.DeleteProjectRequest,
) (*empty.Empty, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.ProjectId == "":
		return nil, errs.MissingField("project id")
	}

	err = projectAPI.SqlDb.Delete("id=?", req.ProjectId).Error
	if err != nil {
		return nil, errs.FailedToDelete("project", err)
	}

	return &emptypb.Empty{}, nil
}

const defaultPageSize = 100

type projectDst struct {
	ProjectId string
}

func (projectAPI *projectAPIServer) ListProjects(
	ctx context.Context, req *project.ListProjectsRequest,
) (*project.ListProjectsResponse, error) {
	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("list request")
	}

	// Get payload
	actor, err := projectAPI.AuthAPI.GetJwtPayload(ctx)
	if err != nil {
		return nil, err
	}

	// Authorization
	if !projectAPI.AuthAPI.IsAdmin(actor.Group) {
		if req.Filter == nil {
			req.Filter = &project.ListProjectsFilter{}
		} else {
			req.Filter.OwnerIds = make([]string, 0, 1)
		}
		if actor.ID != "" {
			req.Filter.OwnerIds = append(req.Filter.OwnerIds, actor.ID)
		}
	}

	pageSize := req.GetPageSize()
	switch {
	case pageSize <= 0:
		pageSize = defaultPageSize
	case pageSize > defaultPageSize:
		if !projectAPI.AuthAPI.IsAdmin(actor.Group) {
			pageSize = defaultPageSize
		}
	}

	var ID uint
	pageToken := req.GetPageToken()
	if pageToken != "" {
		bs, err := base64.StdEncoding.DecodeString(req.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		v, err := strconv.ParseUint(string(bs), 10, 64)
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "incorrect page token")
		}
		ID = uint(v)
	}

	db := projectAPI.SqlDb.Unscoped().Model(&Project{}).Limit(int(pageSize + 1)).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	// Apply tv filters
	if req.Filter != nil {
		if len(req.Filter.OwnerIds) != 0 {
			db = db.Where("owner_id IN (?)", req.Filter.OwnerIds)
		}
		if len(req.Filter.OwnerIds) != 0 {
			projects := make([]*projectDst, 0, 2)
			// Get member project list
			err = projectAPI.SqlDb.Model(&ProjectMember{}).Select("project_id").Where("user_id IN ?", req.Filter.OwnerIds).Scan(&projects).Error
			if err != nil {
				return nil, errs.FailedToFind("project members", err)
			}
			projectIds := make([]string, 0, len(projects))
			if len(projects) != 0 {
				for _, v := range projects {
					projectIds = append(projectIds, v.ProjectId)
				}
				db = db.Where("project_id IN (?)", projectIds)
			}
		}
		if len(req.Filter.Statuses) != 0 {
			db = db.Where("status IN (?)", req.Filter.Statuses)
		}
		if req.Filter.GetCreatedFromTimestamp() < req.Filter.GetCreatedUntilTimestamp() {
			startDate := time.Unix(req.Filter.GetCreatedFromTimestamp(), 0)
			endDate := time.Unix(req.Filter.GetCreatedUntilTimestamp(), 0)
			db = db.Where("created_at BETWEEN ? AND ?", startDate, endDate)
		}
	}

	var collectionCount int64

	// Page token
	if pageToken == "" {
		err = db.Count(&collectionCount).Error
		if err != nil {
			return nil, errs.SQLQueryFailed(err, "count")
		}
	}

	dbs := make([]*Project, 0, pageSize+1)
	err = db.Find(&dbs).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "LIST")
	}

	pbs := make([]*project.Project, 0, len(dbs))
	for i, db := range dbs {
		if i == int(pageSize) {
			break
		}

		pb, err := ProjectProto(db)
		if err != nil {
			return nil, err
		}

		pbs = append(pbs, pb)

		ID = db.ID
	}

	var token string
	if len(dbs) > int(pageSize) {
		// Next page token
		token = base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(ID)))
	}

	return &project.ListProjectsResponse{
		NextPageToken:   token,
		Projects:        pbs,
		CollectionCount: int32(collectionCount),
	}, nil
}
