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
	Key   string
	Name  string
	Type  string
	Attr  string
	Gauge prometheus.Gauge
}

func newGauge(key string, tenant string, service ConfigService) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: service.Type,
		Subsystem: tenant,
		Name:      key,
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

		attrSet := map[string]bool{}
		attrSet["check"] = true
		for _, attr := range svc.Attr {
			attrSet[attr] = true
		}

		for attr, _ := range attrSet {
			newSvc := newService(svc, attr, tenant.Name)
			services[newSvc.Key] = newSvc
		}
	}

	return HcGauge{tenant.Name, urls, services}
}

func newService(svc ConfigService, attr string, tenant string) HcService {
	key := strings.Replace(fmt.Sprintf("%s_%s", svc.Name, attr), "-", "_", -1)
	return HcService{key, svc.Name, svc.Type, attr, newGauge(key, tenant, svc)}
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
			log.Println("Failed to request "+url, err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body: ", err)
			return
		}

		parsed, err := parse(body)
		if err != nil {
			log.Println("Failed to parse XML response: ", err)
			return
		}

		for _, svc := range hc.Services {
			if svc.Type != hostType {
				continue
			}

			result := parsed.findResult(svc.Name)
			value, err := getUpdateValue(svc.Name, svc.Attr, result, body)

			if err != nil {
				log.Println("Failed to read updated value: ", err)
				continue
			}

			svc.Gauge.Set(value)
		}
	}
}

func getUpdateValue(svcName string, svcAttr string, result XmlService, body []byte) (float64, error) {
	var value float64 = 0
	switch svcAttr {
	case "check":
		// NOTE: We consider a service failed on health check only if it
		//       returns check="fail" in response, so that we can omit those
		//       services whose returns empty in `check` attribute.
		if result.Check == "fail" {
			value = 1
		}
	default:
		val, err := findAttribute(body, svcName, svcAttr)
		if err != nil {
			return 0, fmt.Errorf("Failed to read attribute: %s", err)
		}

		value, err = strconv.ParseFloat(val, 10)
		if err != nil {
			return 0, fmt.Errorf("Failed to parse attribute to integer: %s", err)
		}
	}
	return value, nil
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
