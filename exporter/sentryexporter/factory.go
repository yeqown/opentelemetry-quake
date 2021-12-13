// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sentryexporter

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // for debugging

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	typeStr = "sentry"
)

// NewFactory creates a factory for Sentry exporter.
func NewFactory() component.ExporterFactory {
	return exporterhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		exporterhelper.WithTraces(createTracesExporter),
	)
}

func createDefaultConfig() config.Exporter {
	return &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
	}
}

func createTracesExporter(
	_ context.Context,
	params component.ExporterCreateSettings,
	config config.Exporter,
) (component.TracesExporter, error) {
	sentryConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", config)
	}

	if isDebug() {
		go func() {
			println("pprof is listening on: 6060")
			if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
				println("[ERROR] pprof failed to start: %v", err)
			}
		}()
	}

	// Create exporter based on sentry config.
	exp, err := newSentryExporter(sentryConfig, params)
	return exp, err
}
