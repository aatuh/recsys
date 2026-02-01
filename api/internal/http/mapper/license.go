package mapper

import (
	"time"

	"github.com/aatuh/recsys-suite/api/internal/license"
	"github.com/aatuh/recsys-suite/api/src/specs/types"
)

// LicenseStatusDTO maps license info to API response.
func LicenseStatusDTO(info license.Info) types.LicenseStatusResponse {
	out := types.LicenseStatusResponse{
		Status:     string(info.Status),
		Commercial: info.Commercial,
	}
	if info.ExpiresAt != nil {
		out.ExpiresAt = info.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if info.Customer != "" {
		out.Customer = info.Customer
	}
	if len(info.Entitlements) > 0 {
		out.Entitlements = cloneIntMap(info.Entitlements)
	}
	return out
}
