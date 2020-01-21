package main

import (
	"log"
	"net"

	"github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type studyServiceServer struct{}

var conf Config
var dbClient *mongo.Client

func init() {
	initConfig()
	dbInit()
}

func main() {
	lis, err := net.Listen("tcp", ":"+conf.Port)
	log.Println("wait connections on port " + conf.Port)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterStudyServiceApiServer(grpcServer, &studyServiceServer{})

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
