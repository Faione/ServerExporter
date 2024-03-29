package cmd

import (
	"net/http"
	"strings"

	"github.com/Faione/ServerExporter/collector"
	"github.com/Faione/easyxporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	keyWebListenAddress = "web.listen-address"
	keyWebTelemetryPath = "web.telemetry-path"
	keyWebMaxRequests   = "web.max-requests"

	keyLogLevel = "log"
)

func New() *cobra.Command {
	vp := newViper()
	rootCmd := &cobra.Command{
		Use:   "server_exporter ",
		Short: "Server Exporter",
		Long:  "Server Exporter",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runServerExporter(vp, args); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
	}

	flags := rootCmd.Flags()
	flags.String(
		keyWebListenAddress,
		":9900",
		"Address on which to expose metrics and web interface.",
	)
	flags.String(
		keyWebTelemetryPath,
		"/metrics",
		"Path under which to expose metrics.",
	)
	flags.Int(
		keyWebMaxRequests,
		40,
		"Maximum number of parallel scrape requests. Use 0 to disable.",
	)

	flags.AddFlagSet(easyxporter.Flags())
	vp.BindPFlags(flags)
	return rootCmd
}

func newViper() *viper.Viper {
	vp := viper.New()
	vp.SetEnvPrefix("se")
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.AutomaticEnv()
	return vp
}

func newLogger(level string) *logrus.Logger {
	logger := logrus.New()
	switch level {
	case "ERROR":
		logger.SetLevel(logrus.ErrorLevel)
	case "WARN":
		logger.SetLevel(logrus.WarnLevel)
	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)
	case "TRACE":
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

func runServerExporter(vp *viper.Viper, args []string) error {

	var (
		listenAddress = vp.GetString(keyWebListenAddress)
		metricsPath   = vp.GetString(keyWebTelemetryPath)
		maxRequests   = vp.GetInt(keyWebMaxRequests)
		logLevel      = vp.GetString(keyLogLevel)
	)

	logger := newLogger(logLevel)
	logger.Debug("msg", "init server", "metricsPath: ", metricsPath, " listenAddress: ", listenAddress)

	return easyxporter.Build(
		listenAddress, collector.RootNamespace,
		easyxporter.WithLogger(logger),
		easyxporter.WithMaxRequests(maxRequests),
		easyxporter.WithMetricPath(metricsPath),
	).Run()

}
