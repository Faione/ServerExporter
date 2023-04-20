package collector

import (
	"context"
	"regexp"

	"github.com/Faione/easyxporter"
	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	sshAuthInfoCollectorSubsystem = "sshauth"
)

var (
	sshAuthLogPath = easyxporter.CollectorFlags.String("collector.sshauth.path", "/var/log/secure", "SSH Auth log path")
)

// 解析 ssh auth 日志，根据accept/disconnect记录模糊计算ssh在线的用户
type sshAuthCollector struct {
	counterVec *prometheus.CounterVec
	livingVec  *prometheus.GaugeVec
	livingMap  map[string][]string

	logger *logrus.Logger
}

func NewSSHAuthCollector(logger *logrus.Logger) (easyxporter.AsyncCollector, error) {
	return &sshAuthCollector{
		counterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: easyxporter.GetNameSpace(),
				Subsystem: sshAuthInfoCollectorSubsystem,
				Name:      "count",
				Help:      "SSH auth count from /var/log/secure",
			},
			[]string{"auth_type", "username"},
		),
		livingVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: easyxporter.GetNameSpace(),
				Subsystem: sshAuthInfoCollectorSubsystem,
				Name:      "living",
				Help:      "SSH auth accept from /var/log/secure",
			},
			[]string{"auth_type", "username", "remote_ip", "remote_port"},
		),
		livingMap: make(map[string][]string),
		logger:    logger,
	}, nil
}

func (s *sshAuthCollector) Update(ch chan<- prometheus.Metric) error {
	s.counterVec.Collect(ch)
	s.livingVec.Collect(ch)
	return nil
}

func (s *sshAuthCollector) AsyncCollect(ctx context.Context) error {
	t, err := tail.TailFile(*sshAuthLogPath, tail.Config{Follow: true})
	if err != nil {
		return err
	}
	defer t.Cleanup()

	s.logger.Debug("collector ssh auth start to tail file: ", *sshAuthLogPath)

	lineParsers := s.lineParsers()
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case line := <-t.Lines:

			for _, parse := range lineParsers {
				if parse(line.Text) {
					break
				}
			}
		}
	}
	return nil
}

func (s *sshAuthCollector) lineParsers() (lineParsers []func(str string) bool) {
	reAccept := regexp.MustCompile(`sshd\[(\d+)\]: Accepted (password|publickey) for (\w+) from ([^ ]+) port (\d+)`)
	lineParsers = append(
		lineParsers,
		func(str string) bool {
			matches := reAccept.FindStringSubmatch(str)
			if len(matches) != 6 {
				return false
			}

			s.livingMap[matches[1]] = matches[2:]
			return true
		},
	)

	reSession := regexp.MustCompile(`sshd\[(\d+)\]: pam_unix\(sshd:session\): session (opened|closed)`)
	lineParsers = append(lineParsers, func(str string) bool {
		matches := reSession.FindStringSubmatch(str)
		if len(matches) != 3 {
			return false
		}

		labelVals, ok := s.livingMap[matches[1]]
		if !ok {
			return true
		}

		switch matches[2] {
		case "opened":
			{

				s.livingVec.WithLabelValues(labelVals...).Set(1)
				s.counterVec.WithLabelValues(labelVals[:2]...).Inc()

			}
		case "closed":
			{

				s.livingVec.DeleteLabelValues(labelVals...)
				delete(s.livingMap, matches[1])

			}
		}

		return true
	})

	return

}

func init() {
	easyxporter.RegisterAsyncCollector(sshAuthInfoCollectorSubsystem, true, NewSSHAuthCollector)
}
