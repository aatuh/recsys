package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type Backfill struct {
	maxDays int
}

func NewBackfill(maxDays int) *Backfill {
	return &Backfill{maxDays: maxDays}
}

type BackfillFn func(ctx context.Context, w windows.Window) error

func (b *Backfill) Execute(ctx context.Context, startDay, endDay time.Time, fn BackfillFn) error {
	startDay = time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.UTC)
	endDay = time.Date(endDay.Year(), endDay.Month(), endDay.Day(), 0, 0, 0, 0, time.UTC)
	if endDay.Before(startDay) {
		return fmt.Errorf("endDay must be >= startDay")
	}
	days := 0
	for day := startDay; !day.After(endDay); day = day.Add(24 * time.Hour) {
		days++
		if b.maxDays > 0 && days > b.maxDays {
			return fmt.Errorf("backfill exceeds max days: %d > %d", days, b.maxDays)
		}
		if err := fn(ctx, windows.DayWindowUTC(day)); err != nil {
			return err
		}
	}
	return nil
}
