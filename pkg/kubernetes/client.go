package kubernetes

import (
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewCoreClientForContext(contextName string) (corev1.CoreV1Interface, error) {
	conf, err := ResolveKubeConfig(contextName)
	if err != nil {
		return nil, err
	}
	return corev1.NewForConfig(conf)
}
