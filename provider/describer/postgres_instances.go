package describer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opengovern/og-describer-template/pkg/sdk/models"
	"github.com/opengovern/og-describer-template/provider/model"
	"net/http"
	"net/url"
	"sync"
)

func ListPostgresInstances(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
	var wg sync.WaitGroup
	renderChan := make(chan models.Resource)
	errorChan := make(chan error, 1) // Buffered channel to capture errors

	go func() {
		defer close(renderChan)
		defer close(errorChan)
		if err := processPostgresInstances(ctx, handler, renderChan, &wg); err != nil {
			errorChan <- err // Send error to the error channel
		}
		wg.Wait()
	}()

	var values []models.Resource
	for {
		select {
		case value, ok := <-renderChan:
			if !ok {
				return values, nil
			}
			if stream != nil {
				if err := (*stream)(value); err != nil {
					return nil, err
				}
			} else {
				values = append(values, value)
			}
		case err := <-errorChan:
			return nil, err
		}
	}
}

func GetPostgresInstance(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*models.Resource, error) {
	postgres, err := processPostgresInstance(ctx, handler, resourceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   postgres.ID,
		Name: postgres.Name,
		Description: JSONAllFieldsMarshaller{
			Value: postgres,
		},
	}
	return &value, nil
}

func processPostgresInstances(ctx context.Context, handler *RenderAPIHandler, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var postgresInstances []model.PostgresDescription
	var postgresListResponse []model.PostgresResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/postgres"
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", limit)
		params.Set("includeReplicas", includeReplicas)
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		finalURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		req, err := http.NewRequest("GET", finalURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		requestFunc := func(req *http.Request) (*http.Response, error) {
			var e error
			resp, e = handler.Client.Do(req)
			if e != nil {
				return nil, fmt.Errorf("request execution failed: %w", e)
			}
			defer resp.Body.Close()

			if e = json.NewDecoder(resp.Body).Decode(&postgresListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, postgresResp := range postgresListResponse {
				postgresInstances = append(postgresInstances, postgresResp.Postgres)
				if i == len(postgresListResponse)-1 {
					cursor = postgresResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(postgresListResponse) < 100 {
			break
		}
	}
	for _, postgres := range postgresInstances {
		wg.Add(1)
		go func(postgres model.PostgresDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   postgres.ID,
				Name: postgres.Name,
				Description: JSONAllFieldsMarshaller{
					Value: postgres,
				},
			}
			renderChan <- value
		}(postgres)
	}
	return nil
}

func processPostgresInstance(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*model.PostgresDescription, error) {
	var postgres model.PostgresDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/postgres/"

	finalURL := fmt.Sprintf("%s%s", baseURL, resourceID)
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	requestFunc := func(req *http.Request) (*http.Response, error) {
		var e error
		resp, e = handler.Client.Do(req)
		if e != nil {
			return nil, fmt.Errorf("request execution failed: %w", e)
		}
		defer resp.Body.Close()

		if e = json.NewDecoder(resp.Body).Decode(&postgres); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &postgres, nil
}
