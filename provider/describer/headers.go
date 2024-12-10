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

func ListHeaders(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
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
			if err := processHeaders(ctx, handler, service.ID, renderChan, &wg); err != nil {
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

func processHeaders(ctx context.Context, handler *RenderAPIHandler, serviceID string, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var headers []model.HeaderDescription
	var headerListResponse []model.HeaderResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", limit)
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		finalURL := fmt.Sprintf("%s%sheaders?%s", baseURL, serviceID, params.Encode())

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

			if e = json.NewDecoder(resp.Body).Decode(&headerListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, headerResp := range headerListResponse {
				headers = append(headers, headerResp.Header)
				if i == len(headerListResponse)-1 {
					cursor = headerResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(headerListResponse) < 100 {
			break
		}
	}
	for _, header := range headers {
		wg.Add(1)
		go func(header model.HeaderDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   header.ID,
				Name: header.Name,
				Description: JSONAllFieldsMarshaller{
					Value: header,
				},
			}
			renderChan <- value
		}(header)
	}
	return nil
}
