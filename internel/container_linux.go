package internal

import (
	"context"
	"time"

	"github.com/Faione/easyxporter"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	containerInfoCollectorSubsystem = "container"

	timeout = 10
)

type containerInfoCollector struct {
	containerTotal *prometheus.Desc
	containerCount *prometheus.Desc
	cli            *docker.Client
}

func NewContainerInfoCollector(logger *logrus.Logger) (easyxporter.Collector, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &containerInfoCollector{
		containerTotal: prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), containerInfoCollectorSubsystem, "total"),
			"Container total count from docker",
			nil, nil,
		),
		containerCount: prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), containerInfoCollectorSubsystem, "count"),
			"Container count from docker",
			[]string{"state"}, nil,
		),
		cli: cli,
	}, nil
}

func (c *containerInfoCollector) Update(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	containers, err := c.cli.ListContainers(docker.ListContainersOptions{All: true, Context: ctx})
	if err != nil {
		return err
	}

	stateMap := make(map[string]float64)
	for _, container := range containers {

		if _, ok := stateMap[container.State]; !ok {
			stateMap[container.State] = 1
		} else {
			stateMap[container.State] += 1
		}

	}

	for state, v := range stateMap {
		ch <- prometheus.MustNewConstMetric(
			c.containerCount,
			prometheus.GaugeValue,
			v,
			state,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.containerTotal,
		prometheus.GaugeValue,
		float64(len(containers)),
	)
	return nil
}

func init() {
	easyxporter.RegisterCollector(containerInfoCollectorSubsystem, true, NewContainerInfoCollector)
}
