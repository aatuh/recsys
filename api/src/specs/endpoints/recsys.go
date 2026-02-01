package endpoints

// Recsys service endpoints (v1).
const (
	RecsysBase        = "/v1"
	Recommend         = RecsysBase + "/recommend"
	RecommendValidate = Recommend + "/validate"
	Similar           = RecsysBase + "/similar"
	LicenseStatus     = RecsysBase + "/license"
)
