package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace  = "rabbitmq"
	configPath = "config.json"
)

var log = logrus.New()

// Listed available metrics
var (
	connectionsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connections_total",
			Help:      "Total number of open connections.",
		},
	)
	channelsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "channels_total",
			Help:      "Total number of open channels.",
		},
	)
	queuesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queues_total",
			Help:      "Total number of queues in use.",
		},
	)
	consumersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumers_total",
			Help:      "Total number of message consumers.",
		},
	)
	exchangesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exchanges_total",
			Help:      "Total number of exchanges in use.",
		},
	)
)

type Config struct {
	Nodes    *[]Node `json:"nodes"`
	Port     string  `json:"port"`
	Interval string  `json:"req_interval"`
}

type Node struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Uname    string `json:"uname"`
	Password string `json:"password"`
	Interval string `json:"req_interval,omitempty"`
}

func unpackMetrics(d *json.Decoder) map[string]float64 {
	var output map[string]interface{}

	if err := d.Decode(&output); err != nil {
		log.Error(err)
	}
	metrics := make(map[string]float64)

	for k, v := range output["object_totals"].(map[string]interface{}) {
		metrics[k] = v.(float64)
	}
	return metrics
}

func getOverview(hostname, username, password string) *json.Decoder {
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostname+"/api/overview", nil)
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}
	return json.NewDecoder(resp.Body)
}

func updateNodesStats(config *Config) {
	for _, node := range *config.Nodes {

		if len(node.Interval) == 0 {
			node.Interval = config.Interval
		}
		go runRequestLoop(node)
	}
}

func runRequestLoop(node Node) {
	for {
		decoder := getOverview(node.Url, node.Uname, node.Password)
		metrics := unpackMetrics(decoder)

		updateMetrics(metrics)
		log.Info("Metrics updated successfully.")

		dt, err := time.ParseDuration(node.Interval)
		if err != nil {
			log.Warn(err)
			dt = 30 * time.Second
		}
		time.Sleep(dt)
	}
}

func updateMetrics(metrics map[string]float64) {
	channelsTotal.Set(metrics["channels"])
	connectionsTotal.Set(metrics["connections"])
	consumersTotal.Set(metrics["consumers"])
	queuesTotal.Set(metrics["queues"])
	exchangesTotal.Set(metrics["exchanges"])
}

func newConfig(path string) (*Config, error) {
	var config Config

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	return &config, err
}

func main() {
	log.Out = os.Stdout
	config, _ := newConfig(configPath)
	updateNodesStats(config)

	http.Handle("/metrics", prometheus.Handler())
	log.Infof("Starting RabbitMQ exporter on port: %s.", config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}

// Register metrics to Prometheus
func init() {
	prometheus.MustRegister(channelsTotal)
	prometheus.MustRegister(connectionsTotal)
	prometheus.MustRegister(queuesTotal)
	prometheus.MustRegister(exchangesTotal)
	prometheus.MustRegister(consumersTotal)
}
