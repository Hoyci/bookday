package routing

import "github.com/hoyci/bookday/pkg/tsp"

// GetCoordinates faz com que a nossa struct deliveryPoint satisfaça a interface tsp.Point.
func (dp deliveryPoint) GetCoordinates() (lat, lon float64) {
	return dp.Latitude, dp.Longitude
}

// optimizeRoute pega uma fatia de pontos de entrega (um cluster) e retorna
// a mesma fatia, mas ordenada de forma otimizada para a entrega.
func optimizeRoute(points []deliveryPoint) []deliveryPoint {
	if len(points) <= 1 {
		return points
	}

	// 1. Converte a nossa fatia de tipo específico para a interface genérica tsp.Point.
	tspPoints := make([]tsp.Point, len(points))
	for i, p := range points {
		tspPoints[i] = p
	}

	// 2. Chama o otimizador genérico do pacote TSP.
	optimizedPoints := tsp.OptimizeRouteNearestNeighbor(tspPoints)

	// 3. Converte a fatia otimizada de volta para o nosso tipo específico.
	optimizedDeliveryPoints := make([]deliveryPoint, len(optimizedPoints))
	for i, p := range optimizedPoints {
		optimizedDeliveryPoints[i] = p.(deliveryPoint)
	}

	return optimizedDeliveryPoints
}
