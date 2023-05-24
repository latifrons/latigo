package program

import (
	"go.uber.org/zap"
	"strings"
)

type Component interface {
	Start()
	Stop()
	// Get the component name
	Name() string
}

type ComponentProvider interface {
	ProvideAllComponents() []Component
	ProvideEnabledComponent() map[string]bool
}

type ComponentService struct {
	Logger            *zap.SugaredLogger
	ComponentProvider ComponentProvider
	components        []Component
	componentsEnabled map[string]bool
}

func (n *ComponentService) InitComponents() {
	if n.ComponentProvider == nil {
		n.components = []Component{}
		return
	}
	components := n.ComponentProvider.ProvideAllComponents()
	n.componentsEnabled = n.ComponentProvider.ProvideEnabledComponent()

	for _, component := range components {
		if v, ok := n.componentsEnabled[strings.ToLower(component.Name())]; ok && v {
			n.components = append(n.components, component)
		} else {
			n.Logger.Infow("component disabled", "name", component.Name())
		}
	}
}

func (n *ComponentService) AddComponent(component Component) {
	n.components = append(n.components, component)
}

func (n *ComponentService) Start() {
	for _, component := range n.components {
		n.Logger.Infow("starting component", "name", component.Name())
		component.Start()
		n.Logger.Infow("started component", "name", component.Name())
	}
	n.Logger.Infow("all components started")
}

func (n *ComponentService) Stop() {
	for i := len(n.components) - 1; i >= 0; i-- {
		comp := n.components[i]
		n.Logger.Infow("stopping component", "name", comp.Name())

		comp.Stop()
		n.Logger.Infow("stopped component", "name", comp.Name())
	}
	n.Logger.Infow("all components stopped")
}
