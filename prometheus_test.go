package prometheusmetrics

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

func TestPrometheusRegistration(t *testing.T) {
	defaultRegistry := prometheus.DefaultRegisterer
	prometheusLabels := prometheus.Labels{}
	pClient := NewPrometheusProvider(metrics.DefaultRegistry, "test", "subsys", prometheusLabels, defaultRegistry, 1*time.Second)
	if pClient.promRegistry != defaultRegistry {
		t.Fatalf("Failed to pass prometheus registry to go-metrics provider")
	}
}

func TestUpdatePrometheusMetricsOnce(t *testing.T) {
	prometheusRegistry := prometheus.NewRegistry()
	prometheusLabels := prometheus.Labels{}
	metricsRegistry := metrics.NewRegistry()
	pClient := NewPrometheusProvider(metricsRegistry, "test", "subsys", prometheusLabels, prometheusRegistry, 1*time.Second)
	metricsRegistry.Register("counter", metrics.NewCounter())
	pClient.UpdatePrometheusMetricsOnce()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "test",
		Subsystem: "subsys",
		Name:      "counter",
		Help:      "counter",
	})
	err := pClient.promRegistry.Register(gauge)
	if err == nil {
		t.Fatalf("Go-metrics registry didn't get registered to prometheus registry")
	}

}

func TestUpdatePrometheusMetrics(t *testing.T) {
	prometheusRegistry := prometheus.NewRegistry()
	prometheusLabels := prometheus.Labels{}
	metricsRegistry := metrics.NewRegistry()
	pClient := NewPrometheusProvider(metricsRegistry, "test", "subsys", prometheusLabels, prometheusRegistry, 1*time.Second)
	metricsRegistry.Register("counter", metrics.NewCounter())
	go pClient.UpdatePrometheusMetrics()
	time.Sleep(2 * time.Second)
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "test",
		Subsystem: "subsys",
		Name:      "counter",
		Help:      "counter",
	})
	err := pClient.promRegistry.Register(gauge)
	if err == nil {
		t.Fatalf("Go-metrics registry didn't get registered to prometheus registry")
	}

}

func TestPrometheusCounterGetUpdated(t *testing.T) {
	prometheusRegistry := prometheus.NewRegistry()
	prometheusLabels := prometheus.Labels{}
	metricsRegistry := metrics.NewRegistry()
	pClient := NewPrometheusProvider(metricsRegistry, "test", "subsys", prometheusLabels, prometheusRegistry, 3*time.Second)
	cntr := metrics.NewCounter()
	metricsRegistry.Register("counter", cntr)
	cntr.Inc(2)
	go pClient.UpdatePrometheusMetrics()
	cntr.Inc(13)
	time.Sleep(5 * time.Second)
	metrics, _ := prometheusRegistry.Gather()
	serialized := fmt.Sprint(metrics[0])
	expected := fmt.Sprintf("name:\"test_subsys_counter\" help:\"counter\" type:COUNTER metric:<counter:<value:%d > > ", cntr.Count())
	if serialized != expected {
		t.Fatalf("Go-metrics value and prometheus metrics value do not match")
	}
}

func TestPrometheusGaugeGetUpdated(t *testing.T) {
	prometheusRegistry := prometheus.NewRegistry()
	prometheusLabels := prometheus.Labels{}
	metricsRegistry := metrics.NewRegistry()
	pClient := NewPrometheusProvider(metricsRegistry, "test", "subsys", prometheusLabels, prometheusRegistry, 1*time.Second)
	gm := metrics.NewGauge()
	metricsRegistry.Register("gauge", gm)
	gm.Update(2)
	go pClient.UpdatePrometheusMetrics()
	gm.Update(13)
	time.Sleep(5 * time.Second)
	metrics, _ := prometheusRegistry.Gather()
	if len(metrics) == 0 {
		t.Fatalf("prometheus was unable to register the metric")
	}
	serialized := fmt.Sprint(metrics[0])
	expected := fmt.Sprintf("name:\"test_subsys_gauge\" help:\"gauge\" type:GAUGE metric:<gauge:<value:%d > > ", gm.Value())
	if serialized != expected {
		t.Fatalf("Go-metrics value and prometheus metrics value do not match")
	}
}
