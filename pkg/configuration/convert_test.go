package configuration

import (
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestToDistributionConfigurations(t *testing.T) {
	data, _ := os.ReadFile("../../cmd/registry-proxy-cache/config-dev.yaml")
	c, _ := Parse(bytes.NewBuffer(data))
	_ = yaml.NewEncoder(os.Stdout).Encode(ToDistributionConfigurations(c))
}
