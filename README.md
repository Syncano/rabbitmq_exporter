# RabbitMQ Exporter

Prometheus exporter for RabbitMQ metrics, based on RabbitMQ HTTP API.

### Dependencies

* Prometheus [client](https://github.com/prometheus/client_golang) for Golang
* RabbitMQ HTTP API [client](https://github.com/michaelklishin/rabbit-hole)
* [Logging](https://github.com/Sirupsen/logrus)

### Setting up locally

1. You need **RabbitMQ**. For local setup I recommend this [docker box](https://github.com/mikaelhg/docker-rabbitmq). It's "one-click" solution.

2. For OS-specific **Docker** installation checkout these [instructions](https://docs.docker.com/installation/).

3. Building rabbitmq_exporter:

        $ docker build -t rabbitmq_exporter .

4. Running:

        $ docker run --publish 6060:9672 --rm rabbitmq_exporter

Now your metrics are available through [http://localhost:6060/metrics](http://localhost:6060/metrics).

### Metrics

Currently available:

* channels - total number
* connections - total number
