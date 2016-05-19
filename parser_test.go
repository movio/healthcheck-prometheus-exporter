package main

import (
	"reflect"
	"testing"
)

const passed = `
<health-check>
  <service check="pass" db="vc_test"/>
</health-check>
`

const failed = `
<health-check>
	<service1 check="fail" db="vc_test"/>
	<service2 check="pass" connections="45"/>
	<service3 check="fail" stale-campaigns="0"/>
</health-check>
`

const invalidXml = `
<health-check>
  <service check="pass" /
</health-check>
`

func expect(t *testing.T, got interface{}, expected interface{}) {
	if got != expected {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)",
			expected, reflect.TypeOf(expected), got, reflect.TypeOf(got))
	}
}

func TestParseSingleService(t *testing.T) {
	hc, err := parse([]byte(passed))

	expect(t, err, nil)
	expect(t, hc.XMLName.Local, "health-check")
	expect(t, len(hc.ServiceList), 1)
	expect(t, hc.ServiceList[0].XMLName.Local, "service")
	expect(t, hc.ServiceList[0].Check, "pass")
}

func TestParseMultipleService(t *testing.T) {
	hc, err := parse([]byte(failed))

	expect(t, err, nil)
	expect(t, hc.XMLName.Local, "health-check")
	expect(t, len(hc.ServiceList), 3)
	expect(t, hc.ServiceList[0].XMLName.Local, "service1")
	expect(t, hc.ServiceList[0].Check, "fail")
	expect(t, hc.ServiceList[1].XMLName.Local, "service2")
	expect(t, hc.ServiceList[1].Check, "pass")
	expect(t, hc.ServiceList[2].XMLName.Local, "service3")
	expect(t, hc.ServiceList[2].Check, "fail")
}

func TestFailedOnInvalidXML(t *testing.T) {
	_, err := parse([]byte(invalidXml))

	expect(t, err != nil, true)
}
