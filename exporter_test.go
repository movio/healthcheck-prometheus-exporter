package main

import (
	"reflect"
	"testing"
)

const sample = `
<health-check>
  <service2 check="pass" connections="45" emptyattr=""/>
</health-check>
`

func TestFindCheckedResultValue(t *testing.T) {
	val, err := findAttribute([]byte(sample), "service2", "check")
	expect(t, err, nil)
	expect(t, val, "pass")
}

func TestFindEmptyAttributeValue(t *testing.T) {
	val, err := findAttribute([]byte(sample), "service2", "emptyattr")
	expect(t, err, nil)
	expect(t, val, "")
}

func TestFindCustomAttributeValue(t *testing.T) {
	val, err := findAttribute([]byte(sample), "service2", "connections")
	expect(t, err, nil)
	expect(t, val, "45")
}

func TestFindCustomAttributeValueNotFound(t *testing.T) {
	val, err := findAttribute([]byte(sample), "service2", "notpresent")
	if err == nil {
		t.Errorf("Expected error - Got %v (type %v)", val, reflect.TypeOf(val))
	}
}
