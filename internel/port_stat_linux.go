package internal

import (
	"strconv"
	"sync"

	"github.com/Faione/easyxporter"
	"github.com/bastjan/netstat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	netStatCollectorSubsystem = "net"
)

// TODO: filter
type filter func(conn *netstat.Connection) bool

// 通过 netstat 获取当前TCP/UDP连接数量
type portStatCollector struct {
	filters []filter

	portInfo *prometheus.Desc
}

func NewPortStatCollector(logger *logrus.Logger) (easyxporter.Collector, error) {
	return &portStatCollector{
		portInfo: prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), netStatCollectorSubsystem, "port_conntection"),
			"Port conntection from netstat",
			[]string{"uid", "pid", "cmd", "protocol", "ip", "port", "remote_ip", "remote_port", "state"}, nil,
		),
		filters: nil,
	}, nil
}

func connectionToLabelValue(conn *netstat.Connection) []string {
	labelValues := make([]string, 9)

	// uid
	labelValues[0] = conn.UserID
	// pid
	labelValues[1] = strconv.Itoa(conn.Pid)
	// cmd
	if len(conn.Cmdline) > 0 {
		labelValues[2] = conn.Cmdline[0]

	} else {
		labelValues[2] = ""
	}

	// protocol
	labelValues[3] = conn.Protocol.Name
	// ip
	labelValues[4] = conn.IP.String()
	// port
	labelValues[5] = strconv.Itoa(conn.Port)
	// remote_ip
	labelValues[6] = conn.RemoteIP.String()
	// remote_port
	labelValues[7] = strconv.Itoa(conn.RemotePort)
	// state
	labelValues[8] = conn.State.String()

	return labelValues

}

func (p *portStatCollector) Update(ch chan<- prometheus.Metric) error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		tcpStats, err := netstat.TCP.Connections()
		if err != nil {
			return
		}

		for _, tcp := range tcpStats {
			ch <- prometheus.MustNewConstMetric(
				p.portInfo,
				prometheus.CounterValue,
				1,
				connectionToLabelValue(tcp)...,
			)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		udpStats, err := netstat.UDP.Connections()
		if err != nil {
			return
		}

		for _, upd := range udpStats {
			ch <- prometheus.MustNewConstMetric(
				p.portInfo,
				prometheus.CounterValue,
				1,
				connectionToLabelValue(upd)...,
			)
		}
	}()

	wg.Wait()

	return nil
}

func init() {
	easyxporter.RegisterCollector(netStatCollectorSubsystem, true, NewPortStatCollector)
}
