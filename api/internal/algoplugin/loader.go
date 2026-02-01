package algoplugin

import (
	"fmt"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/aatuh/recsys-suite/api/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"
)

// Load opens a Go plugin and constructs a custom algorithm implementation.
func Load(path string, store recmodel.EngineStore, rulesManager *rules.Manager, cfg algorithm.Config) (algorithm.Algorithm, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("plugin path is required")
	}
	abs := path
	if !filepath.IsAbs(path) {
		abs = filepath.Clean(path)
	}
	p, err := plugin.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}
	symbol, err := p.Lookup(algorithm.PluginSymbol)
	if err != nil {
		return nil, fmt.Errorf("lookup plugin symbol %s: %w", algorithm.PluginSymbol, err)
	}

	var plug *algorithm.Plugin
	switch typed := symbol.(type) {
	case *algorithm.Plugin:
		plug = typed
	case algorithm.Plugin:
		plug = &typed
	default:
		return nil, fmt.Errorf("plugin symbol %s has incompatible type", algorithm.PluginSymbol)
	}
	if err := plug.Validate(); err != nil {
		return nil, err
	}
	algo, err := plug.New(store, rulesManager, cfg)
	if err != nil {
		return nil, err
	}
	if algo == nil {
		return nil, fmt.Errorf("plugin returned nil algorithm")
	}
	return algo, nil
}
