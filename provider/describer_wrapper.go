package provider

import (
	"errors"
	model "github.com/opengovern/og-describer-template/pkg/sdk/models"
	"github.com/opengovern/og-describer-template/provider/configs"
	"github.com/opengovern/og-describer-template/provider/describer"
	"github.com/opengovern/og-util/pkg/describe/enums"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"time"
)

// DescribeListByRender A wrapper to pass render authorization to describer functions
func DescribeListByRender(describe func(context.Context, *describer.RenderAPIHandler, *model.StreamSender) ([]model.Resource, error)) model.ResourceDescriber {
	return func(ctx context.Context, cfg configs.IntegrationCredentials, triggerType enums.DescribeTriggerType, additionalParameters map[string]string, stream *model.StreamSender) ([]model.Resource, error) {
		ctx = describer.WithTriggerType(ctx, triggerType)

		var err error
		// Check for the api key
		if cfg.APIKey == "" {
			return nil, errors.New("api key must be configured")
		}

		renderAPIHandler := describer.NewRenderAPIHandler(cfg.APIKey, rate.Every(time.Minute/400), 1, 10, 5, 5*time.Minute)

		// Get values from describer
		var values []model.Resource
		result, err := describe(ctx, renderAPIHandler, stream)
		if err != nil {
			return nil, err
		}
		values = append(values, result...)
		return values, nil
	}
}

// DescribeSingleByRender A wrapper to pass render authorization to describer functions
func DescribeSingleByRender(describe func(context.Context, *describer.RenderAPIHandler, string) (*model.Resource, error)) model.SingleResourceDescriber {
	return func(ctx context.Context, cfg configs.IntegrationCredentials, triggerType enums.DescribeTriggerType, additionalParameters map[string]string, resourceID string) (*model.Resource, error) {
		ctx = describer.WithTriggerType(ctx, triggerType)

		var err error
		// Check for the api key
		if cfg.APIKey == "" {
			return nil, errors.New("api key must be configured")
		}

		renderAPIHandler := describer.NewRenderAPIHandler(cfg.APIKey, rate.Every(time.Minute/400), 1, 10, 5, 5*time.Minute)

		// Get value from describer
		value, err := describe(ctx, renderAPIHandler, resourceID)
		if err != nil {
			return nil, err
		}
		return value, nil
	}
}
