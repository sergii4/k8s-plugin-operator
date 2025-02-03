package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	scheme = runtime.NewScheme()
	log    = ctrl.Log.WithName("operator")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

type PluginManager struct {
	manager ctrl.Manager
}

func NewPluginManager(mgr ctrl.Manager) *PluginManager {
	return &PluginManager{
		manager: mgr,
	}
}

func (pm *PluginManager) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	newControllerSym, err := p.Lookup("New")
	if err != nil {
		return fmt.Errorf("new not found: %w", err)
	}

	newControllerPtr := newControllerSym.(*func() (reconcile.Reconciler, error))
	newController := *newControllerPtr
	controller, err := newController()
	if err != nil {
		return fmt.Errorf("failed to create controller: %w", err)
	}

	// Get the client from manager and inject it into the controller if it needs it
	if clientAware, ok := controller.(interface{ SetClient(client.Client) }); ok {
		clientAware.SetClient(pm.manager.GetClient())
	}

	// Setup the controller with manager
	err = ctrl.NewControllerManagedBy(pm.manager).
		For(&corev1.ConfigMap{}). // This would ideally be determined by the plugin
		Complete(controller)
	if err != nil {
		return fmt.Errorf("failed to setup controller: %w", err)
	}

	return nil
}

func main() {
	var (
		metricsAddr          string
		probeAddr            string
		pluginsDir           string
		enableLeaderElection bool
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	flag.StringVar(&pluginsDir, "plugins-dir", "./plugins", "Directory containing controller plugins")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "plugin-operator-leader-election",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create plugin manager
	pluginMgr := NewPluginManager(mgr)

	// Load plugins from directory
	err = filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".so" {
			log.Info("loading plugin", "path", path)
			if err := pluginMgr.LoadPlugin(path); err != nil {
				log.Error(err, "failed to load plugin", "path", path)
				return nil // continue with other plugins
			}
			log.Info("successfully loaded plugin", "path", path)
		}
		return nil
	})
	if err != nil {
		log.Error(err, "error walking plugins directory")
		os.Exit(1)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
