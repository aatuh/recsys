//go:build !duckdb

package usecase

func duckdbSupported() bool { return false }
