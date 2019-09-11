package grafana

import (
	"github.com/grafana/grafana-plugin-model/go/datasource"
	plugin "github.com/hashicorp/go-plugin"
)

type Plugin interface {
	ID() string
}

type Grafana struct {
	plugins []DatasourcePlugin
}

func New() *Grafana {
	return &Grafana{
		plugins: make([]DatasourcePlugin, 0),
	}
}

func (g *Grafana) Register(p DatasourcePlugin) error {
	g.plugins = append(g.plugins, p)
	return nil
}

func (g *Grafana) Run() error {
	plugins := make(map[string]plugin.Plugin)

	for _, p := range g.plugins {
		plugins[p.ID()] = &datasource.DatasourcePluginImpl{
			Plugin: &datasourcePlugin{
				plugin: p,
			},
		}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: plugins,
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})

	return nil
}
