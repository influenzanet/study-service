package types

import (
	"github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyMessage struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Type          string             `bson:"type,omitempty"`
	Payload       map[string]string  `bson:"payload,omitempty"`
	ParticipantID string             `bson:"participantID,omitempty"`
}

func (m StudyMessage) ToAPI() *api.StudyMessage {
	return &api.StudyMessage{
		Id:            m.ID.Hex(),
		Type:          m.Type,
		Payload:       m.Payload,
		ParticipantId: m.ParticipantID,
	}
}
