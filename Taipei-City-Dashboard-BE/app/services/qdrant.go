package services

import (
	"TaipeiCityDashboardBE/app/models"
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strings" // Added for string manipulation
	"sync/atomic"
)

// isQdrantRebuilding is an atomic boolean to prevent concurrent rebuilds.
var isQdrantRebuilding atomic.Bool

// qdrantPoint represents a single point to be upserted to Qdrant.
type qdrantPoint struct {
	// The ID is an interface{} to accommodate both integer and UUID string IDs.
	ID      interface{}            `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// RebuildQdrantPublicCollection fetches the latest public component data, generates vectors, and rebuilds the Qdrant collection.
// It is designed to be run asynchronously and is concurrency-safe.
func RebuildQdrantPublicCollection() ([]models.QuertChartAndConponentForQdrant, error) {
	// Use CompareAndSwap to ensure only one instance runs at a time.
	if !isQdrantRebuilding.CompareAndSwap(false, true) {
		log.Println("Qdrant rebuild is already in progress. Skipping.")
		return nil, fmt.Errorf("qdrant rebuild is already in progress")
	}
	// Ensure the flag is reset when the function exits.
	defer isQdrantRebuilding.Store(false)

	log.Println("Starting Qdrant public collection rebuild...")
	ctx := context.Background()

	// 1. Fetch data from PostgreSQL using the model function
	data, err := fetchPublicComponentData()
	if err != nil {
		log.Printf("Error fetching public component data: %v", err)
		return nil, err
	}
	if len(data) == 0 {
		log.Println("No public component data found. Aborting Qdrant rebuild.")
		return data, nil
	}

	// 2. Generate vectors for each data point	points, vectorSize, err := generateVectors(data)
	points, vectorSize, err := generateVectors(data)
	if err != nil {
		log.Printf("Error generating vectors: %v", err)
		return data, err
	}

	// 3. Recreate Qdrant collection
	collectionName := os.Getenv("QDRANT_COLLECTION_NAME")
	if collectionName == "" {
		collectionName = "query_charts" // Default collection name
	}

	err = recreateCollection(ctx, collectionName, uint64(vectorSize))
	if err != nil {
		log.Printf("Error recreating Qdrant collection: %v", err)
		return data, err
	}
	// 4. Upsert new points to Qdrant
	err = upsertPoints(ctx, collectionName, points)
	if err != nil {
		log.Printf("Error upserting points to Qdrant: %v", err)
		return data, err
	}

	log.Printf("Successfully rebuilt Qdrant collection '%s' with %d points.", collectionName, len(points))
	return data, nil
}

// fetchPublicComponentData now calls the model function.
func fetchPublicComponentData() ([]models.QuertChartAndConponentForQdrant, error) {
	return models.GetPublicComponentsForQdrant()
}

func generateVectors(data []models.QuertChartAndConponentForQdrant) ([]qdrantPoint, int, error) {
	var points []qdrantPoint
	var vectorSize int

	for _, item := range data {
		// Combine text fields for vector generation
		combinedText := item.LongDesc
		if item.UseCase != "" {
			if combinedText != "" {
				combinedText += " "
			}
			combinedText += item.UseCase
		}

		// Sanitize newline characters, as suspected by the user.
		// Replace all `\r\n`, `\r`, and `\n` with a single space.
		combinedText = strings.ReplaceAll(combinedText, "\r\n", " ")
		combinedText = strings.ReplaceAll(combinedText, "\r", " ")
		combinedText = strings.ReplaceAll(combinedText, "\n", " ")

		if combinedText == "" {
			log.Printf("Skipping item ID %d (%s) due to empty combined text for vector generation.", item.ID, item.Name)
			continue
		}

		// Generate vector
		vector, err := models.GenVector(combinedText)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to generate vector for item %s/%s: %w", item.Index, item.Name, err)
		}

		// Set vector size if not already set
		if vectorSize == 0 {
			vectorSize = len(vector)
		}

		// Create payload
		payload := map[string]interface{}{
			"id":        item.ID,
			"index":     item.Index,
			"name":      item.Name,
			"city":      item.City,
			"long_desc": item.LongDesc,
			"use_case":  item.UseCase,
		}

		points = append(points, qdrantPoint{
			ID:      qdrantComponentPointID(item),
			Vector:  vector,
			Payload: payload,
		})
	}

	return points, vectorSize, nil
}

func qdrantComponentPointID(item models.QuertChartAndConponentForQdrant) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(fmt.Sprintf("%d:%s:%s", item.ID, item.City, item.Index)))
	return hash.Sum64()
}
