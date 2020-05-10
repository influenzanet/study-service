package globaldb

import (
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (dbService *GlobalDBService) GetAllInstances() ([]types.Instance, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	cur, err := dbService.collectionRefInstances().Find(
		ctx,
		filter,
	)

	if err != nil {
		return []types.Instance{}, err
	}
	defer cur.Close(ctx)

	instances := []types.Instance{}
	for cur.Next(ctx) {
		var result types.Instance
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
