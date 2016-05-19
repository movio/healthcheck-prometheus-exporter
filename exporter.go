package main

import (
	"net/http"

	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"strings"
)

type HcGauge struct {
	Name     string
	URLs     map[string]string    // Type to URL
	Services map[string]HcService // Service name to Service
}

type HcService struct {
	Name  string
	Type  string
	Gauge prometheus.Gauge
}

func newGauge(tenant string, service ConfigService) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: service.Type,
		Subsystem: tenant,
		Name:      strings.Replace(service.Name, "-", "_", -1),
		Help:      service.Help,
	})
}

func newHcGauge(tenant Tenant, config Config) HcGauge {
	urls := map[string]string{}
	for key, template := range config.Templates {
		urls[key] = fmt.Sprintf(template, tenant.Host)
	}

	services := map[string]HcService{}
	for _, svc := range config.Services {
		services[svc.Name] = HcService{
			svc.Name, svc.Type, newGauge(tenant.Name, svc),
		}
	}

	return HcGauge{tenant.Name, urls, services}
}

func (hc HcGauge) update() {
	for _, url := range hc.URLs {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("Failed to request "+url, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Failed to read response body", err)
		}

		parsed, err := parse(body)
		if err != nil {
			log.Fatal("Failed to parse XML response", err)
		}

		for _, svc := range hc.Services {
			result := parsed.findResult(svc.Name)

			// NOTE: We consider a service failed on health check only if it
			//       returns check="fail" in response, so that we can omit those
			//       services whose returns empty in `check` attribute.
			var value float64
			if result.Check == "fail" {
				value = 1
			} else {
				value = 0
			}

			svc.Gauge.Set(value)
		}
	}
}
