package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SecretController struct {
	client client.Client
}

func NewController() (reconcile.Reconciler, error) {
	return &SecretController{}, nil
}

func (c *SecretController) SetClient(client client.Client) {
	c.client = client
}

func (c *SecretController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{}
	err := c.client.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, secret)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Printf("Secret reconciled: %s/%s\n", secret.Namespace, secret.Name)
	return reconcile.Result{}, nil
}

// Export the constructor for the plugin system
var New = NewController
