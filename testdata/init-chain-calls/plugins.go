package main

import "fmt"

// Plugin system with init registration

type Plugin interface {
	Name() string
	Initialize() error
}

var plugins []Plugin

// Concrete plugins that register themselves

type LogPlugin struct{}

func (p *LogPlugin) Name() string {
	return "logger"
}

func (p *LogPlugin) Initialize() error {
	recordInit("LogPlugin.Initialize")
	return nil
}

func init() {
	recordInit("init#LogPlugin")
	// Self-registering plugin
	registerPlugin(&LogPlugin{})
}

type MetricsPlugin struct{}

func (p *MetricsPlugin) Name() string {
	return "metrics"
}

func (p *MetricsPlugin) Initialize() error {
	recordInit("MetricsPlugin.Initialize")
	return nil
}

func init() {
	recordInit("init#MetricsPlugin")
	// Self-registering plugin
	registerPlugin(&MetricsPlugin{})
}

// Plugin registration

func registerPlugin(p Plugin) {
	recordInit("registerPlugin")
	plugins = append(plugins, p)
}

func initializePlugins() error {
	recordInit("initializePlugins")
	for _, p := range plugins {
		if err := p.Initialize(); err != nil {
			return fmt.Errorf("plugin %s failed: %w", p.Name(), err)
		}
	}
	return nil
}

func init() {
	recordInit("init#plugins")
	// Initialize all registered plugins
	if err := initializePlugins(); err != nil {
		panic(fmt.Sprintf("plugin initialization failed: %v", err))
	}
}

// Unused plugin that doesn't register itself

type UnusedPlugin struct{}

func (p *UnusedPlugin) Name() string {
	// Not registered, should be reported as unused
	return "unused"
}

func (p *UnusedPlugin) Initialize() error {
	// Not registered, should be reported as unused
	return nil
}

// Factory pattern with init registration

type HandlerFactory func() Handler

var handlerFactories = make(map[string]HandlerFactory)

func RegisterHandlerFactory(name string, factory HandlerFactory) {
	recordInit("RegisterHandlerFactory:" + name)
	handlerFactories[name] = factory
}

func init() {
	recordInit("init#factories")
	// Register factories
	RegisterHandlerFactory("json", jsonHandlerFactory)
	RegisterHandlerFactory("xml", xmlHandlerFactory)
}

func jsonHandlerFactory() Handler {
	// Used via factory registration
	return func(msg string) error {
		fmt.Printf("JSON: %s\n", msg)
		return nil
	}
}

func xmlHandlerFactory() Handler {
	// Used via factory registration
	return func(msg string) error {
		fmt.Printf("XML: %s\n", msg)
		return nil
	}
}

func unusedHandlerFactory() Handler {
	// Not registered, should be reported as unused
	return func(msg string) error {
		return nil
	}
}
