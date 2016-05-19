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

	// Perform initial update
	log.Println("Updating gauges at", time.Now())
	updateAll(hcGauges)

	// Update all guages previously
	ticker := time.Tick(time.Second * 30)
	go func() {
		for now := range ticker {
			log.Println("Updating gauges at", now)
			updateAll(hcGauges)
		}
	}()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":8080", nil)
}
