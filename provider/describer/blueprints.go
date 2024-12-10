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

func ListBluePrints(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
	var wg sync.WaitGroup
	renderChan := make(chan models.Resource)
	errorChan := make(chan error, 1) // Buffered channel to capture errors

	go func() {
		defer close(renderChan)
		defer close(errorChan)
		if err := processBluePrints(ctx, handler, renderChan, &wg); err != nil {
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

func GetBluePrint(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*models.Resource, error) {
	blueprint, err := processBluePrint(ctx, handler, resourceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   blueprint.ID,
		Name: blueprint.Name,
		Description: JSONAllFieldsMarshaller{
			Value: blueprint,
		},
	}
	return &value, nil
}

func processBluePrints(ctx context.Context, handler *RenderAPIHandler, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var blueprints []model.BlueprintDescription
	var blueprintListResponse []model.BlueprintResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/blueprints"
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", limit)
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

			if e = json.NewDecoder(resp.Body).Decode(&blueprintListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, blueprintResp := range blueprintListResponse {
				blueprints = append(blueprints, blueprintResp.BluePrint)
				if i == len(blueprintListResponse)-1 {
					cursor = blueprintResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(blueprintListResponse) < 100 {
			break
		}
	}
	for _, blueprint := range blueprints {
		wg.Add(1)
		go func(blueprint model.BlueprintDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   blueprint.ID,
				Name: blueprint.Name,
				Description: JSONAllFieldsMarshaller{
					Value: blueprint,
				},
			}
			renderChan <- value
		}(blueprint)
	}
	return nil
}

func processBluePrint(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*model.ProjectDescription, error) {
	var project model.ProjectDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/blueprints/"

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

		if e = json.NewDecoder(resp.Body).Decode(&project); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &project, nil
}
