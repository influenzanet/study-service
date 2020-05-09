package globaldb

import (
	"github.com/influenzanet/study-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) GetAllInstances() ([]models.Instance, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	cur, err := dbService.collectionRefInstances().Find(
		ctx,
		filter,
	)

	if err != nil {
		return []models.Instance{}, err
	}
	defer cur.Close(ctx)

	instances := []models.Instance{}
	for cur.Next(ctx) {
		var result models.Instance
		err := cur.Decode(&result)
		if err != nil {
			return instances, err
		}

		instances = append(instances, result)
	}
	if err := cur.Err(); err != nil {
		return instances, err
	}

	return instances, nil
}
