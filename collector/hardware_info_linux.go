package collector

import (
	"errors"
	"os/user"
	"strconv"

	"github.com/Faione/easyxporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/zcalusic/sysinfo"
)

const (
	hardwareInfoCollectorSubsystem = "hardware"
)

// 通过 sysinfo 获取server的硬件信息
type hardwareInfoCollector struct {
	nodeInfo    *prometheus.Desc
	osInfo      *prometheus.Desc
	kernelInfo  *prometheus.Desc
	productInfo *prometheus.Desc
	boardInfo   *prometheus.Desc
	chassisInfo *prometheus.Desc
	biosInfo    *prometheus.Desc
	cpuInfo     *prometheus.Desc
	memoryInfo  *prometheus.Desc
	storageInfo *prometheus.Desc
	networkInfo *prometheus.Desc
}

func NewHardwareInfoCollector(logger *logrus.Logger) (easyxporter.Collector, error) {
	current, err := user.Current()
	if err != nil {
		logrus.Fatal(err)
	}

	if current.Uid != "0" {
		return nil, errors.New("requires superuser privilege")
	}

	return &hardwareInfoCollector{
		nodeInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "node"),
			"Node information from sysinfo",
			[]string{"hostname", "machineid"}, nil,
		),
		osInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "os"),
			"OS information from sysinfo",
			[]string{"name", "vendor", "version", "release", "architecture"}, nil,
		),
		kernelInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "kernel"),
			"Kernel information from sysinfo",
			[]string{"release", "version", "architecture"}, nil,
		),
		productInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "product"),
			"Product information from sysinfo",
			[]string{"name", "vendor", "version", "serial"}, nil,
		),
		boardInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "board"),
			"Board information from sysinfo",
			[]string{"name", "vendor", "version", "serial", "assettag"}, nil,
		),
		chassisInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "chassis"),
			"Chassis information from sysinfo",
			[]string{"type", "vendor", "version", "serial", "assettag"}, nil,
		),
		biosInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "bios"),
			"Bios information from sysinfo",
			[]string{"vendor", "version", "date"}, nil,
		),
		cpuInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "cpu"),
			"Cpu information from sysinfo",
			[]string{"vendor", "model", "speed", "cache", "cpus", "cores", "threads"}, nil,
		),
		memoryInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "memory"),
			"Memory information from sysinfo",
			[]string{"type", "speed", "size"}, nil,
		),
		storageInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "storage"),
			"Storage information from sysinfo",
			[]string{"name", "driver", "vendor", "model", "serial", "size"}, nil,
		),
		networkInfo: prometheus.NewDesc(
			prometheus.BuildFQName(RootNamespace, hardwareInfoCollectorSubsystem, "network"),
			"Network information from sysinfo",
			[]string{"name", "driver", "macaddress", "port", "speed"}, nil,
		),
	}, nil

}

func (h *hardwareInfoCollector) Update(ch chan<- prometheus.Metric) error {
	var si sysinfo.SysInfo
	si.GetSysInfo()

	ch <- prometheus.MustNewConstMetric(
		h.nodeInfo,
		prometheus.CounterValue,
		1,
		si.Node.Hostname, si.Node.MachineID,
	)

	ch <- prometheus.MustNewConstMetric(
		h.osInfo,
		prometheus.CounterValue,
		1,
		si.OS.Name, si.OS.Vendor, si.OS.Version, si.OS.Release, si.OS.Architecture,
	)

	ch <- prometheus.MustNewConstMetric(
		h.kernelInfo,
		prometheus.CounterValue,
		1,
		si.Kernel.Release, si.Kernel.Version, si.Kernel.Architecture,
	)

	ch <- prometheus.MustNewConstMetric(
		h.productInfo,
		prometheus.CounterValue,
		1,
		si.Product.Name, si.Product.Vendor, si.Product.Version, si.Product.Serial,
	)

	ch <- prometheus.MustNewConstMetric(
		h.boardInfo,
		prometheus.CounterValue,
		1,
		si.Board.Name, si.Board.Vendor, si.Board.Version, si.Board.Serial, si.Board.Serial,
	)

	ch <- prometheus.MustNewConstMetric(
		h.chassisInfo,
		prometheus.CounterValue,
		1,
		uintToStr(si.Chassis.Type), si.Board.Vendor, si.Board.Version, si.Board.Serial, si.Board.Serial,
	)

	ch <- prometheus.MustNewConstMetric(
		h.biosInfo,
		prometheus.CounterValue,
		1,
		si.BIOS.Vendor, si.BIOS.Version, si.BIOS.Date,
	)

	ch <- prometheus.MustNewConstMetric(
		h.cpuInfo,
		prometheus.CounterValue,
		1,
		si.CPU.Vendor, si.CPU.Model, uintToStr(si.CPU.Speed), uintToStr(si.CPU.Cache), uintToStr(si.CPU.Cpus), uintToStr(si.CPU.Cores), uintToStr(si.CPU.Threads),
	)

	ch <- prometheus.MustNewConstMetric(
		h.memoryInfo,
		prometheus.CounterValue,
		1,
		si.Memory.Type, uintToStr(si.Memory.Speed), uintToStr(si.Memory.Size),
	)

	for _, sta := range si.Storage {
		ch <- prometheus.MustNewConstMetric(
			h.storageInfo,
			prometheus.CounterValue,
			1,
			sta.Name, sta.Driver, sta.Vendor, sta.Model, sta.Serial, uintToStr(sta.Size),
		)
	}

	for _, net := range si.Network {
		ch <- prometheus.MustNewConstMetric(
			h.networkInfo,
			prometheus.CounterValue,
			1,
			net.Name, net.Driver, net.MACAddress, net.Port, uintToStr(net.Speed),
		)
	}

	return nil
}

func uintToStr(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

func init() {
	easyxporter.RegisterCollector(hardwareInfoCollectorSubsystem, true, NewHardwareInfoCollector)
}
