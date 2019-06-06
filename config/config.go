package config

import (
	"fmt"
	"os"
	"sync"

	yaml "gopkg.in/yaml.v3"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	configReloadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "onlineregistration_exporter",
		Name:      "config_last_reload_successful",
		Help:      "Online-registration exporter config loaded successfully.",
	})

	configReloadSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "onlineregistration_exporter",
		Name:      "config_last_reload_success_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
)

func init() {
	prometheus.MustRegister(configReloadSuccess)
	prometheus.MustRegister(configReloadSeconds)
}

type API struct {
	URL   string `yaml:"url"`
	User  string `yaml:"user"`
	Token string `yaml:"token"`
}

type Config struct {
	API   API      `yaml:"api"`
	Plans []string `yaml:"plans"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

func (sc *SafeConfig) ReloadConfig(confFile string) (err error) {
	var c = &Config{}
	defer func() {
		if err != nil {
			configReloadSuccess.Set(0)
		} else {
			configReloadSuccess.Set(1)
			configReloadSeconds.SetToCurrentTime()
		}
	}()

	yamlReader, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}
