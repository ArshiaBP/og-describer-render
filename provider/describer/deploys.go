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

func ListDeploys(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
	var wg sync.WaitGroup
	renderChan := make(chan models.Resource)
	errorChan := make(chan error, 1) // Buffered channel to capture errors
	services, err := getServices(ctx, handler)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(renderChan)
		defer close(errorChan)
		for _, service := range services {
			if err := processDeploys(ctx, handler, service.ID, renderChan, &wg); err != nil {
				errorChan <- err // Send error to the error channel
			}
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

func GetDeploy(ctx context.Context, handler *RenderAPIHandler, resourceID, serviceID string) (*models.Resource, error) {
	deploy, err := processDeploy(ctx, handler, resourceID, serviceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   deploy.ID,
		Name: deploy.Status,
		Description: JSONAllFieldsMarshaller{
			Value: deploy,
		},
	}
	return &value, nil
}

func processDeploys(ctx context.Context, handler *RenderAPIHandler, serviceID string, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var deploys []model.DeployDescription
	var deployListResponse []model.DeployResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", limit)
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		finalURL := fmt.Sprintf("%s%sdeploys?%s", baseURL, serviceID, params.Encode())

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

			if e = json.NewDecoder(resp.Body).Decode(&deployListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, deployResp := range deployListResponse {
				deploys = append(deploys, deployResp.Deploy)
				if i == len(deployListResponse)-1 {
					cursor = deployResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(deployListResponse) < 100 {
			break
		}
	}
	for _, deploy := range deploys {
		wg.Add(1)
		go func(deploy model.DeployDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   deploy.ID,
				Name: deploy.Status,
				Description: JSONAllFieldsMarshaller{
					Value: deploy,
				},
			}
			renderChan <- value
		}(deploy)
	}
	return nil
}

func processDeploy(ctx context.Context, handler *RenderAPIHandler, resourceID, serviceID string) (*model.DeployDescription, error) {
	var deploy model.DeployDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"

	finalURL := fmt.Sprintf("%s%s/deploys/%s", baseURL, serviceID, resourceID)
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

		if e = json.NewDecoder(resp.Body).Decode(&deploy); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &deploy, nil
}
