# k8s-plugin-operator

A Kubernetes Operator that supports dynamic loading of controller plugins at runtime. This operator allows you to extend its functionality by adding new controllers as plugins without rebuilding the operator itself.

## Features

- Dynamic plugin loading at runtime
- Support for multiple controller plugins
- Built-in ConfigMap and Secret controllers as example plugins
- Simple plugin architecture for easy extension
- Kubernetes-native deployment support

## Prerequisites

- Go 1.20 or later
- Access to a Kubernetes cluster
- kubectl configured to communicate with your cluster
- Docker with buildx support for multi-platform builds

## Project Structure

```
.
├── configmap/       # ConfigMap controller plugin
├── deploy/          # Kubernetes deployment manifests
├── plugins/         # Compiled plugin (.so) files directory
├── secret/          # Secret controller plugin
├── Dockerfile       # Container image build file
├── Makefile         # Build and deployment commands
├── go.mod           # Go module definition
└── main.go          # Operator entry point
```

## Building

### Local Build

Build the operator and plugins for Linux ARM64:

```bash
# Build operator binary
make build

# Build plugins
make plugins

# Build both operator and plugins
make all

# Clean build artifacts
make clean
```

### Docker Build and Deploy

The project uses Google Cloud Artifact Registry. Set these environment variables or override them when running make commands:

```bash
REGISTRY=europe-west1-docker.pkg.dev/aura-docker-images/aura
IMAGE_NAME=plugin-operator
TAG=latest
```

Available commands:

```bash
# Build Docker image (linux/amd64)
make docker-build

# Push to registry
make docker-push

# Deploy to Kubernetes
make deploy

# Build, push and deploy in one command
make deploy-all
```

## Creating New Plugins

1. Create a new controller that implements the required interface:

```go
package main

import (
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type YourController struct {}

func NewController() (reconcile.Reconciler, error) {
    return &YourController{}, nil
}

func (c *YourController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    // Your reconciliation logic here
    return reconcile.Result{}, nil
}
```

2. Build your plugin:
```bash
GOOS=linux GOARCH=arm64 go build -buildmode=plugin -o plugins/your-plugin.so your-plugin.go
```

## Configuration

The operator accepts the following command-line flags:

- `--plugins-dir`: Directory containing plugin .so files (default: "./plugins")
- `--metrics-addr`: The address the metric endpoint binds to (default: ":8080")
- `--health-probe-bind-address`: The address the probe endpoint binds to (default: ":8081")
- `--enable-leader-election`: Enable leader election for controller manager (default: false)

## Development

### Requirements

- Go 1.20+
- Kubernetes cluster
- kubectl
- Docker with buildx support
- Access to Google Cloud Artifact Registry (europe-west1-docker.pkg.dev)

### Building for Different Platforms

The Makefile is configured to build for Linux ARM64 by default. The Docker image is built for linux/amd64 using buildx.

### Deployment

The deployment process uses a template approach with the following substitutions:
- `IMAGE_PLACEHOLDER`: Replaced with the full image path
- `{{ TIMESTAMP }}`: Replaced with the current Unix timestamp

To deploy with custom settings:
```bash
make deploy REGISTRY=your-registry IMAGE_NAME=your-image TAG=your-tag
```

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
