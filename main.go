package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "os"
    "time"

    "github.com/michaelklishin/rabbit-hole"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/Sirupsen/logrus"
)

const (
    namespace = "rabbitmq"
    configPath = "config.json"
)

var log = logrus.New()

// Listed available metrics
var (
    channelsTotal = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Namespace: namespace,
            Name:      "channels_total",
            Help:      "Total number of open channels.",
        },
    )
    connectionsTotal = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Namespace: namespace,
            Name:      "connections_total",
            Help:      "Total number of open connections.",
        },
    )
)

type Config struct {
    Nodes      *[]Node     `json:"nodes"`
    Port       string      `json:"port"`
    Interval   string      `json:"req_interval"`
}

type Node struct {
    Name       string      `json:"name"`
    Url        string      `json:"url"`
    Uname      string      `json:"uname"`
    Password   string      `json:"password"`
    Interval   string      `json:"req_interval,omitempty"`
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
        rmqc, err := rabbithole.NewClient(node.Url, node.Uname, node.Password)

        updateMetrics(rmqc)

        dt, err := time.ParseDuration(node.Interval)
        if err != nil {
            log.Warningln(err)
            dt = 30 * time.Second
        }
        time.Sleep(dt)
    }
}

func updateMetrics(client *rabbithole.Client) {
    r1, _ := client.ListConnections()
    r2, _ := client.ListChannels()

    channelsTotal.Set(float64(len(r1)))
    connectionsTotal.Set(float64(len(r2)))
}

func newConfig(path string) (*Config, error) {
    var config Config

    file, err := ioutil.ReadFile(path)
    if err != nil {
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
    log.Infof("Starting RabbitMQ exporter on port: %s.\n", config.Port)
    http.ListenAndServe(":" + config.Port, nil)
}

// Register metrics to Prometheus
func init() {
    prometheus.MustRegister(channelsTotal)
    prometheus.MustRegister(connectionsTotal)
}
