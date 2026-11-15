package k8s

import (
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// k8sConfigmap fetchtes a configmap based on kubeContext, namespace and configmap name
func (e *K8sEngine) k8sConfigmap(kubeContext, namespace, configmap string) (interface{}, error) {
	e.Logger.Info(`fetching Kubernetes ConfigMap Resource`, slog.String("kubeContext", kubeContext), slog.String("namespace", namespace), slog.String("configmap", configmap))

	cacheKey := e.GenerateCacheKey(`k8sConfigmap`, kubeContext, namespace, configmap)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		client, err := e.KubeClient(kubeContext)
		if err != nil {
			return nil, err
		}

		opts := metav1.GetOptions{}
		res, err := client.CoreV1().ConfigMaps(namespace).Get(e.Ctx, configmap, opts)
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(res)
	})
}
