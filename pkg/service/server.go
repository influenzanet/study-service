package service

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"google.golang.org/grpc"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

type studyServiceServer struct {
	studyDBservice    *studydb.StudyDBService
	globalDBService   *globaldb.GlobalDBService
	StudyGlobalSecret string
}

// NewUserManagementServer creates a new service instance
func NewStudyServiceServer(
	studyDBservice *studydb.StudyDBService,
	globalDBservice *globaldb.GlobalDBService,
	studyGlobalSectret string,
) api.StudyServiceApiServer {
	return &studyServiceServer{
		studyDBservice:    studyDBservice,
		globalDBService:   globalDBservice,
		StudyGlobalSecret: studyGlobalSectret,
	}
}

// RunServer runs gRPC service to publish ToDo service
func RunServer(ctx context.Context, port string,
	studyDBservice *studydb.StudyDBService,
	globalDBservice *globaldb.GlobalDBService,
	globalStudySecret string,
) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// register service
	server := grpc.NewServer()
	api.RegisterStudyServiceApiServer(server, NewStudyServiceServer(
		studyDBservice,
		globalDBservice,
		globalStudySecret,
	))

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Println("starting gRPC server...")
	log.Println("wait connections on port " + port)
	return server.Serve(lis)
}
