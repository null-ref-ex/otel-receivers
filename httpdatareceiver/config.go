// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package httpdatareceiver // import "github.com/null-ref-ex/otel-receivers/httpdatareceiver"

import (
	"errors"
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"

	"github.com/null-ref-ex/otel-receivers/httpdatareceiver/internal/metadata"
)

// Predefined error responses for configuration validation failures
var (
	errMissingEndpoint = errors.New(`"endpoint" must be specified`)
	errInvalidEndpoint = errors.New(`"endpoint" must be in the form of <scheme>://<hostname>[:<port>]`)
	errMissingRequestBody = errors.New(`If using POST|PATCH HTTP verbs you must supply a "body" option`)
)

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	metadata.MetricsBuilderConfig           `mapstructure:",squash"`
	Targets                                 []*targetConfig `mapstructure:"targets"`
}

type targetConfig struct {
	confighttp.HTTPClientSettings `mapstructure:",squash"`
	Method                        string `mapstructure:"method"`
	Body						  string `mapstructure:"body"`
	JPath						  string `mapstructure:"jpath"`
}

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *targetConfig) Validate() error {
	var err error

	if cfg.Endpoint == "" {
		err = multierr.Append(err, errMissingEndpoint)
	} else {
		_, parseErr := url.ParseRequestURI(cfg.Endpoint)
		if parseErr != nil {
			err = multierr.Append(err, fmt.Errorf("%s: %w", errInvalidEndpoint.Error(), parseErr))
		}
	}

	if cfg.Method == "POST" || cfg.Method == "PATCH" {
		if cfg.Body == "" {
			err = multierr.Append(err, errMissingRequestBody)
		}
	} 

	return err
}

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *Config) Validate() error {
	var err error

	if len(cfg.Targets) == 0 {
		err = multierr.Append(err, errors.New("no targets configured"))
	}

	for _, target := range cfg.Targets {
		err = multierr.Append(err, target.Validate())
	}

	return err
}
