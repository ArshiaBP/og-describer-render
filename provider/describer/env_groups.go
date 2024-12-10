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

func ListEnvGroups(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
	var wg sync.WaitGroup
	renderChan := make(chan models.Resource)
	errorChan := make(chan error, 1) // Buffered channel to capture errors

	go func() {
		defer close(renderChan)
		defer close(errorChan)
		if err := processServices(ctx, handler, renderChan, &wg); err != nil {
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

func GetEnvGroup(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*models.Resource, error) {
	service, err := processService(ctx, handler, resourceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   service.ID,
		Name: service.Name,
		Description: JSONAllFieldsMarshaller{
			Value: service,
		},
	}
	return &value, nil
}

func processEnvGroups(ctx context.Context, handler *RenderAPIHandler, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var envGroups []model.EnvGroupDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/env-groups"
	params := url.Values{}
	params.Set("limit", limit)
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

		if e = json.NewDecoder(resp.Body).Decode(&envGroups); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, nil
	}
	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return fmt.Errorf("error during request handling: %w", err)
	}
	for _, envGroup := range envGroups {
		wg.Add(1)
		go func(envGroup model.EnvGroupDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   envGroup.ID,
				Name: envGroup.Name,
				Description: JSONAllFieldsMarshaller{
					Value: envGroup,
				},
			}
			renderChan <- value
		}(envGroup)
	}
	return nil
}

func processEnvGroup(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*model.ServiceDescription, error) {
	var service model.ServiceDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"

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

		if e = json.NewDecoder(resp.Body).Decode(&service); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &service, nil
}
