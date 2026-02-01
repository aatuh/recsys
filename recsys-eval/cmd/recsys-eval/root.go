package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/clock/system"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/logger/std"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/html"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/json"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/reporting/markdown"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/app/workflow"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/reporting"
)

func newRootCmd() *cobra.Command {
	var (
		datasetPath  string
		configPath   string
		outputPath   string
		baselinePath string
		mode         string
		experimentID string
		outputFormat string
	)

	rootCmd := &cobra.Command{
		Use:   "recsys-eval",
		Short: "Recsys evaluation CLI",
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run offline, experiment, OPE, or interleaving evaluation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if datasetPath == "" || configPath == "" || outputPath == "" {
				return errors.New("--dataset, --config, and --output are required")
			}

			var datasetCfg usecase.DatasetConfig
			if err := loadYAMLStrict(datasetPath, &datasetCfg); err != nil {
				return err
			}
			var evalCfg usecase.EvalConfig
			if err := loadYAMLStrict(configPath, &evalCfg); err != nil {
				return err
			}
			if mode == "" {
				mode = evalCfg.Mode
			}
			if mode == "" {
				return errors.New("mode must be set in config or --mode")
			}

			if err := usecase.NormalizeAndValidate(&datasetCfg, &evalCfg, mode); err != nil {
				return err
			}

			clock := system.Clock{}
			logger := std.Logger{}
			reporter, err := selectReporter(outputFormat)
			if err != nil {
				return err
			}

			ctx := context.Background()

			switch strings.ToLower(mode) {
			case "offline":
				exposureReader, err := workflow.BuildExposureReader(datasetCfg.Exposures)
				if err != nil {
					return err
				}
				outcomeReader, err := workflow.BuildOutcomeReader(datasetCfg.Outcomes)
				if err != nil {
					return err
				}
				meta, err := buildReportMetadata(datasetCfg, evalCfg, mode)
				if err != nil {
					return err
				}
				use := usecase.OfflineEvalUsecase{
					Exposures: exposureReader,
					Outcomes:  outcomeReader,
					Reporter:  reporter,
					Clock:     clock,
					Logger:    logger,
					Metadata:  meta,
					Scale:     evalCfg.Scale,
				}
				_, err = use.Run(ctx, evalCfg.Offline, outputPath, baselinePath)
				return err
			case "experiment":
				exposureReader, err := workflow.BuildExposureReader(datasetCfg.Exposures)
				if err != nil {
					return err
				}
				outcomeReader, err := workflow.BuildOutcomeReader(datasetCfg.Outcomes)
				if err != nil {
					return err
				}
				assignmentReader, err := workflow.BuildAssignmentReader(datasetCfg.Assignments)
				if err != nil {
					return err
				}
				if experimentID != "" {
					evalCfg.Experiment.ExperimentID = experimentID
				}
				meta, err := buildReportMetadata(datasetCfg, evalCfg, mode)
				if err != nil {
					return err
				}
				use := usecase.ExperimentUsecase{
					Exposures:   exposureReader,
					Outcomes:    outcomeReader,
					Assignments: assignmentReader,
					Reporter:    reporter,
					Decision:    json.DecisionWriter{},
					Clock:       clock,
					Logger:      logger,
					Metadata:    meta,
					Scale:       evalCfg.Scale,
				}
				decision, runErr := use.RunWithDecision(ctx, evalCfg.Experiment, outputPath)
				if runErr != nil {
					return runErr
				}
				if decision != nil {
					code := decision.ExitCode()
					if code != 0 {
						return ExitError{Code: code, Err: fmt.Errorf("decision=%s", decision.Decision)}
					}
				}
				return nil
			case "ope":
				exposureReader, err := workflow.BuildExposureReader(datasetCfg.Exposures)
				if err != nil {
					return err
				}
				outcomeReader, err := workflow.BuildOutcomeReader(datasetCfg.Outcomes)
				if err != nil {
					return err
				}
				meta, err := buildReportMetadata(datasetCfg, evalCfg, mode)
				if err != nil {
					return err
				}
				use := usecase.OPEUsecase{
					Exposures: exposureReader,
					Outcomes:  outcomeReader,
					Reporter:  reporter,
					Clock:     clock,
					Logger:    logger,
					Metadata:  meta,
					Scale:     evalCfg.Scale,
				}
				_, err = use.Run(ctx, evalCfg.OPE, outputPath)
				return err
			case "interleaving":
				if datasetCfg.Interleaving == nil {
					return errors.New("interleaving dataset config is required")
				}
				rankerA, err := workflow.BuildRankListReader(datasetCfg.Interleaving.RankerA)
				if err != nil {
					return err
				}
				rankerB, err := workflow.BuildRankListReader(datasetCfg.Interleaving.RankerB)
				if err != nil {
					return err
				}
				outcomeReader, err := workflow.BuildOutcomeReader(datasetCfg.Interleaving.Outcomes)
				if err != nil {
					return err
				}
				meta, err := buildReportMetadata(datasetCfg, evalCfg, mode)
				if err != nil {
					return err
				}
				use := usecase.InterleavingUsecase{
					RankerA:  rankerA,
					RankerB:  rankerB,
					Outcomes: outcomeReader,
					Reporter: reporter,
					Clock:    clock,
					Logger:   logger,
					Metadata: meta,
				}
				_, err = use.Run(ctx, evalCfg.Interleaving, outputPath)
				return err
			case "aa-check":
				assignmentReader, err := workflow.BuildAssignmentReader(datasetCfg.Assignments)
				if err != nil {
					return err
				}
				exposureReader, err := workflow.BuildExposureReader(datasetCfg.Exposures)
				if err != nil {
					return err
				}
				outcomeReader, err := workflow.BuildOutcomeReader(datasetCfg.Outcomes)
				if err != nil {
					return err
				}
				meta, err := buildReportMetadata(datasetCfg, evalCfg, mode)
				if err != nil {
					return err
				}
				use := usecase.AACheckUsecase{
					Exposures:   exposureReader,
					Outcomes:    outcomeReader,
					Assignments: assignmentReader,
					Reporter:    reporter,
					Clock:       clock,
					Logger:      logger,
					Metadata:    meta,
				}
				_, err = use.Run(ctx, evalCfg.Experiment, outputPath)
				return err
			default:
				return fmt.Errorf("unsupported mode: %s", mode)
			}
		},
	}
	addRunFlags(runCmd, &datasetPath, &configPath, &outputPath, &baselinePath, &mode, &experimentID, &outputFormat)

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JSONL input against schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaPath, _ := cmd.Flags().GetString("schema")
			inputPath, _ := cmd.Flags().GetString("input")
			if schemaPath == "" || inputPath == "" {
				return errors.New("--schema and --input are required")
			}
			if !strings.HasSuffix(schemaPath, ".json") {
				schemaPath = filepath.Join("schemas", schemaPath+".json")
			}
			_, err := usecase.ValidateJSONL(context.Background(), schemaPath, inputPath)
			return err
		},
	}
	validateCmd.Flags().String("schema", "", "Schema name or path")
	validateCmd.Flags().String("input", "", "JSONL file to validate")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(newVersionCmd())
	return rootCmd
}

func addRunFlags(cmd *cobra.Command, datasetPath, configPath, outputPath, baselinePath, mode, experimentID, outputFormat *string) {
	cmd.Flags().StringVar(datasetPath, "dataset", "", "Dataset config path (YAML)")
	cmd.Flags().StringVar(configPath, "config", "", "Evaluation config path (YAML)")
	cmd.Flags().StringVar(outputPath, "output", "", "Output report path")
	cmd.Flags().StringVar(baselinePath, "baseline", "", "Baseline report path (JSON)")
	cmd.Flags().StringVar(mode, "mode", "", "Evaluation mode: offline | experiment | ope | interleaving | aa-check")
	cmd.Flags().StringVar(experimentID, "experiment-id", "", "Experiment ID override")
	cmd.Flags().StringVar(outputFormat, "output-format", "json", "Output format: json | markdown | html")
}

func selectReporter(format string) (reporting.Writer, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return json.Writer{}, nil
	case "md", "markdown":
		return markdown.Writer{}, nil
	case "html":
		return html.Writer{}, nil
	default:
		return nil, fmt.Errorf("unknown output format: %s", format)
	}
}

func loadYAMLStrict(path string, out any) error {
	// #nosec G304 -- config path provided by operator
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	return dec.Decode(out)
}
