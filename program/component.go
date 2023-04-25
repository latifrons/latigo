package program

import (
	"github.com/sirupsen/logrus"
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
			logrus.WithField("name", component.Name).Info("component disabled")
		}
	}
}

func (n *ComponentService) AddComponent(component Component) {
	n.components = append(n.components, component)
}

func (n *ComponentService) Start() {
	for _, component := range n.components {
		logrus.Infof("Starting %s", component.Name())
		component.Start()
		logrus.Infof("Started: %s", component.Name())
	}
	logrus.Info("All components started")
}

func (n *ComponentService) Stop() {
	for i := len(n.components) - 1; i >= 0; i-- {
		comp := n.components[i]
		logrus.Infof("Stopping %s", comp.Name())
		comp.Stop()
		logrus.Infof("Stopped: %s", comp.Name())
	}
	logrus.Info("All components stopped")
}
