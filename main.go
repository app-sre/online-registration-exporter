package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/app-sre/online-registration-exporter/config"
	"github.com/app-sre/online-registration-exporter/onlinereg"
)

var (
	sc = &config.SafeConfig{
		C: &config.Config{},
	}

	configFile    = kingpin.Flag("config.file", "Online-registration exporter configuration file.").Default("config.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9115").String()
	configCheck   = kingpin.Flag("config.check", "If true validate the config file and then exit.").Default().Bool()

	isHiddenGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "onlinereg_hidden",
		Help: "Indicates if a plan is hidden",
	}, []string{"plan"})
	planSubscriberLimitVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "onlinereg_subscriber_limit",
		Help: "Subscriber limit for a plan",
	}, []string{"plan"})
	planCapacityConsumedVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "onlinereg_capacity_consumed",
		Help: "Capacity consumed for a plan",
	}, []string{"plan"})
	planCapacityRemainingVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "onlinereg_capacity_remaining",
		Help: "Capacity remaining for a plan",
	}, []string{"plan"})

	exporterRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "onlinereg_exporter_requests",
		Help: "Number of requests made to the exporter",
	})
	exporterErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "onlinereg_exporter_errors",
		Help: "Number of errors getting plan capacity",
	})
)

func metricsHandler(w http.ResponseWriter, r *http.Request, c *config.Config, logger log.Logger) {
	if len(c.Plans) == 0 {
		return
	}

	orconfig := onlinereg.Config{
		APIUrl:   c.API.URL,
		APIUser:  c.API.User,
		APIToken: c.API.Token,
	}
	orclient := onlinereg.NewClient(orconfig)

	registry := prometheus.NewRegistry()
	registry.MustRegister(isHiddenGaugeVec)
	registry.MustRegister(planSubscriberLimitVec)
	registry.MustRegister(planCapacityConsumedVec)
	registry.MustRegister(planCapacityRemainingVec)
	registry.MustRegister(exporterRequests)
	registry.MustRegister(exporterErrors)

	for _, p := range c.Plans {
		exporterRequests.Inc()
		c, err := orclient.GetPlanCapacity(p)
		if err != nil {
			exporterErrors.Inc()
			continue
		}
		if c.Plan.IsHidden {
			isHiddenGaugeVec.WithLabelValues(p).Set(1)
		} else {
			isHiddenGaugeVec.WithLabelValues(p).Set(0)
		}
		planSubscriberLimitVec.WithLabelValues(p).Set(float64(c.Plan.SubscriberLimit))
		planCapacityConsumedVec.WithLabelValues(p).Set(float64(c.Plan.CapacityConsumed))
		planCapacityRemainingVec.WithLabelValues(p).Set(float64(c.Plan.CapacityRemaining))
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	os.Exit(run())
}

func run() int {
	allowedLevel := promlog.AllowedLevel{}
	promConfig := promlog.Config{
		Level: &allowedLevel,
	}
	flag.AddFlags(kingpin.CommandLine, &promConfig)
	kingpin.Version(version.Print("online-registration-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(&promConfig)

	level.Info(logger).Log("msg", "Starting online-registration exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", version.BuildContext())

	if err := sc.ReloadConfig(*configFile); err != nil {
		level.Error(logger).Log("msg", "Error loading config", "err", err)
		return 1
	}

	if *configCheck {
		level.Info(logger).Log("msg", "Config file is ok exiting...")
		return 0
	}

	level.Info(logger).Log("msg", "Loaded config file")

	hup := make(chan os.Signal, 1)
	reloadCh := make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
					continue
				}
				level.Info(logger).Log("msg", "Reloaded config file")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
					rc <- err
				} else {
					level.Info(logger).Log("msg", "Reloaded config file")
					rc <- nil
				}
			}
		}
	}()

	http.HandleFunc("/-/reload",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "This endpoint requires a POST request.\n")
				return
			}

			rc := make(chan error)
			reloadCh <- rc
			if err := <-rc; err != nil {
				http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
			}
		})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		metricsHandler(w, r, conf, logger)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
    <head><title>Online-registration Exporter</title></head>
    <body>
    <h1>Online-registration Exporter</h1>
    <p><a href="/metrics">Metrics</a></p>
	</body>
    </html>`))
	})

	srv := http.Server{Addr: *listenAddress}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			return 0
		case <-srvc:
			return 1
		}
	}
}
