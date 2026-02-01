package jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

const maxLineSize = 16 * 1024 * 1024

// ExposureReader reads exposures from a JSONL file.
type ExposureReader struct{ path string }

func NewExposureReader(path string) ExposureReader { return ExposureReader{path: path} }

func (r ExposureReader) Read(_ context.Context) ([]dataset.Exposure, error) {
	return readJSONL[dataset.Exposure](r.path)
}

func (r ExposureReader) Stream(ctx context.Context) (<-chan dataset.Exposure, <-chan error) {
	return streamJSONL[dataset.Exposure](ctx, r.path)
}

// OutcomeReader reads outcomes from a JSONL file.
type OutcomeReader struct{ path string }

func NewOutcomeReader(path string) OutcomeReader { return OutcomeReader{path: path} }

func (r OutcomeReader) Read(_ context.Context) ([]dataset.Outcome, error) {
	return readJSONL[dataset.Outcome](r.path)
}

func (r OutcomeReader) Stream(ctx context.Context) (<-chan dataset.Outcome, <-chan error) {
	return streamJSONL[dataset.Outcome](ctx, r.path)
}

// AssignmentReader reads assignments from a JSONL file.
type AssignmentReader struct{ path string }

func NewAssignmentReader(path string) AssignmentReader { return AssignmentReader{path: path} }

func (r AssignmentReader) Read(_ context.Context) ([]dataset.Assignment, error) {
	return readJSONL[dataset.Assignment](r.path)
}

// RankListReader reads rank lists from a JSONL file.
type RankListReader struct{ path string }

func NewRankListReader(path string) RankListReader { return RankListReader{path: path} }

func (r RankListReader) Read(_ context.Context) ([]dataset.RankList, error) {
	return readJSONL[dataset.RankList](r.path)
}

func streamJSONL[T any](ctx context.Context, path string) (<-chan T, <-chan error) {
	out := make(chan T)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		// #nosec G304 -- input path provided by operator
		file, err := os.Open(path)
		if err != nil {
			errCh <- err
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, maxLineSize)

		line := 0
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			line++
			data := scanner.Bytes()
			if len(data) == 0 {
				continue
			}
			var item T
			if err := json.Unmarshal(data, &item); err != nil {
				errCh <- fmt.Errorf("jsonl parse error at line %d: %w", line, err)
				return
			}
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case out <- item:
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
			return
		}
	}()

	return out, errCh
}

func readJSONL[T any](path string) ([]T, error) {
	// #nosec G304 -- input path provided by operator
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxLineSize)

	var out []T
	line := 0
	for scanner.Scan() {
		line++
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("jsonl parse error at line %d: %w", line, err)
		}
		out = append(out, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
