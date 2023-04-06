package main

import (
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/spf13/cobra"
	"net/http"
	"nginx-exporter/pkg/exporter"
	"os"
)

func main() {
	var address string
	logger := promlog.New(&promlog.Config{})
	rootCmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			addr, err := cmd.Flags().GetString("address")
			if err != nil {
				level.Error(logger).Log("msg", "Error address", "err", err)
				os.Exit(1)
			}
			reg := prometheus.NewRegistry()
			reg.MustRegister(exporter.New(addr, logger))
			http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
			if err := http.ListenAndServe(":2113", nil); err != nil {
				level.Error(logger).Log("msg", "running http", "err", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&address, "address", "d", "", "nginx-status address")
	rootCmd.MarkFlagsRequiredTogether("address")
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
