package staging

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/fsutil"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

// Store manages local staging of artifacts produced by compute jobs.
//
// The layout is:
//
//	<base>/<tenant>/<surface>/<segment>/<type>/<start>_<end>/
//	  - <version>.json
//	  - current.version
type Store struct {
	baseDir string
}

func New(baseDir string) Store { return Store{baseDir: baseDir} }

func (s Store) Put(ctx context.Context, ref artifacts.Ref, blob []byte) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if err := ref.Key.Validate(); err != nil {
		return "", err
	}
	if ref.Version == "" {
		return "", fmt.Errorf("ref version must be set")
	}
	winDir := s.windowDir(ref.Key, ref.Window)
	path := filepath.Join(winDir, ref.Version+".json")
	if err := fsutil.WriteFileAtomic(path, blob, 0o644); err != nil {
		return "", err
	}
	cur := filepath.Join(winDir, "current.version")
	if err := fsutil.WriteFileAtomic(cur, []byte(ref.Version+"\n"), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// LoadCurrent loads the artifact blob for the current version for the given
// key+window. If no artifact is staged, (false, nil, nil) is returned.
func (s Store) LoadCurrent(
	ctx context.Context,
	key artifacts.Key,
	w windows.Window,
) (artifacts.Ref, []byte, bool, error) {
	select {
	case <-ctx.Done():
		return artifacts.Ref{}, nil, false, ctx.Err()
	default:
	}
	if err := key.Validate(); err != nil {
		return artifacts.Ref{}, nil, false, err
	}
	winDir := s.windowDir(key, w)
	cur := filepath.Join(winDir, "current.version")
	b, err := os.ReadFile(cur)
	if err != nil {
		if os.IsNotExist(err) {
			// Fallback: if there is a single .json file, pick it.
			ref, blob, ok, err := s.loadSingleJSON(key, w, winDir)
			return ref, blob, ok, err
		}
		return artifacts.Ref{}, nil, false, err
	}
	ver := strings.TrimSpace(string(b))
	if ver == "" {
		return artifacts.Ref{}, nil, false, fmt.Errorf("staging current.version is empty")
	}
	path := filepath.Join(winDir, ver+".json")
	blob, err := os.ReadFile(path)
	if err != nil {
		return artifacts.Ref{}, nil, false, err
	}
	ref := artifacts.Ref{Key: key, Window: w, Version: ver, BuiltAt: time.Time{}}
	return ref, blob, true, nil
}

func (s Store) windowDir(key artifacts.Key, w windows.Window) string {
	start := w.Start.UTC().Format("2006-01-02")
	end := w.End.UTC().Format("2006-01-02")
	win := start + "_" + end
	return filepath.Join(
		s.baseDir,
		key.Tenant,
		key.Surface,
		key.Segment,
		string(key.Type),
		win,
	)
}

func (s Store) loadSingleJSON(
	key artifacts.Key,
	w windows.Window,
	winDir string,
) (artifacts.Ref, []byte, bool, error) {
	ents, err := os.ReadDir(winDir)
	if err != nil {
		if os.IsNotExist(err) {
			return artifacts.Ref{}, nil, false, nil
		}
		return artifacts.Ref{}, nil, false, err
	}
	var files []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			files = append(files, filepath.Join(winDir, name))
		}
	}
	if len(files) == 0 {
		return artifacts.Ref{}, nil, false, nil
	}
	sort.Strings(files)
	path := files[len(files)-1]
	blob, err := os.ReadFile(path)
	if err != nil {
		return artifacts.Ref{}, nil, false, err
	}
	ver := strings.TrimSuffix(filepath.Base(path), ".json")
	ref := artifacts.Ref{Key: key, Window: w, Version: ver, BuiltAt: time.Time{}}
	return ref, blob, true, nil
}

func (s Store) EnsureBaseDir(perm fs.FileMode) error {
	return os.MkdirAll(s.baseDir, perm)
}
