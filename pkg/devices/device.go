package devices

import "regexp"

// Device : struct representing a generic device
type Device struct {
	Addr       string
	Port       string
	User       string
	Pass1      string
	Pass2      string
	DeviceType string
	PromptRE   *regexp.Regexp
}
