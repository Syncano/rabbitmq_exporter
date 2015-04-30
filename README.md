# RabbitMQ Exporter

Prometheus exporter for RabbitMQ metrics, based on RabbitMQ HTTP API.

## Dependencies

Prometheus client for Golang:

    github.com/prometheus/client_golang/prometheus

RabbitMQ HTTP API client:

    github.com/michaelklishin/rabbit-hole

Logging:

    github.com/Sirupsen/logrus

## Usage

    $ go run main.go

## Metrics

Currently available:

    * channels_total
    * connections_total
