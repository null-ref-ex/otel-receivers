package httpdatareceiver

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr         = "httpdatareceiver"
	defaultInterval = 1 * time.Minute
)

var errConfigNotOK = errors.New("config was not the correct receiver config")

// NewFactory creates a new receiver factory
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	return &Config{
		Interval: string(defaultInterval),
	}
}

func createMetricsReceiver(_ context.Context, params receiver.CreateSettings, rConf component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return nil, nil
	// cfg, ok := rConf.(*Config)
	// if !ok {
	// 	return nil, errConfigNotOK
	// }

	// receiver := newReceiver(cfg, params)
	// receiver, err := scraperhelper.NewScraper(metadata.Type, receiver.scrape, scraperhelper.WithStart(receiver.start))
	// if err != nil {
	// 	return nil, err
	// }

	// return scraperhelper.NewScraperControllerReceiver(&cfg.ScraperControllerSettings, params, consumer, scraperhelper.AddScraper(scraper))
}
