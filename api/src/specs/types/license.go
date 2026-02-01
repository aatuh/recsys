package types

// LicenseStatusResponse describes the license status payload.
type LicenseStatusResponse struct {
	Status       string         `json:"status"`
	Commercial   bool           `json:"commercial"`
	ExpiresAt    string         `json:"expires_at,omitempty"`
	Customer     string         `json:"customer,omitempty"`
	Entitlements map[string]int `json:"entitlements,omitempty"`
}
