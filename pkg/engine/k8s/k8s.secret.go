package k8s

import (
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret fetches a secret based on current kubernetes context, namespace and secret name
func (e *KubeClient) Secret(namespace, secret string) (interface{}, error) {
	e.Logger.Info(`fetching Kubernetes Secret Resource`, slog.String("namespace", namespace), slog.String("secret", secret))

	cacheKey := e.GenerateCacheKey(`Secret`, namespace, secret)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		opts := metav1.GetOptions{}
		res, err := e.kubeClient().CoreV1().Secrets(namespace).Get(e.Ctx, secret, opts)
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(res)
	})
}
