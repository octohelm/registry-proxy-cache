package storage

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/octohelm/registry-proxy-cache/pkg/kubernetes"

	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
)

func TestClusterGarbageCollect(t *testing.T) {
	data, _ := os.ReadFile("../../cmd/registry-proxy-cache/config-dev.yaml")
	c, _ := configuration.Parse(bytes.NewBuffer(data))

	ctx := context.Background()

	images, err := kubernetes.GetClusterContainerImages(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := ClusterGarbageCollect(ctx, configuration.ToDistributionConfigurations(c), images, false); err != nil {
		t.Fatal(err)
	}
}
