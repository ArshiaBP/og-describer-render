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

func ListDisks(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
	var wg sync.WaitGroup
	renderChan := make(chan models.Resource)
	errorChan := make(chan error, 1) // Buffered channel to capture errors

	go func() {
		defer close(renderChan)
		defer close(errorChan)
		if err := processDisks(ctx, handler, renderChan, &wg); err != nil {
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

func GetDisk(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*models.Resource, error) {
	disk, err := processDisk(ctx, handler, resourceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   disk.ID,
		Name: disk.Name,
		Description: JSONAllFieldsMarshaller{
			Value: disk,
		},
	}
	return &value, nil
}

func processDisks(ctx context.Context, handler *RenderAPIHandler, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var disks []model.DiskDescription
	var diskListResponse []model.DiskResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/disks"
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

			if e = json.NewDecoder(resp.Body).Decode(&diskListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, diskResp := range diskListResponse {
				disks = append(disks, diskResp.Disk)
				if i == len(diskListResponse)-1 {
					cursor = diskResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(diskListResponse) < 100 {
			break
		}
	}
	for _, disk := range disks {
		wg.Add(1)
		go func(disk model.DiskDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   disk.ID,
				Name: disk.Name,
				Description: JSONAllFieldsMarshaller{
					Value: disk,
				},
			}
			renderChan <- value
		}(disk)
	}
	return nil
}

func processDisk(ctx context.Context, handler *RenderAPIHandler, resourceID string) (*model.DiskDescription, error) {
	var disk model.DiskDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/disks/"

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

		if e = json.NewDecoder(resp.Body).Decode(&disk); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &disk, nil
}
