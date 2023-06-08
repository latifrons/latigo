package program

import (
	"github.com/rs/zerolog/log"
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
	ProvideDisabledComponents() map[string]bool
}

type ComponentService struct {
	ComponentProvider  ComponentProvider
	components         []Component
	componentsDisabled map[string]bool
}

func (n *ComponentService) InitComponents() {
	if n.ComponentProvider == nil {
		n.components = []Component{}
		return
	}
	components := n.ComponentProvider.ProvideAllComponents()
	n.componentsDisabled = n.ComponentProvider.ProvideDisabledComponents()

	for _, component := range components {
		if _, ok := n.componentsDisabled[strings.ToLower(component.Name())]; ok {
			log.Info().Str("name", component.Name()).Msg("component disabled")
		} else {
			n.components = append(n.components, component)
		}
	}
}

func (n *ComponentService) AddComponent(component Component) {
	n.components = append(n.components, component)
}

func (n *ComponentService) Start() {
	for _, component := range n.components {
		log.Info().Str("name", component.Name()).Msg("starting component")
		component.Start()
		log.Info().Str("name", component.Name()).Msg("started component")
	}
	log.Info().Msg("all components started")
}

func (n *ComponentService) Stop() {
	for i := len(n.components) - 1; i >= 0; i-- {
		comp := n.components[i]
		log.Info().Str("name", comp.Name()).Msg("stopping component")

		comp.Stop()
		log.Info().Str("name", comp.Name()).Msg("stopped component")
	}
	log.Info().Msg("all components stopped")
}
