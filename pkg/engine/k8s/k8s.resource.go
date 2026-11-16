package k8s

import (
	"fmt"
	"log/slog"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

// Resource fetches a resource based on current kubernetes context, namespace, groupVersionKind and name
func (e *KubeClient) Resource(gvk, namespace, name string) (interface{}, error) {
	e.Logger.Info(`fetching Kubernetes Resource`, slog.String("namespace", namespace), slog.String("groupVersionKind", gvk), slog.String("name", name))

	gvkDefinition := schema.GroupVersionResource{}
	gvkParts := strings.Split(strings.TrimSpace(gvk), "/")
	switch len(gvkParts) {
	case 2:
		gvkDefinition.Group = "" // core
		gvkDefinition.Version = strings.TrimSpace(gvkParts[0])
		gvkDefinition.Resource = strings.TrimSpace(gvkParts[1])
	case 3:
		gvkDefinition.Group = strings.TrimSpace(gvkParts[0])
		gvkDefinition.Version = strings.TrimSpace(gvkParts[1])
		gvkDefinition.Resource = strings.TrimSpace(gvkParts[2])
	default:
		return nil, fmt.Errorf("invalid GroupVersionKind (group/version/kind or version/kind for core): %s", gvk)
	}

	cacheKey := e.GenerateCacheKey(`Resource`, gvk, namespace, name)
	return e.CacheResult(cacheKey, func() (interface{}, error) {
		res, err := e.dynClient().Resource(gvkDefinition).Namespace(namespace).Get(e.Ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		return e.TransformToInterface(res)
	})
}
