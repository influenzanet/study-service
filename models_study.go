package main

import "github.com/influenzanet/study-service/api"

// Study defines the structure how a study is stored into the DB
type Study struct {
}

func studyFromAPI(p *api.Study) Study {
	return Study{}
}

// ToAPI converts a study from DB format into the API format
func (p Study) ToAPI() *api.Study {
	return &api.Study{}
}
