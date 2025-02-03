package plugins

import (
    "fmt"
    "plugin"
    "sync"
    
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    ctrl "sigs.k8s.io/controller-runtime"
)

type SimplePluginManager struct {
    manager   ctrl.Manager
    plugins   map[string]reconcile.Reconciler
    pluginsMu sync.RWMutex
}

func NewPluginManager(mgr ctrl.Manager) *SimplePluginManager {
    return &SimplePluginManager{
        manager: mgr,
        plugins: make(map[string]reconcile.Reconciler),
    }
}

func (pm *SimplePluginManager) LoadPlugin(path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open plugin: %w", err)
    }

    // Look up the controller constructor
    newControllerSym, err := p.Lookup("NewController")
    if err != nil {
        return fmt.Errorf("NewController not found: %w", err)
    }

    // Get the controller name
    nameSym, err := p.Lookup("ControllerName")
    if err != nil {
        return fmt.Errorf("ControllerName not found: %w", err)
    }
    name := *nameSym.(*string)

    // Create new controller instance
    newController := newControllerSym.(func() (reconcile.Reconciler, error))
    controller, err := newController()
    if err != nil {
        return fmt.Errorf("failed to create controller: %w", err)
    }

    // Store the plugin
    pm.pluginsMu.Lock()
    pm.plugins[name] = controller
    pm.pluginsMu.Unlock()

    return nil
}

