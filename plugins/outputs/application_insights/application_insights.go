package application_insights

import (
	"log"
	"time"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type ApplicationInsights struct {
	InstrumentationKey string
	Timeout            internal.Duration
}

var (
	sampleConfig = `
## Instrumentation key of the Application Insights resource.
instrumentationKey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxx"

## Timeout on close. If not provided, will default to 5s. 0s means no timeout (not recommended).
# timeout = "5s"
`
	client appinsights.TelemetryClient
)

func (a *ApplicationInsights) SampleConfig() string {
	return sampleConfig
}

func (a *ApplicationInsights) Description() string {
	return "Send telegraf metrics to Azure Application Insights"
}

func (a *ApplicationInsights) Connect() error {
	client = appinsights.NewTelemetryClient(a.InstrumentationKey)
	return nil
}

func (a *ApplicationInsights) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		if metric.IsAggregate() {
			// TODO: Need a better library support before parsing the aggregated metric
			// Submitted a feature request: https://github.com/influxdata/telegraf/issues/3670
			log.Println("Aggregated metric is not supported by Application Insights output")
			return nil
		} else {
			metricName := metric.Name()
			for fieldName, value := range metric.Fields() {
				fullMetricName := metricName + "_" + fieldName
				telemetry := appinsights.NewMetricTelemetry(fullMetricName, value.(float64))
				telemetry.Properties = metric.Tags()
				telemetry.Timestamp = metric.Time()

				client.Track(telemetry)
			}
		}
	}

	return nil
}

func (a *ApplicationInsights) Close() error {
	select {
	case <-client.Channel().Close(0 * time.Second):
		log.Println("Application Insights output closed successfully")
	case <-time.After(a.Timeout.Duration):
		log.Println("Application Insights output timed out after", a.Timeout.Duration)
	}

	return nil
}

func init() {
	outputs.Add("application_insights", func() telegraf.Output {
		return &ApplicationInsights{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
