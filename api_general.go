package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/influenzanet/study-service/api"
)

// Status endpoint should return internal status of the system if running correctly
func (s *studyServiceServer) Status(ctx context.Context, _ *empty.Empty) (*api.Status, error) {
	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "service running",
	}, nil
}
