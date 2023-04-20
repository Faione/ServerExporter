package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Faione/easyxporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	sensorCollectorSubsystem = "sensor"
)

var (
	sensorStatePath = easyxporter.Flags().String("collector.sensor.state.path", "", "StateLog path")
	sensorStateHead = easyxporter.Flags().Bool("collector.sensor.state.head", true, "Table head of statelog")
)

type sensorColletor struct {
	stateDesc *prometheus.Desc
}

func NewSensorCollector(logger *logrus.Logger) (easyxporter.Collector, error) {
	return &sensorColletor{
		stateDesc: nil,
	}, nil
}

func stateLogToLabelValue(line string) []string {
	splits := strings.Split(line, "|")

	for i, split := range splits {
		splits[i] = strings.TrimSpace(split)
	}

	return splits
}

func fetchNewlyStates() error {
	var states [][]string
	f, err := os.Open(*sensorStatePath)
	if err != nil {
		return err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	// 跳过表头
	if *sensorStateHead {
		scanner.Scan()
	}

	for scanner.Scan() {

		lvals := stateLogToLabelValue(scanner.Text())
		if len(lvals) == 0 {
			continue
		}

		states = append(states, lvals)
	}

	if len(states) != 0 {
		sensorStates = states
	}

	return scanner.Err()
}

var sensorStates [][]string

func (s *sensorColletor) Update(ch chan<- prometheus.Metric) (err error) {

	if err := fetchNewlyStates(); err != nil && len(sensorStates) == 0 {
		return errors.New("file not exist or busy")
	}

	// 初始化desc
	if s.stateDesc == nil {
		lnames := make([]string, len(sensorStates[0]))

		for i := range lnames {
			lnames[i] = fmt.Sprintf("col_%d", i)
		}

		s.stateDesc = prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), sensorCollectorSubsystem, "reading"),
			"Sensor Reading from sensor state log",
			lnames, nil,
		)

	}

	for _, state := range sensorStates {
		if len(state) != len(sensorStates[0]) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			s.stateDesc,
			prometheus.CounterValue,
			0,
			state...,
		)
	}

	return nil
}

func init() {
	easyxporter.RegisterCollector(sensorCollectorSubsystem, true, NewSensorCollector)
}
