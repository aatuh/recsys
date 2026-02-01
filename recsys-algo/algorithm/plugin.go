package algorithm

import (
	"errors"
	"fmt"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"
)

// PluginSymbol is the exported symbol name expected from Go plugins.
const PluginSymbol = "RecsysAlgorithmPlugin"

// Plugin describes a custom algorithm plugin entrypoint.
type Plugin struct {
	ContractVersion string
	New             func(store recmodel.EngineStore, rulesManager *rules.Manager, cfg Config) (Algorithm, error)
}

// Validate ensures the plugin contract is compatible.
func (p Plugin) Validate() error {
	if p.ContractVersion != ContractVersion {
		return fmt.Errorf("plugin contract version mismatch: %s", p.ContractVersion)
	}
	if p.New == nil {
		return errors.New("plugin factory missing")
	}
	return nil
}
