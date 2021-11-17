package project

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	project "github.com/gidyon/services/pkg/api/project/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ValidateProjectMember(pb *project.ProjectMember) error {
	switch {
	case pb == nil:
		return errs.NilObject("project")
	case pb.MemberId == "":
		return errs.MissingField("member id")
	case pb.ProjectId == "":
		return errs.MissingField("project id")
	}
	return nil
}

func (projectAPI *projectAPIServer) CreateProjectMember(
	ctx context.Context, req *project.CreateProjectMemberRequest,
) (*project.ProjectMember, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.NilObject("request")
	default:
		err = ValidateProjectMember(req.ProjectMember)
		if err != nil {
			return nil, err
		}
	}

	db, err := ProjectMemberModel(req.ProjectMember)
	if err != nil {
		return nil, err
	}

	// Create in the database
	err = projectAPI.SqlDb.Create(db).Error
	if err == nil {
		return nil, errs.FailedToSave("project", err)
	}

	return ProjectMemberProto(db)
}

func (projectAPI *projectAPIServer) DeleteProjectMember(
	ctx context.Context, req *project.DeleteProjectMemberRequest,
) (*empty.Empty, error) {
	var err error

	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.MemberId == "":
		return nil, errs.MissingField("member id")
	}

	err = projectAPI.SqlDb.Delete("id=?", req.MemberId).Error
	if err != nil {
		return nil, errs.FailedToDelete("project", err)
	}

	return &emptypb.Empty{}, nil
}

func (projectAPI *projectAPIServer) ListProjectMembers(
	ctx context.Context, req *project.ListProjectMembersRequest,
) (*project.ListProjectMembersResponse, error) {
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
			req.Filter = &project.ListProjectsMemberFilter{}
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

	db := projectAPI.SqlDb.Unscoped().Model(&ProjectMember{}).Limit(int(pageSize + 1)).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	// Apply tv filters
	if req.Filter != nil {
		if len(req.Filter.OwnerIds) != 0 {
			db = db.Where("owner_id IN (?)", req.Filter.OwnerIds)
		}
		if len(req.Filter.OwnerIds) != 0 {
			db = db.Where("owner_id IN (?)", req.Filter.OwnerIds)
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

	dbs := make([]*ProjectMember, 0, pageSize+1)
	err = db.Find(&dbs).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "LIST")
	}

	pbs := make([]*project.ProjectMember, 0, len(dbs))
	for i, db := range dbs {
		if i == int(pageSize) {
			break
		}

		pb, err := ProjectMemberProto(db)
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

	return &project.ListProjectMembersResponse{
		NextPageToken:   token,
		ProjectMembers:  pbs,
		CollectionCount: int32(collectionCount),
	}, nil
}
