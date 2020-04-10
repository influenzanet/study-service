// Study service is responsible for handling study related logic, such as
// handling participants, receiving survey responses etc.
package main

import (
	"log"
	"math/rand"
	"net"
	"time"

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

func startTimerThread(timeCheckIntervalMin int, timeCheckIntervalRange int) {
	// TODO: turn of gracefully
	for {
		delay := rand.Intn(timeCheckIntervalRange) + timeCheckIntervalMin
		<-time.After(time.Duration(delay) * time.Second)
		go StudyTimerEvent()
	}
}

func main() {
	go startTimerThread(conf.Study.TimerEventCheckIntervalMin, conf.Study.TimerEventCheckIntervalVar)

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
