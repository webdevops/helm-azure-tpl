package k8s

import (
	"fmt"
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (e *K8sEngine) kubeConfigByContext(kubeContext string) (*rest.Config, error) {
	// LoadingRules handles explicit file paths, KUBECONFIG env vars, or default ~/.kube/config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// Override the active context for this specific client instance
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubeContext,
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load REST config for context %q: %w", kubeContext, err)
	}

	return restConfig, nil
}

func (e *K8sEngine) setupKubeClients(kubeContext string) error {
	kubeContext = strings.TrimSpace(kubeContext)

	if _, ok := e.clients[kubeContext]; !ok {
		e.clients[kubeContext] = &KubeClient{}
	}

	clientConfig, err := e.kubeConfigByContext(kubeContext)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset for context %q: %w", kubeContext, err)
	}
	e.clients[kubeContext].kubernetes = kubeClient

	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset for context %q: %w", kubeContext, err)
	}
	e.clients[kubeContext].dynamic = dynamicClient

	return nil
}

func (e *K8sEngine) KubeClient(kubeContext string) (*kubernetes.Clientset, error) {
	kubeContext = strings.TrimSpace(kubeContext)

	if _, ok := e.clients[kubeContext]; !ok {
		e.clients[kubeContext] = &KubeClient{}
	}

	if clients, ok := e.clients[kubeContext]; ok && clients.kubernetes != nil {
		return clients.kubernetes, nil
	}

	err := e.setupKubeClients(kubeContext)
	if err != nil {
		return nil, err
	}

	return e.clients[kubeContext].kubernetes, nil
}

func (e *K8sEngine) DynamicClient(kubeContext string) (*dynamic.DynamicClient, error) {
	kubeContext = strings.TrimSpace(kubeContext)

	if _, ok := e.clients[kubeContext]; !ok {
		e.clients[kubeContext] = &KubeClient{}
	}

	if clients, ok := e.clients[kubeContext]; ok && clients.dynamic != nil {
		return clients.dynamic, nil
	}

	err := e.setupKubeClients(kubeContext)
	if err != nil {
		return nil, err
	}

	return e.clients[kubeContext].dynamic, nil
}
