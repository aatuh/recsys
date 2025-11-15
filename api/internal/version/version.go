package version

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Info summarizes build/runtime metadata for the service.
type Info struct {
	GitCommit    string `json:"git_commit"`
	BuildTime    string `json:"build_time"`
	ModelVersion string `json:"model_version"`
}

var (
	gitCommitOnce sync.Once
	gitCommitVal  string

	buildTimeOnce sync.Once
	buildTimeVal  string
)

// GitCommit returns the commit SHA baked into the binary (or "unknown").
func GitCommit() string {
	gitCommitOnce.Do(func() {
		if val := strings.TrimSpace(os.Getenv("RECSYS_GIT_COMMIT")); val != "" {
			gitCommitVal = val
			return
		}
		if commit, err := detectGitCommit(); err == nil && commit != "" {
			gitCommitVal = commit
			return
		}
		gitCommitVal = "unknown"
	})
	return gitCommitVal
}

// BuildTime returns the UTC timestamp supplied via env (or process start time).
func BuildTime() string {
	buildTimeOnce.Do(func() {
		if val := strings.TrimSpace(os.Getenv("RECSYS_BUILD_TIME")); val != "" {
			buildTimeVal = val
			return
		}
		buildTimeVal = time.Now().UTC().Format(time.RFC3339)
	})
	return buildTimeVal
}

// Snapshot returns a stable Info payload with the supplied model version.
func Snapshot(modelVersion string) Info {
	if strings.TrimSpace(modelVersion) == "" {
		modelVersion = "unknown"
	}
	return Info{
		GitCommit:    GitCommit(),
		BuildTime:    BuildTime(),
		ModelVersion: modelVersion,
	}
}

func detectGitCommit() (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output() // #nosec G204 -- need git metadata
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
