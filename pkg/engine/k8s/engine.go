package k8s

import (
	"context"
	textTemplate "text/template"

	"github.com/go-logr/logr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/webdevops/helm-azure-tpl/pkg/engine/base"
)

type (
	K8sEngine struct {
		base.Engine
		*base.BaseEngine

		clients map[string]*KubeClient
	}

	KubeClient struct {
		kubernetes *kubernetes.Clientset
		dynamic    *dynamic.DynamicClient
	}
)

func New(base *base.BaseEngine) *K8sEngine {
	ret := K8sEngine{
		BaseEngine: base,
		clients:    make(map[string]*KubeClient),
	}
	ret.Init()
	return &ret
}

func (e *K8sEngine) Init() {
	// kube logger (with translator)
	logrHandler := logr.NewContextWithSlogLogger(context.Background(), e.Logger.Slog())
	kubeLogger, err := logr.FromContext(logrHandler)
	if err != nil {
		panic(err.Error())
	}
	log.SetLogger(kubeLogger)
}

func (e *K8sEngine) TxtFuncMap(tmpl *textTemplate.Template) textTemplate.FuncMap {
	funcMap := textTemplate.FuncMap{
		"k8sConfigMap": e.k8sConfigmap,
	}

	return funcMap
}
