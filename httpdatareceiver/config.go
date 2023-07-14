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

	"github.com/ohler55/ojg/jp"
)

// Predefined error responses for configuration validation failures
var (
	errMissingEndpoint = errors.New(`"endpoint" must be specified`)
	errInvalidEndpoint = errors.New(`"endpoint" must be in the form of <scheme>://<hostname>[:<port>]`)
	errInvalidJPath = errors.New(`"jpath" must be valid jsonpath`)
	errMissingRequestBody = errors.New(`If using POST|PATCH HTTP verbs you must supply a "body" option`)
	errMissingMetricName = errors.New(`"metric" must be specified, this is the metric name to create for the target data`)
	errInvalidType = errors.New(`"type" must be 'hex' or 'numeric'`)
)

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	metadata.MetricsBuilderConfig           `mapstructure:",squash"`
	Targets                                 []*targetConfig `mapstructure:"targets"`
}

type targetConfig struct {
	confighttp.HTTPClientSettings        `mapstructure:",squash"`
	Timeout						  int    `mapstructure:"timeout"`
	Method                        string `mapstructure:"method"`
	Body						  string `mapstructure:"body"`
	JPath						  string `mapstructure:"jpath"`
	Type					  	  string `mapstructure:"type"`
	Metric					  	  string `mapstructure:"metric"`
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

	_, parseErr := jp.ParseString(cfg.JPath)
    if err != nil {
        err = multierr.Append(err, fmt.Errorf("%s: %w", errInvalidJPath.Error(), parseErr))
    }

	if cfg.Metric == "" {
		err = multierr.Append(err, errMissingMetricName)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 2 // in the absence of a defined timeout we set an aggressive one
	}

	if cfg.Type == "" {
		cfg.Type = "numeric" // in the absence of a defined type we assume numeric
	} else {
		if cfg.Type != "hex" && cfg.Type != "numeric" {
			err = multierr.Append(err, errInvalidType)
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
