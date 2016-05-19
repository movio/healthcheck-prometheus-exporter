package main

import (
	"net/http"

	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
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
	Vkey  string
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
			svc.Name, svc.Type, svc.Vkey, newGauge(tenant.Name, svc),
		}
	}

	return HcGauge{tenant.Name, urls, services}
}

func updateAll(hcs []HcGauge) {
	for _, guage := range hcs {
		guage.update()
	}
}

func (hc HcGauge) update() {
	for hostType, url := range hc.URLs {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("Failed to request "+url, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Failed to read response body: ", err)
		}

		parsed, err := parse(body)
		if err != nil {
			log.Fatal("Failed to parse XML response: ", err)
		}

		for _, svc := range hc.Services {
			if svc.Type != hostType {
				continue
			}

			result := parsed.findResult(svc.Name)

			var value float64 = 0
			if svc.Vkey == "" {
				// NOTE: We consider a service failed on health check only if it
				//       returns check="fail" in response, so that we can omit those
				//       services whose returns empty in `check` attribute.
				if result.Check == "fail" {
					value = 1
				}
			} else {
				val, err := findAttribute(body, svc.Name, svc.Vkey)
				if err != nil {
					log.Fatal("Failed to read attribute: ", err)
				}

				value, err = strconv.ParseFloat(val, 10)
				if err != nil {
					log.Fatal("Failed to parse attribute to integer: ", err)
				}
			}

			svc.Gauge.Set(value)
		}
	}
}

func findAttribute(source []byte, svcName string, attrName string) (string, error) {
	r := regexp.MustCompile(fmt.Sprintf(`%s .* %s="([0-9]+(\\.[0-9]+)?)"`, svcName, attrName))
	m := r.FindStringSubmatch(string(source[:]))

	if len(m) <= 0 {
		errorMsg := fmt.Sprintf(
			"No attribute [%s] found for [%s] from %s\n", attrName, svcName, source)
		return "", errors.New(errorMsg)
	}

	return m[1], nil
}
