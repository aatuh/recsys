package main

import (
	"github.com/aatuh/recsys-suite/recsys-eval/internal/app/usecase"
)

func buildReportMetadata(dataset usecase.DatasetConfig, eval usecase.EvalConfig, mode string) (usecase.ReportMetadata, error) {
	return usecase.BuildReportMetadata(dataset, eval, mode)
}

// ExitError allows cobra to return specific exit codes.
type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "exit"
}

func (e ExitError) ExitCode() int {
	if e.Code == 0 {
		return 1
	}
	return e.Code
}
