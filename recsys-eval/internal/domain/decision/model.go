package decision

import "time"

const (
	DecisionShip = "ship"
	DecisionHold = "hold"
	DecisionFail = "fail"
)

// Artifact captures the decision outcome for an experiment.
type Artifact struct {
	Decision         string              `json:"decision"`
	Reason           string              `json:"reason"`
	ControlVariant   string              `json:"control_variant"`
	CandidateVariant string              `json:"candidate_variant"`
	PrimaryMetrics   []MetricDecision    `json:"primary_metrics"`
	Guardrails       []GuardrailDecision `json:"guardrails"`
	Thresholds       map[string]float64  `json:"thresholds"`
	DatasetWindow    Window              `json:"dataset_window"`
	Versions         Versions            `json:"versions"`
	SRM              *SRMDecision        `json:"srm,omitempty"`
}

// MetricDecision captures a primary metric delta.
type MetricDecision struct {
	Metric        string  `json:"metric"`
	Delta         float64 `json:"delta"`
	RelativeDelta float64 `json:"relative_delta"`
	Threshold     float64 `json:"threshold"`
}

// GuardrailDecision captures guardrail evaluation.
type GuardrailDecision struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Passed    bool    `json:"passed"`
}

// Window captures dataset window.
type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Versions captures build metadata.
type Versions struct {
	BinaryVersion string `json:"binary_version"`
	GitCommit     string `json:"git_commit"`
}

// SRMDecision captures SRM gating info.
type SRMDecision struct {
	Detected bool    `json:"detected"`
	PValue   float64 `json:"p_value"`
	Alpha    float64 `json:"alpha"`
}

// ExitCode returns exit status for this decision.
func (a Artifact) ExitCode() int {
	switch a.Decision {
	case DecisionShip:
		return 0
	case DecisionHold:
		return 2
	case DecisionFail:
		return 3
	default:
		return 1
	}
}
