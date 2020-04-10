package main

import (
	"log"

	"github.com/influenzanet/study-service/models"
	"github.com/influenzanet/study-service/studyengine"
	"go.mongodb.org/mongo-driver/bson"
)

func StudyTimerEvent() {
	instances, err := getAllInstances()
	if err != nil {
		log.Printf("unexpected error: %s", err.Error())
	}
	for _, instance := range instances {
		studies, err := getStudiesByStatus(instance.InstanceID, "active", true)
		if err != nil {
			log.Printf("unexpected error: %s", err.Error())
			return
		}
		for _, study := range studies {
			if err := shouldPerformTimerEvent(instance.InstanceID, study.Key, conf.Study.TimerEventFrequency); err != nil {
				continue
			}
			log.Printf("performing timer event for study: %s - %s", instance.InstanceID, study.Key)

			rules, err := getStudyRules(instance.InstanceID, study.Key)
			if err != nil {
				continue
			}

			studyEvent := models.StudyEvent{
				Type: "TIMER",
			}

			// Get all participants
			ctx, cancel := getContext()
			defer cancel()

			filter := bson.M{"studyStatus": "active"}
			cur, err := collectionRefStudyParticipant(instance.InstanceID, study.Key).Find(ctx, filter)
			if err != nil {
				continue
			}
			defer cur.Close(ctx)

			for cur.Next(ctx) {
				// Update state of every participant
				var pState models.ParticipantState
				err := cur.Decode(&pState)
				if err != nil {
					continue
				}

				for _, rule := range rules {
					pState, err = studyengine.ActionEval(rule, pState, studyEvent)
					if err != nil {
						continue
					}
				}
				// save state back to DB
				_, err = saveParticipantStateDB(instance.InstanceID, study.Key, pState)
				if err != nil {
					continue
				}
			}
			if err := cur.Err(); err != nil {
				continue
			}
		}
	}
}
