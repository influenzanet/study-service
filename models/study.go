package models

import (
	"github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Study struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string             `bson:"key"`
	SecretKey string             `bson:"secretKey"`
	Status    string             `bson:"status"`
	Members   []StudyMember      `bson:"members"` // users with access to manage study
	Rules     []Expression       `bson:"rules"`   // defining how the study should run
}

type StudyMember struct {
	UserID   string `bson:"userID"`
	UserName string `bson:"username"`
	Role     string `bson:"role"`
}

type StudyProps struct {
	Name        []LocalisedObject `bson:"name"`
	Description []LocalisedObject `bson:"description"`
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
		rules[i] = ExpressionFromAPI(r)
	}
	_id, _ := primitive.ObjectIDFromHex(s.Id)
	return Study{
		ID:        _id,
		Key:       s.Key,
		SecretKey: s.SecretKey,
		Status:    s.Status,
		Members:   members,
		Rules:     rules,
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
	return &api.Study_Props{
		Name:        name,
		Description: description,
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
	return StudyProps{
		Name:        name,
		Description: description,
	}
}
