package metrics

// MetricSpec configures a metric instance.
type MetricSpec struct {
	Name string `yaml:"name"`
	K    int    `yaml:"k,omitempty"`
}
