package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/stew/objects"
	"gopkg.in/yaml.v2"
)

const (
	hekaStatsHelpPrefix = "heka stats counter "
)

var (
	listeningAddress = flag.String("web.listen-address", ":9111", "Address on which to expose metrics and web interface.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "Path under which to expose Prometheus metrics.")
	target           = flag.String("heka.url", "", "URL of expvar endpoint to expose.")
	namespace        = flag.String("heka.namespace", "heka", "Namespace/prefix for expvar metrics.")
	errorCounter     = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: *namespace,
		Name:      "errors_total",
		Help:      "Number of errors when collecting heka metrics.",
	})
)

// type expVars map[string]interface{}

type collector struct {
	target  *url.URL
	client  *http.Client
	exports map[string]*metric
}

type valueType prometheus.ValueType

func (t *valueType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	str := ""
	if err := unmarshal(&str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "counter":
		*t = valueType(prometheus.CounterValue)
	case "gauge":
		*t = valueType(prometheus.GaugeValue)
	default:
		*t = valueType(prometheus.UntypedValue)
	}
	return nil
}

type metric struct {
	Name      string // the section of the heka config that will be used to add a name label
  MetricName string
	Subsystem string
	Help   string
	Type   prometheus.ValueType
	Value  float64
}

// NewCollector returns a collector implementing prometheus.Collector.
func NewCollector(target *url.URL, exports map[string]*metric) *collector {
	return &collector{
		target:  target,
		client:  &http.Client{},
		exports: exports,
	}
}

func hekaToPrometheusTypes(rep string) prometheus.ValueType {
  switch rep {
  case "count":
    return prometheus.CounterValue
  case "ns":
    return prometheus.GaugeValue
  case "B":
    return prometheus.GaugeValue
  }
  return prometheus.UntypedValue
}

func getSystemMetrics(data map[string]interface{}) []interface{} {
  var myMetrics = []interface{}{}
  systems := []string{"decoders", "encoders", "globals", "outputs"}
  for _, system := range systems {
    mapList := data[system]
    maps := mapList.([]interface{})
    for _, m := range maps {
      y := m.(map[string]interface{})
      myName := y["Name"].(string)
      for key, val := range y {
        var myMetric metric
        switch key {
        case "Name":
          continue
        default:
          myMetric.Name = myName
					myMetric.Subsystem = system
          myMetric.MetricName = toUnderscore(key)
          myMetric.Help = key + " for " + myMetric.MetricName + " in " + system

          valueMap := val.(map[string]interface{})
          rep := valueMap["representation"].(string)

          myMetric.Value = valueMap["value"].(float64)
          myMetric.Type = hekaToPrometheusTypes(rep)
          myMetrics = append(myMetrics, myMetric)
        }
      }
    }
  }
  return myMetrics
}

// Collect implements prometheus.Collector.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	data, err := c.get()
	if err != nil {
		errorCounter.Inc()
		log.Println(err)
		return
	}

	stats, err := objects.NewMapFromJSON(string(data))
	if err != nil {
		errorCounter.Inc()
		log.Println(err)
		return
	}
	collectHekaStats(stats, ch)
}

func toUnderscore(camelCase string) string {
	ret := []rune{}
	add := false
	for _, r := range camelCase {
		c := r
		if unicode.IsUpper(c) {
			if add {
				ret = append(ret, '_')
				add = false
			}
			c = unicode.ToLower(c)
		} else {
			add = true
		}
		ret = append(ret, c)
	}
	return string(ret)
}

func collectHekaStats(stats map[string]interface{}, ch chan<- prometheus.Metric) {
	metrics := getSystemMetrics(stats)
	for _, m := range metrics {
		metric := m.(metric)
		name := prometheus.BuildFQName(*namespace, metric.Subsystem, metric.MetricName)
		//help := metric.Help
		//var labels map[string]string
		label_names := []string{"name"}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(name, "", label_names, nil),
			metric.Type,
			metric.Value,
			metric.Name,
		)
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- errorCounter.Desc()
}

func (c *collector) get() ([]byte, error) {
	resp, err := c.client.Get(c.target.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status %d unexpected", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func readConfig(file string) (metrics map[string]*metric, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return metrics, yaml.Unmarshal(data, &metrics)
}

func main() {
	flag.Parse()
	t, err := url.Parse(*target)
	if *target == "" {
		log.Fatal("-heka.url required")
	}
	if err != nil {
		log.Fatal(err)
	}
	if t.Host == "" {
		log.Fatal("-heka.url invalid")
	}

	metrics := map[string]*metric{}

	http.Handle(*metricsPath, prometheus.Handler())
	prometheus.MustRegister(NewCollector(t, metrics))
	log.Printf("Exposing heka metrics for %#v on %s%s", t, *listeningAddress, *metricsPath)
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
