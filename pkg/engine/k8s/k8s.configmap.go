package k8s

import (
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMaps fetches all configmaps based on current kubernetes context and namespace
func (e *KubeClient) ConfigMaps(namespace string) (interface{}, error) {
	e.Logger.Info(`fetching Kubernetes ConfigMaps`, slog.String("namespace", namespace))

	cacheKey := e.GenerateCacheKey(`ConfigMaps`, namespace)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		opts := metav1.ListOptions{}
		res, err := e.kubeClient().CoreV1().ConfigMaps(namespace).List(e.Ctx, opts)
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(res)
	})
}

// ConfigMap fetches a configmap based on current kubernetes context, namespace and configmap name
func (e *KubeClient) ConfigMap(namespace, configmap string) (interface{}, error) {
	e.Logger.Info(`fetching Kubernetes ConfigMap`, slog.String("namespace", namespace), slog.String("configmap", configmap))

	cacheKey := e.GenerateCacheKey(`ConfigMap`, namespace, configmap)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		opts := metav1.GetOptions{}
		res, err := e.kubeClient().CoreV1().ConfigMaps(namespace).Get(e.Ctx, configmap, opts)
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(res)
	})
}
