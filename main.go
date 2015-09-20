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
	connectionsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connections_total",
			Help:      "Total number of open connections.",
		},
		[]string{
			// Which node was checked?
			"node",
		},
	)
	channelsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "channels_total",
			Help:      "Total number of open channels.",
		},
		[]string{
			"node",
		},
	)
	queuesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queues_total",
			Help:      "Total number of queues in use.",
		},
		[]string{
			"node",
		},
	)
	consumersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumers_total",
			Help:      "Total number of message consumers.",
		},
		[]string{
			"node",
		},
	)
	exchangesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exchanges_total",
			Help:      "Total number of exchanges in use.",
		},
		[]string{
			"node",
		},
	)
	messagesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "messages_total",
			Help:      "Total number of messages in all queues.",
		},
		[]string{
			"node",
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

func sendApiRequest(hostname, username, password, query string) *json.Decoder {
    client := &http.Client{}
    req, err := http.NewRequest("GET", hostname+query, nil)
    req.SetBasicAuth(username, password)

    resp, err := client.Do(req)

    if err != nil {
        log.Error(err)
    }
    return json.NewDecoder(resp.Body)
}

func getOverview(hostname, username, password string) {
	decoder := sendApiRequest(hostname, username, password, "/api/overview")
    response := decodeObj(decoder)

    metrics := make(map[string]float64)
    for k, v := range response["object_totals"].(map[string]interface{}) {
        metrics[k] = v.(float64)
    }
    nodename, _ := response["node"].(string)

    channelsTotal.WithLabelValues(nodename).Set(metrics["channels"])
    connectionsTotal.WithLabelValues(nodename).Set(metrics["connections"])
    consumersTotal.WithLabelValues(nodename).Set(metrics["consumers"])
    queuesTotal.WithLabelValues(nodename).Set(metrics["queues"])
    exchangesTotal.WithLabelValues(nodename).Set(metrics["exchanges"])
}

func getNumberOfMessages(hostname, username, password string) {
	decoder := sendApiRequest(hostname, username, password, "/api/queues")
    response := decodeObjArray(decoder)
    nodename := response[0]["node"].(string)

	total_messages := 0.0
	for _, v := range response {
        total_messages += v["messages"].(float64)
	}
	messagesTotal.WithLabelValues(nodename).Set(total_messages)
}

func decodeObj(d *json.Decoder) map[string]interface{} {
    var response map[string]interface{}

    if err := d.Decode(&response); err != nil {
        log.Error(err)
    }
    return response
}

func decodeObjArray(d *json.Decoder) []map[string]interface{} {
    var response []map[string]interface{}

    if err := d.Decode(&response); err != nil {
        log.Error(err)
    }
    return response
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
        getOverview(node.Url, node.Uname, node.Password)
        getNumberOfMessages(node.Url, node.Uname, node.Password)

        log.Info("Metrics updated successfully.")

		dt, err := time.ParseDuration(node.Interval)
		if err != nil {
			log.Warn(err)
			dt = 30 * time.Second
		}
		time.Sleep(dt)
	}
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>RabbitMQ Exporter</title></head>
             <body>
             <h1>RabbitMQ Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
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
	prometheus.MustRegister(messagesTotal)
}
