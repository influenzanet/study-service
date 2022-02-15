package service

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
	"google.golang.org/grpc"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

type studyServiceServer struct {
	api.UnimplementedStudyServiceApiServer
	clients           *types.APIClients
	studyDBservice    *studydb.StudyDBService
	globalDBService   *globaldb.GlobalDBService
	StudyGlobalSecret string
}

// NewUserManagementServer creates a new service instance
func NewStudyServiceServer(
	clients *types.APIClients,
	studyDBservice *studydb.StudyDBService,
	globalDBservice *globaldb.GlobalDBService,
	studyGlobalSectret string,
) api.StudyServiceApiServer {

	return &studyServiceServer{
		clients:           clients,
		studyDBservice:    studyDBservice,
		globalDBService:   globalDBservice,
		StudyGlobalSecret: studyGlobalSectret,
	}
}

// RunServer runs gRPC service to publish ToDo service
func RunServer(ctx context.Context, port string,
	clients *types.APIClients,
	studyDBservice *studydb.StudyDBService,
	globalDBservice *globaldb.GlobalDBService,
	globalStudySecret string,
	maxMsgSize int,
) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// register service
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)
	api.RegisterStudyServiceApiServer(server, NewStudyServiceServer(
		clients,
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
			logger.Info.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// start gRPC server
	logger.Info.Println("starting gRPC server...")
	logger.Info.Println("wait connections on port " + port)
	return server.Serve(lis)
}
