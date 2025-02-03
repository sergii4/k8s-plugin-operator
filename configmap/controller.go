package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var ControllerName = "configmap-controller"

type ConfigMapController struct {
	client client.Client
}

func NewController() (reconcile.Reconciler, error) {
	return &ConfigMapController{}, nil
}

func (c *ConfigMapController) SetClient(client client.Client) {
	c.client = client
}

func (c *ConfigMapController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	configMap := &corev1.ConfigMap{}
	err := c.client.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, configMap)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Printf("ConfigMap reconciled: %s/%s\n", configMap.Namespace, configMap.Name)
	return reconcile.Result{}, nil
}

// Export the constructor for the plugin system
var New = NewController
