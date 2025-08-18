package routing

import (
	"fmt"
	"math"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

const maxStopsPerRoute = 20

func clusterStops(points []deliveryPoint) ([][]deliveryPoint, error) {
	if len(points) == 0 {
		return nil, nil
	}

	var observations clusters.Observations
	coordToPointMap := make(map[string]deliveryPoint)

	for _, point := range points {
		coords := clusters.Coordinates{point.Longitude, point.Latitude}
		observations = append(observations, coords)
		coordKey := fmt.Sprintf("%.6f,%.6f", point.Longitude, point.Latitude)
		coordToPointMap[coordKey] = point
	}

	k := int(math.Ceil(float64(len(points)) / float64(maxStopsPerRoute)))
	if k == 0 && len(points) > 0 {
		k = 1
	}

	km := kmeans.New()
	clusterResult, err := km.Partition(observations, k)
	if err != nil {
		return nil, err
	}

	resultClusters := make([][]deliveryPoint, len(clusterResult))
	for i, c := range clusterResult {
		var pointCluster []deliveryPoint
		for _, obs := range c.Observations {
			coordKey := fmt.Sprintf("%.6f,%.6f", obs.Coordinates()[0], obs.Coordinates()[1])
			if originalPoint, ok := coordToPointMap[coordKey]; ok {
				pointCluster = append(pointCluster, originalPoint)
				delete(coordToPointMap, coordKey)
			}
		}
		resultClusters[i] = pointCluster
	}

	return resultClusters, nil
}
