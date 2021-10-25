package types

import (
	"github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	STUDY_ROLE_OWNER      = "owner"
	STUDY_ROLE_MAINTAINER = "maintainer"
)

const (
	STUDY_STATUS_ACTIVE   = "active"
	STUDY_STATUS_INACTIVE = "inactive"
)

type Study struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key            string             `bson:"key"`
	SecretKey      string             `bson:"secretKey"`
	Status         string             `bson:"status"`
	Members        []StudyMember      `bson:"members"` // users with access to manage study
	Rules          []Expression       `bson:"rules"`   // defining how the study should run
	Props          StudyProps         `bson:"props"`
	NextTimerEvent int64              `bson:"nextTimerEventAfter"`
	Stats          StudyStats         `bson:"studyStats"`
	Configs        StudyConfigs       `bson:"configs"`
}

type StudyMember struct {
	UserID   string `bson:"userID"`
	UserName string `bson:"username"`
	Role     string `bson:"role"`
}

type StudyProps struct {
	Name               []LocalisedObject `bson:"name"`
	Description        []LocalisedObject `bson:"description"`
	Tags               []Tag             `bson:"tags"`
	StartDate          int64             `bson:"startDate"`
	EndDate            int64             `bson:"endDate"`
	SystemDefaultStudy bool              `bson:"systemDefaultStudy"`
}

type StudyConfigs struct {
	ParticipantFileUploadRule *Expression `bson:"participantFileUploadRule"`
	IdMappingMethod           string      `bson:"idMappingMethod"`
}

type StudyStats struct {
	ParticipantCount int64 `bson:"participantCount"`
	ResponseCount    int64 `bson:"responseCount"`
}

type Tag struct {
	Label []LocalisedObject `bson:"label"`
}

func (s Study) ToAPI() *api.Study {
	members := make([]*api.Study_Member, len(s.Members))
	for i, r := range s.Members {
		members[i] = r.ToAPI()
	}
	rules := make([]*api.Expression, len(s.Rules))
	for i, r := range s.Rules {
		rules[i] = r.ToAPI()
	}

	return &api.Study{
		Id:        s.ID.Hex(),
		Key:       s.Key,
		SecretKey: s.SecretKey,
		Status:    s.Status,
		Members:   members,
		Rules:     rules,
		Props:     s.Props.ToAPI(),
		Stats:     s.Stats.ToAPI(),
		Congfigs:  s.Configs.ToAPI(),
	}
}

func (s StudyConfigs) ToAPI() *api.Study_Configs {
	return &api.Study_Configs{
		ParticipantFileUploadRule: s.ParticipantFileUploadRule.ToAPI(),
		IdMappingMethod:           s.IdMappingMethod,
	}
}

func StudyConfigsFromAPI(s *api.Study_Configs) StudyConfigs {
	if s == nil {
		return StudyConfigs{}
	}
	return StudyConfigs{
		ParticipantFileUploadRule: ExpressionFromAPI(s.ParticipantFileUploadRule),
		IdMappingMethod:           s.IdMappingMethod,
	}
}

func StudyFromAPI(s *api.Study) Study {
	if s == nil {
		return Study{}
	}
	members := make([]StudyMember, len(s.Members))
	for i, r := range s.Members {
		members[i] = StudyMemberFromAPI(r)
	}
	rules := make([]Expression, len(s.Rules))
	for i, r := range s.Rules {
		rules[i] = *ExpressionFromAPI(r)
	}
	_id, _ := primitive.ObjectIDFromHex(s.Id)
	return Study{
		ID:        _id,
		Key:       s.Key,
		SecretKey: s.SecretKey,
		Status:    s.Status,
		Members:   members,
		Rules:     rules,
		Props:     StudyPropsFromAPI(s.Props),
		Stats:     StudyStatsFromAPI(s.Stats),
		Configs:   StudyConfigsFromAPI(s.Congfigs),
	}
}

func (sm StudyMember) ToAPI() *api.Study_Member {
	return &api.Study_Member{
		UserId:   sm.UserID,
		Role:     sm.Role,
		Username: sm.UserName,
	}
}

func StudyMemberFromAPI(sm *api.Study_Member) StudyMember {
	if sm == nil {
		return StudyMember{}
	}
	return StudyMember{
		UserID:   sm.UserId,
		Role:     sm.Role,
		UserName: sm.Username,
	}
}

func (sp StudyProps) ToAPI() *api.Study_Props {
	name := make([]*api.LocalisedObject, len(sp.Name))
	for i, r := range sp.Name {
		name[i] = r.ToAPI()
	}
	description := make([]*api.LocalisedObject, len(sp.Description))
	for i, r := range sp.Description {
		description[i] = r.ToAPI()
	}
	tags := make([]*api.Tag, len(sp.Tags))
	for i, r := range sp.Tags {
		tags[i] = r.ToAPI()
	}
	return &api.Study_Props{
		Name:               name,
		Description:        description,
		Tags:               tags,
		StartDate:          sp.StartDate,
		EndDate:            sp.EndDate,
		SystemDefaultStudy: sp.SystemDefaultStudy,
	}
}

func StudyPropsFromAPI(sp *api.Study_Props) StudyProps {
	if sp == nil {
		return StudyProps{}
	}
	name := make([]LocalisedObject, len(sp.Name))
	for i, r := range sp.Name {
		name[i] = LocalisedObjectFromAPI(r)
	}
	description := make([]LocalisedObject, len(sp.Description))
	for i, r := range sp.Description {
		description[i] = LocalisedObjectFromAPI(r)
	}
	tags := make([]Tag, len(sp.Tags))
	for i, r := range sp.Tags {
		tags[i] = TagFromAPI(r)
	}
	return StudyProps{
		Name:               name,
		Description:        description,
		Tags:               tags,
		StartDate:          sp.StartDate,
		EndDate:            sp.EndDate,
		SystemDefaultStudy: sp.SystemDefaultStudy,
	}
}

func TagFromAPI(t *api.Tag) Tag {
	if t == nil {
		return Tag{}
	}
	label := make([]LocalisedObject, len(t.Label))
	for i, r := range t.Label {
		label[i] = LocalisedObjectFromAPI(r)
	}
	return Tag{
		Label: label,
	}
}

func (t Tag) ToAPI() *api.Tag {
	label := make([]*api.LocalisedObject, len(t.Label))
	for i, r := range t.Label {
		label[i] = r.ToAPI()
	}
	return &api.Tag{
		Label: label,
	}
}

func StudyStatsFromAPI(t *api.Study_Stats) StudyStats {
	if t == nil {
		return StudyStats{}
	}
	return StudyStats{
		ParticipantCount: t.ParticipantCount,
		ResponseCount:    t.ResponseCount,
	}
}

func (t StudyStats) ToAPI() *api.Study_Stats {
	return &api.Study_Stats{
		ParticipantCount: t.ParticipantCount,
		ResponseCount:    t.ResponseCount,
	}
}
