package main

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type HealthCheck struct {
	XMLName     xml.Name     `xml:"health-check"`
	ServiceList []XmlService `xml:",any"`
}

type XmlService struct {
	XMLName xml.Name `xml:""`
	Check   string   `xml:"check,attr"`
}

func (s XmlService) String() string {
	return fmt.Sprintf(`{ %s: %s }`, s.XMLName.Local, s.Check)
}

func (hc HealthCheck) findResult(svcName string) XmlService {
	for _, service := range hc.ServiceList {
		if service.XMLName.Local == svcName {
			return service
		}
	}
	return XmlService{}
}

func parse(source []byte) (HealthCheck, error) {
	hc := HealthCheck{}
	err := xml.Unmarshal(source, &hc)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to parse XML [%s], caused by %v\n", source, err)
		return hc, errors.New(errorMsg)
	}
	return hc, nil
}
