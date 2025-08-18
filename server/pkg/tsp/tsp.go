// Package tsp provides heuristics for solving the Traveling Salesperson Problem.
package tsp

import (
	"math"
)

// Point representa uma localização com coordenadas geográficas que pode ser usada pelo otimizador.
type Point interface {
	GetCoordinates() (latitude, longitude float64)
}

// haversineDistance calcula a distância em quilômetros entre dois pontos na Terra.
func haversineDistance(p1, p2 Point) float64 {
	const R = 6371 // Raio da Terra em quilômetros
	lat1, lon1 := p1.GetCoordinates()
	lat2, lon2 := p2.GetCoordinates()

	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// OptimizeRouteNearestNeighbor encontra um caminho curto usando a heurística do Vizinho Mais Próximo.
// Ele começa no primeiro ponto e repetidamente visita o ponto não visitado mais próximo.
func OptimizeRouteNearestNeighbor(points []Point) []Point {
	if len(points) <= 1 {
		return points
	}

	unvisited := make(map[int]Point)
	for i, p := range points {
		unvisited[i] = p
	}

	// Começa no primeiro ponto da lista (índice 0)
	currentIdx := 0
	currentPoint := unvisited[currentIdx]
	delete(unvisited, currentIdx)

	orderedRoute := []Point{currentPoint}

	// Loop até que todos os pontos tenham sido visitados
	for len(unvisited) > 0 {
		nearestIdx := -1
		minDist := math.MaxFloat64

		// Encontra o ponto não visitado mais próximo do ponto atual
		for idx, p := range unvisited {
			dist := haversineDistance(currentPoint, p)
			if dist < minDist {
				minDist = dist
				nearestIdx = idx
			}
		}

		// Move para o ponto mais próximo encontrado
		currentIdx = nearestIdx
		currentPoint = unvisited[currentIdx]
		delete(unvisited, currentIdx)
		orderedRoute = append(orderedRoute, currentPoint)
	}

	return orderedRoute
}
