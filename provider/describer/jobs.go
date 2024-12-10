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

func ListJobs(ctx context.Context, handler *RenderAPIHandler, stream *models.StreamSender) ([]models.Resource, error) {
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
			if err := processJobs(ctx, handler, service.ID, renderChan, &wg); err != nil {
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

func GetJob(ctx context.Context, handler *RenderAPIHandler, resourceID, serviceID string) (*models.Resource, error) {
	job, err := processJob(ctx, handler, resourceID, serviceID)
	if err != nil {
		return nil, err
	}
	value := models.Resource{
		ID:   job.ID,
		Name: job.Status,
		Description: JSONAllFieldsMarshaller{
			Value: job,
		},
	}
	return &value, nil
}

func processJobs(ctx context.Context, handler *RenderAPIHandler, serviceID string, renderChan chan<- models.Resource, wg *sync.WaitGroup) error {
	var jobs []model.JobDescription
	var jobListResponse []model.JobResponse
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"
	cursor := ""

	for {
		params := url.Values{}
		params.Set("limit", limit)
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		finalURL := fmt.Sprintf("%s%sjobs?%s", baseURL, serviceID, params.Encode())

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

			if e = json.NewDecoder(resp.Body).Decode(&jobListResponse); e != nil {
				return nil, fmt.Errorf("failed to decode response: %w", e)
			}
			for i, jobResp := range jobListResponse {
				jobs = append(jobs, jobResp.Job)
				if i == len(jobListResponse)-1 {
					cursor = jobResp.Cursor
				}
			}
			return resp, nil
		}
		err = handler.DoRequest(ctx, req, requestFunc)
		if err != nil {
			return fmt.Errorf("error during request handling: %w", err)
		}

		if len(jobListResponse) < 100 {
			break
		}
	}
	for _, job := range jobs {
		wg.Add(1)
		go func(job model.JobDescription) {
			defer wg.Done()
			value := models.Resource{
				ID:   job.ID,
				Name: job.Status,
				Description: JSONAllFieldsMarshaller{
					Value: job,
				},
			}
			renderChan <- value
		}(job)
	}
	return nil
}

func processJob(ctx context.Context, handler *RenderAPIHandler, resourceID, serviceID string) (*model.JobDescription, error) {
	var job model.JobDescription
	var resp *http.Response
	baseURL := "https://api.render.com/v1/services/"

	finalURL := fmt.Sprintf("%s%s/jobs/%s", baseURL, serviceID, resourceID)
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

		if e = json.NewDecoder(resp.Body).Decode(&job); e != nil {
			return nil, fmt.Errorf("failed to decode response: %w", e)
		}
		return resp, e
	}

	err = handler.DoRequest(ctx, req, requestFunc)
	if err != nil {
		return nil, fmt.Errorf("error during request handling: %w", err)
	}
	return &job, nil
}
