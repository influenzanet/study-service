package studytimer

import (
	"math/rand"
	"time"

	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

type StudyTimerService struct {
	globalDBService            *globaldb.GlobalDBService
	studyDBService             *studydb.StudyDBService
	TimerEventFrequency        int64 // how often the timer event should be performed (only from one instance of the service) - seconds
	TimerEventCheckIntervalMin int   // approx. how often this serice should check if to perform the timer event - seconds
	TimerEventCheckIntervalVar int   // range of the uniform random distribution - varying the check interval to avoid a steady collisions
}

func NewStudyTimerService(config types.StudyConfig, studyDBServ *studydb.StudyDBService, globalDBServ *globaldb.GlobalDBService) *StudyTimerService {
	return &StudyTimerService{
		globalDBService:            globalDBServ,
		studyDBService:             studyDBServ,
		TimerEventFrequency:        config.TimerEventFrequency,
		TimerEventCheckIntervalMin: config.TimerEventCheckIntervalMin,
		TimerEventCheckIntervalVar: config.TimerEventCheckIntervalVar,
	}
}

func (s *StudyTimerService) Run() {
	go s.startTimerThread(s.TimerEventCheckIntervalMin, s.TimerEventCheckIntervalVar)
}

func (s *StudyTimerService) startTimerThread(timeCheckIntervalMin int, timeCheckIntervalRange int) {
	// TODO: turn of gracefully
	for {
		delay := rand.Intn(timeCheckIntervalRange) + timeCheckIntervalMin
		<-time.After(time.Duration(delay) * time.Second)
		go s.StudyTimerEvent()
	}
}
