package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/interleaving"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/statistics"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/clock"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

// InterleavingUsecase orchestrates interleaving analysis.
type InterleavingUsecase struct {
	RankerA  datasource.RankListReader
	RankerB  datasource.RankListReader
	Outcomes datasource.OutcomeReader
	Reporter reporting.Writer
	Clock    clock.Clock
	Logger   logger.Logger
	Metadata ReportMetadata
}

func (u InterleavingUsecase) Run(ctx context.Context, cfg InterleavingConfig, outputPath string) (report.Report, error) {
	if u.RankerA == nil || u.RankerB == nil || u.Outcomes == nil {
		return report.Report{}, errors.New("ranker_a, ranker_b, and outcomes readers are required")
	}

	listsA, err := u.RankerA.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	listsB, err := u.RankerB.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}
	outcomes, err := u.Outcomes.Read(ctx)
	if err != nil {
		return report.Report{}, err
	}

	mapA := map[string]dataset.RankList{}
	for _, l := range listsA {
		mapA[l.RequestID] = l
	}
	mapB := map[string]dataset.RankList{}
	for _, l := range listsB {
		mapB[l.RequestID] = l
	}

	outByReq := map[string][]dataset.Outcome{}
	for _, o := range outcomes {
		outByReq[o.RequestID] = append(outByReq[o.RequestID], o)
	}

	// #nosec G404 -- deterministic RNG for reproducible interleaving
	rng := rand.New(rand.NewSource(cfg.Seed))
	algo := strings.ToLower(cfg.Algorithm)
	if algo == "" {
		algo = "team_draft"
	}

	winsA := 0
	winsB := 0
	ties := 0
	requests := 0

	reqIDs := make([]string, 0, len(mapA))
	for reqID := range mapA {
		if _, ok := mapB[reqID]; ok {
			reqIDs = append(reqIDs, reqID)
		}
	}
	sort.Strings(reqIDs)

	for _, reqID := range reqIDs {
		listA := mapA[reqID]
		listB := mapB[reqID]
		requests++

		var res interleaving.Result
		switch algo {
		case "balanced":
			res = interleaving.BalancedInterleaving(listA.Items, listB.Items, cfg.MaxResults, rng)
		case "optimized":
			res = interleaving.OptimizedInterleaving(listA.Items, listB.Items, cfg.MaxResults, rng)
		default:
			res = interleaving.TeamDraft(listA.Items, listB.Items, cfg.MaxResults, rng)
		}

		wins := countWins(res, outByReq[reqID])
		if wins.A > wins.B {
			winsA++
		} else if wins.B > wins.A {
			winsB++
		} else {
			ties++
		}
	}

	pValue := 1.0
	if winsA+winsB > 0 {
		z := (float64(winsA) - float64(winsA+winsB)*0.5) / math.Sqrt(float64(winsA+winsB)*0.25)
		pValue = 2 * (1 - statistics.NormalCDF(math.Abs(z)))
		if pValue < 0 {
			pValue = 0
		}
		if pValue > 1 {
			pValue = 1
		}
	}

	winRateA := 0.0
	winRateB := 0.0
	if winsA+winsB > 0 {
		winRateA = float64(winsA) / float64(winsA+winsB)
		winRateB = float64(winsB) / float64(winsA+winsB)
	}

	rep := report.Report{
		RunID:                   fmt.Sprintf("interleave-%s", u.Clock.Now().UTC().Format("20060102T150405Z")),
		Mode:                    "interleaving",
		CreatedAt:               u.Clock.Now().UTC(),
		Version:                 "0.1.0",
		BinaryVersion:           u.Metadata.BinaryVersion,
		GitCommit:               u.Metadata.GitCommit,
		EffectiveConfig:         u.Metadata.EffectiveConfig,
		InputDatasetFingerprint: u.Metadata.InputDatasetFingerprint,
		Artifacts:               u.Metadata.Artifacts,
		Summary: report.Summary{
			CasesEvaluated: requests,
		},
		Interleaving: &report.InterleavingReport{
			Algorithm: algo,
			RankerA:   "A",
			RankerB:   "B",
			Requests:  requests,
			WinsA:     winsA,
			WinsB:     winsB,
			Ties:      ties,
			WinRateA:  winRateA,
			WinRateB:  winRateB,
			PValue:    pValue,
		},
	}
	rep.Summary.Executive = buildExecutiveSummary(rep)

	if err := u.Reporter.Write(ctx, rep, outputPath); err != nil {
		return report.Report{}, err
	}

	return rep, nil
}

type winCount struct {
	A int
	B int
}

func countWins(res interleaving.Result, outcomes []dataset.Outcome) winCount {
	attrByItem := map[string]string{}
	for i, item := range res.Items {
		if i < len(res.Attribution) {
			attrByItem[item] = res.Attribution[i]
		}
	}
	wins := winCount{}
	for _, o := range outcomes {
		if strings.ToLower(o.EventType) != "click" {
			continue
		}
		switch attrByItem[o.ItemID] {
		case "A":
			wins.A++
		case "B":
			wins.B++
		}
	}
	return wins
}
