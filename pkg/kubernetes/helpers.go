package kubernetes

import (
	"context"
	"strings"

	"github.com/octohelm/registry-proxy-cache/pkg/container"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetClusterContainerImages(ctx context.Context) (map[string]map[string]*container.ContainerImage, error) {
	c, err := NewCoreClientForContext("")
	if err != nil {
		return nil, err
	}

	list, err := c.Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	images := map[string]map[string]*container.ContainerImage{}

	for _, n := range list.Items {
		for _, i := range n.Status.Images {
			for _, image := range i.Names {
				if strings.Contains(image, "@sha256") {
					continue
				}
				ci := container.ParseContainerImage(image)
				if ci == nil {
					continue
				}
				if images[ci.Hub] == nil {
					images[ci.Hub] = map[string]*container.ContainerImage{}
				}

				images[ci.Hub][ci.Ref()] = ci
			}
		}
	}

	return images, nil
}
