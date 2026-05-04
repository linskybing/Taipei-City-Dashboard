package services

import (
	"TaipeiCityDashboardBE/global"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func recreateCollection(ctx context.Context, collectionName string, vectorSize uint64) error {
	qdrantConfig := global.Qdrant
	qdrantURL := qdrantConfig.Url

	log.Printf("Attempting to delete Qdrant collection '%s'...", collectionName)
	deleteURL := fmt.Sprintf("%s/collections/%s", qdrantURL, collectionName)
	deleteReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete collection request: %w", err)
	}
	if qdrantConfig.ApiKey != "" {
		deleteReq.Header.Set("api-key", qdrantConfig.ApiKey)
	}

	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to send delete collection request for '%s': %w", collectionName, err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode == http.StatusOK {
		log.Printf("Collection '%s' deleted successfully.", collectionName)
	} else {
		bodyBytes, _ := io.ReadAll(deleteResp.Body)
		log.Printf("Info: Qdrant delete collection returned status %s, body: %s", deleteResp.Status, string(bodyBytes))
	}

	log.Printf("Creating new Qdrant collection '%s' with vector size %d.", collectionName, vectorSize)
	createBody := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     vectorSize,
			"distance": "Cosine",
		},
	}
	createBodyBytes, err := json.Marshal(createBody)
	if err != nil {
		return fmt.Errorf("failed to marshal create collection request body: %w", err)
	}

	createURL := fmt.Sprintf("%s/collections/%s?timeout=30", qdrantURL, collectionName)
	createReq, err := http.NewRequestWithContext(ctx, http.MethodPut, createURL, bytes.NewBuffer(createBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create create-collection request: %w", err)
	}
	createReq.Header.Set("Content-Type", "application/json")
	if qdrantConfig.ApiKey != "" {
		createReq.Header.Set("api-key", qdrantConfig.ApiKey)
	}

	createResp, err := http.DefaultClient.Do(createReq)
	if err != nil {
		return fmt.Errorf("failed to send create-collection request: %w", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("qdrant create collection returned status %s, body: %s", createResp.Status, string(bodyBytes))
	}

	log.Printf("Collection '%s' created successfully.", collectionName)
	return nil
}

func upsertPoints(ctx context.Context, collectionName string, points []qdrantPoint) error {
	if len(points) == 0 {
		log.Println("No points to upsert. Skipping.")
		return nil
	}

	type upsertRequest struct {
		Points []qdrantPoint `json:"points"`
	}
	bodyBytes, err := json.Marshal(upsertRequest{Points: points})
	if err != nil {
		return fmt.Errorf("failed to marshal upsert request body: %w", err)
	}

	qdrantConfig := global.Qdrant
	upsertURL := fmt.Sprintf("%s/collections/%s/points?wait=true", qdrantConfig.Url, collectionName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upsertURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create upsert points request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if qdrantConfig.ApiKey != "" {
		req.Header.Set("api-key", qdrantConfig.ApiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending upsert points request for '%s': %v", collectionName, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant upsert points returned status %s, body: %s", resp.Status, string(bodyBytes))
	}

	log.Printf("Successfully upserted %d points to collection '%s'.", len(points), collectionName)
	return nil
}
