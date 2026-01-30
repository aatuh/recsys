package recsysvc

// ResponseMeta captures algorithm/config metadata for a response.
type ResponseMeta struct {
	AlgoVersion   string
	ConfigVersion string
	RulesVersion  string
}
