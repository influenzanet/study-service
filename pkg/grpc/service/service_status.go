package service

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/influenzanet/study-service/pkg/api"
)

// Status endpoint should return internal status of the system if running correctly
func (s *studyServiceServer) Status(ctx context.Context, _ *empty.Empty) (*api.ServiceStatus, error) {
	return &api.ServiceStatus{
		Status:  api.ServiceStatus_NORMAL,
		Msg:     "service running",
		Version: apiVersion,
	}, nil
}
