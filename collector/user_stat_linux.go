package collector

import (
	"github.com/Faione/easyxporter"
	"github.com/Faione/users"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	userStatCollectorSubsystem = "user"
)

// 读取/etc/passwd与/var/run/utmp，获得总用户与当前登陆的用户信息
type userInfoCollector struct {
	existInfo  *prometheus.Desc
	loggedInfo *prometheus.Desc
}

func NewUserInfoCollector(logger *logrus.Logger) (easyxporter.Collector, error) {
	return &userInfoCollector{
		existInfo: prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), userStatCollectorSubsystem, "exist"),
			"Exsit User info from /etc/passwd",
			[]string{"uid", "gid", "username", "name", "homedir"}, nil,
		),
		loggedInfo: prometheus.NewDesc(
			prometheus.BuildFQName(easyxporter.GetNameSpace(), userStatCollectorSubsystem, "logged"),
			"Exsit User info from /var/run/utmp",
			[]string{"username", "tty", "logintime"}, nil,
		),
	}, nil
}

func (u *userInfoCollector) Update(ch chan<- prometheus.Metric) error {
	existUsers, err := users.ListAll()
	if err != nil {
		return err
	}

	for _, user := range existUsers {
		ch <- prometheus.MustNewConstMetric(
			u.existInfo,
			prometheus.CounterValue,
			1,
			user.Uid, user.Gid, user.Username, user.Name, user.HomeDir,
		)
	}

	loggedUsers, err := users.ListLogged()
	if err != nil {
		return err
	}

	for _, user := range loggedUsers {
		ch <- prometheus.MustNewConstMetric(
			u.loggedInfo,
			prometheus.CounterValue,
			1,
			user.Username, user.Tty, user.LoginTime,
		)
	}

	return nil
}

func init() {
	easyxporter.RegisterCollector(userStatCollectorSubsystem, true, NewUserInfoCollector)
}
