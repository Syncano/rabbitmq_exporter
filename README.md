# RabbitMQ Exporter

Prometheus exporter for RabbitMQ metrics, based on RabbitMQ HTTP API.

### Dependencies

* Prometheus [client](https://github.com/prometheus/client_golang) for Golang
* RabbitMQ HTTP API [client](https://github.com/michaelklishin/rabbit-hole)
* [Logging](https://github.com/Sirupsen/logrus)

### Usage

    $ go run main.go

### Metrics

Currently available:

    * channels_total
    * connections_total
