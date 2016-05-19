package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	hcGauges := []HcGauge{}
	for _, tenant := range config.Tenants {
		log.Printf("Register guage for tenant [%s]\n", tenant)

		hcGuage := newHcGauge(tenant, config)
		hcGauges = append(hcGauges, hcGuage)

		for name, service := range hcGuage.Services {
			log.Printf("  - service: %s\n", name)
			prometheus.MustRegister(service.Gauge)
		}
	}

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for t := range ticker.C {
			log.Println("Updating gauges at", t)
			for _, guage := range hcGauges {
				guage.update()
			}
		}
	}()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":8080", nil)
}
