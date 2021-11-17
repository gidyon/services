package project

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	project "github.com/gidyon/services/pkg/api/project/v1"
	"gorm.io/gorm"
)

const defaultProjectTable = "projects"

var projectsTable = ""

type Project struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OwnerId     string    `gorm:"index;type:varchar(50);not null"`
	OwnerEmail  string    `gorm:"type:varchar(50)"`
	OwnerNames  string    `gorm:"type:varchar(50)"`
	ProjectName string    `gorm:"type:varchar(50);not null"`
	Description string    `gorm:"type:varchar(150);"`
	Status      string    `gorm:"index;type:varchar(50)"`
	Scopes      []byte    `gorm:"type:json"`
	CreatedAt   time.Time `gorm:"autoCreateTime;->;<-:create;not null;type:datetime(6)"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;<-;type:datetime(6)"`
	DeletedAt   gorm.DeletedAt
}

func (*Project) TableName() string {
	if projectsTable != "" {
		return projectsTable
	}
	return defaultProjectTable
}

func ProjectProto(db *Project) (*project.Project, error) {
	pb := &project.Project{
		ProjectId:   fmt.Sprint(db.ID),
		ProjectName: db.ProjectName,
		OwnerNames:  db.OwnerNames,
		OwnerEmail:  db.OwnerEmail,
		Description: db.Description,
		Status:      db.Status,
		CreateDate:  db.CreatedAt.UTC().Format(time.RFC3339),
	}
	if len(db.Scopes) != 0 {
		err := json.Unmarshal(db.Scopes, &pb.Scopes)
		if err != nil {
			return nil, errs.FromJSONUnMarshal(err, "project scopes")
		}
	}
	return pb, nil
}

func ProjectModel(pb *project.Project) (*Project, error) {
	db := &Project{
		OwnerId:     pb.OwnerId,
		OwnerEmail:  pb.OwnerEmail,
		OwnerNames:  pb.OwnerNames,
		ProjectName: pb.ProjectName,
		Description: pb.Description,
		Status:      pb.Status,
	}
	if len(pb.Scopes) != 0 {
		bs, err := json.Marshal(pb.Scopes)
		if err != nil {
			return nil, errs.FromJSONMarshal(err, "project scopes")
		}
		db.Scopes = bs
	}
	return db, nil
}

type ProjectMember struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserId    string    `gorm:"index;type:varchar(50);not null"`
	ProjectId string    `gorm:"index;type:varchar(50);not null"`
	Status    string    `gorm:"index;type:varchar(50)"`
	Scopes    []byte    `gorm:"type:json"`
	CreatedAt time.Time `gorm:"autoCreateTime;->;<-:create;not null;type:datetime(6)"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;<-;type:datetime(6)"`
	DeletedAt gorm.DeletedAt
}

func ProjectMemberModel(pb *project.ProjectMember) (*ProjectMember, error) {
	db := &ProjectMember{
		UserId:    pb.UserId,
		ProjectId: pb.ProjectId,
		Status:    pb.Status,
	}
	if len(pb.Scopes) != 0 {
		bs, err := json.Marshal(pb.Scopes)
		if err != nil {
			return nil, errs.FromJSONMarshal(err, "project scopes")
		}
		db.Scopes = bs
	}
	return db, nil
}

func ProjectMemberProto(db *ProjectMember) (*project.ProjectMember, error) {
	pb := &project.ProjectMember{
		MemberId:  fmt.Sprint(db.ID),
		UserId:    db.UserId,
		ProjectId: db.ProjectId,
		JoinDate:  db.CreatedAt.UTC().Format(time.RFC3339),
		Status:    db.Status,
	}
	if len(db.Scopes) != 0 {
		err := json.Unmarshal(db.Scopes, &pb.Scopes)
		if err != nil {
			return nil, errs.FromJSONUnMarshal(err, "project scopes")
		}
	}
	return pb, nil
}
