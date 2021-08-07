package types

import (
	"errors"

	"github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	FILE_STATUS_UPLOADING = "uploading"
	FILE_STATUS_FINISHED  = "finished"
)

type FileInfo struct {
	ID                   primitive.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID        string                `bson:"participantID,omitempty"`
	Status               string                `bson:"status,omitempty"`
	UploadedBy           string                `bson:"uploadedBy,omitempty"` // if not uploaded by the participant
	Path                 string                `bson:"path,omitempty"`
	PreviewPath          string                `bson:"previewPath,omitempty"`
	SubStudy             string                `bson:"subStudy,omitempty"`
	SubmittedAt          int64                 `bson:"submittedAt,omitempty"`
	FileType             string                `bson:"fileType,omitempty"`
	VisibleToParticipant bool                  `bson:"visibleToParticipant,omitempty"`
	Name                 string                `bson:"name,omitempty"`
	Size                 int32                 `bson:"size,omitempty"`
	ReferencedIn         []FileObjectReference `bson:"referencedIn,omitempty"`
}

type FileObjectReference struct {
	ID   string `bson:"id,omitempty"`
	Type string `bson:"type,omitempty"`
	Time int64  `bson:"time,omitempty"`
}

func (f *FileInfo) AddReference(ref FileObjectReference) error {
	f.ReferencedIn = append(f.ReferencedIn, ref)
	return nil
}

func (f FileObjectReference) ToAPI() *api.FileObjectReference {
	return &api.FileObjectReference{
		Id:   f.ID,
		Type: f.Type,
		Time: f.Time,
	}
}

func (f *FileInfo) RemoveReference(refID string) error {
	for i, cf := range f.ReferencedIn {
		if cf.ID == refID {
			f.ReferencedIn = append(f.ReferencedIn[:i], f.ReferencedIn[i+1:]...)
			return nil
		}
	}
	return errors.New("role not found")
}

func (o FileInfo) ToAPI() *api.FileInfo {
	refs := make([]*api.FileObjectReference, len(o.ReferencedIn))
	for i, r := range o.ReferencedIn {
		refs[i] = r.ToAPI()
	}

	return &api.FileInfo{
		Id:                   o.ID.Hex(),
		ParticipantId:        o.ParticipantID,
		Status:               o.Status,
		UploadedBy:           o.UploadedBy,
		Path:                 o.Path,
		PreviewPath:          o.PreviewPath,
		SubStudy:             o.SubStudy,
		SubmittedAt:          o.SubmittedAt,
		FileType:             o.FileType,
		VisibleToParticipant: o.VisibleToParticipant,
		Name:                 o.Name,
		Size:                 o.Size,
		ReferencedIn:         refs,
	}
}

func FileObjectReferenceFromAPI(o *api.FileObjectReference) FileObjectReference {
	if o == nil {
		return FileObjectReference{}
	}
	return FileObjectReference{
		ID:   o.Id,
		Type: o.Type,
		Time: o.Time,
	}
}

func FileInfoFromAPI(o *api.FileInfo) FileInfo {
	if o == nil {
		return FileInfo{}
	}
	refs := make([]FileObjectReference, len(o.ReferencedIn))
	for i, r := range o.ReferencedIn {
		refs[i] = FileObjectReferenceFromAPI(r)
	}
	_id, _ := primitive.ObjectIDFromHex(o.Id)
	return FileInfo{
		ID:                   _id,
		ParticipantID:        o.ParticipantId,
		Status:               o.Status,
		UploadedBy:           o.UploadedBy,
		Path:                 o.Path,
		PreviewPath:          o.PreviewPath,
		SubStudy:             o.SubStudy,
		SubmittedAt:          o.SubmittedAt,
		FileType:             o.FileType,
		VisibleToParticipant: o.VisibleToParticipant,
		Name:                 o.Name,
		Size:                 o.Size,
		ReferencedIn:         refs,
	}
}
