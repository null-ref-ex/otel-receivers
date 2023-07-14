// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package httpdatareceiver // import "github.com/null-ref-ex/otel-receivers/httpdatareceiver"

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
	"strconv"
	"ioutil"
	"bytes"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/null-ref-ex/otel-receivers/httpdatareceiver/internal/metadata"

	"github.com/ohler55/ojg/jp"
    "github.com/ohler55/ojg/oj"
)

var (
	errClientNotInit    = errors.New("client not initialized")
	httpResponseClasses = map[string]int{"1xx": 1, "2xx": 2, "3xx": 3, "4xx": 4, "5xx": 5}
	multipleResultsError = errors.New("A JPath expression should yield only one result")
)

type httpdataScraper struct {
	clients  []*http.Client
	cfg      *Config
	settings component.TelemetrySettings
	mb       *metadata.MetricsBuilder
}

// start starts the scraper by creating a new HTTP Client on the scraper
func (h *httpdataScraper) start(_ context.Context, host component.Host) (err error) {
	for _, target := range h.cfg.Targets {
		client, clentErr := target.ToClient(host, h.settings)
		if clentErr != nil {
			err = multierr.Append(err, clentErr)
		}
		h.clients = append(h.clients, client)
	}
	return
}

// scrape connects to the endpoint and produces metrics based on the response
func (h *httpdataScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	if h.clients == nil || len(h.clients) == 0 {
		return pmetric.NewMetrics(), errClientNotInit
	}

	var wg sync.WaitGroup
	wg.Add(len(h.clients))
	var mux sync.Mutex

	for idx, client := range h.clients {
		go func(targetClient *http.Client, targetIndex int) {
			defer wg.Done()

			now := pcommon.NewTimestampFromTime(time.Now())

			body := http.NoBody
			if h.cfg.Targets[targetIndex].Body != "" {
				body := bytes.NewBuffer([]byte(h.cfg.Targets[targetIndex].Body))
			}
			req, err := http.NewRequestWithContext(ctx, h.cfg.Targets[targetIndex].Method, h.cfg.Targets[targetIndex].Endpoint, body)
			if err != nil {
				h.settings.Logger.Error("failed to create request", zap.Error(err))
				return
			}

			start := time.Now()
			resp, err := targetClient.Do(req)
			mux.Lock()
			h.mb.RecordHttpdataDurationDataPoint(now, time.Since(start).Milliseconds(), h.cfg.Targets[targetIndex].Endpoint)

			statusCode := 0
			if err != nil {
				h.mb.RecordHttpdataErrorDataPoint(now, int64(1), h.cfg.Targets[targetIndex].Endpoint, err.Error())
			} else {
				statusCode = resp.StatusCode
			}

			for class, intVal := range httpResponseClasses {
				if statusCode/100 == intVal {
					h.mb.RecordHttpdataStatusDataPoint(now, int64(1), h.cfg.Targets[targetIndex].Endpoint, int64(statusCode), req.Method, class)
				} else {
					h.mb.RecordHttpdataStatusDataPoint(now, int64(0), h.cfg.Targets[targetIndex].Endpoint, int64(statusCode), req.Method, class)
				}
			}

			// if the user supplied a JPath to 2xx responses
			if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
				// read response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					h.settings.Logger.Error("unable to get response body", zap.Error(err))
				}
				// close response body
				response.Body.Close()
				obj, err := oj.ParseString(string(body))
				if err != nil {
					h.settings.Logger.Error("unable to deserialize response body", zap.Error(err))
				} else {
					x, err := jp.ParseString(h.cfg.Targets[targetIndex].JPath)
					if err != nil {
						h.settings.Logger.Error("unable to parse jpath", zap.Error(err))
					} else {
						var ys []any = x.Get(obj)
						arrayLength := len(ys)
						if arrayLength > 1 {
							h.settings.Logger.Error(fmt.Sprintf("jpath yielded %s results", arrayLength), zap.Error(multipleResultsError))
						} else {
							dataPoint := ys[0]
							if h.cfg.Targets[targetIndex].Type == "hex" {
								value, err := strconv.ParseInt64(dataPoint, 16, 64)
								if err != nil {
									h.settings.Logger.Error(fmt.Sprintf("%s could not be converted as hex -> int", dataPoint), zap.Error(err))
								} else {
									dataPoint = value
								}
								h.mb.RecordHttpdataMetricDataPoint(now, int64(dataPoint), h.cfg.Targets[targetIndex].Metric)
							}
						}
					}					
				}
			}

			mux.Unlock()
		}(client, idx)
	}

	wg.Wait()

	return h.mb.Emit(), nil
}

func newScraper(conf *Config, settings receiver.CreateSettings) *httpdataScraper {
	return &httpdataScraper{
		cfg:      conf,
		settings: settings.TelemetrySettings,
		mb:       metadata.NewMetricsBuilder(conf.MetricsBuilderConfig, settings),
	}
}
