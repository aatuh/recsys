package audit

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"recsys/internal/store"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WriterConfig configures the async decision trace writer.
type WriterConfig struct {
	Enabled           bool
	QueueSize         int
	BatchSize         int
	FlushInterval     time.Duration
	SampleDefaultRate float64
	NamespaceRates    map[string]float64
}

// Recorder defines behaviour for persisting decision traces.
type Recorder interface {
	Record(trace *Trace)
	Close(ctx context.Context) error
}

type Store interface {
	InsertDecisionTraces(ctx context.Context, rows []store.DecisionTraceInsert) error
}

type writer struct {
	cfg     WriterConfig
	store   Store
	logger  *zap.Logger
	ch      chan *Trace
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	enabled bool
	randMu  sync.Mutex
	rnd     *rand.Rand
}

// NewWriter creates a new decision trace writer. When disabled it returns a no-op recorder.
func NewWriter(parent context.Context, s Store, logger *zap.Logger, cfg WriterConfig) Recorder {
	if !cfg.Enabled {
		return noopRecorder{}
	}

	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 200
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 250 * time.Millisecond
	}
	if cfg.SampleDefaultRate <= 0 {
		cfg.SampleDefaultRate = 1.0
	}

	ctx, cancel := context.WithCancel(parent)
	w := &writer{
		cfg:     cfg,
		store:   s,
		logger:  logger,
		ch:      make(chan *Trace, cfg.QueueSize),
		ctx:     ctx,
		cancel:  cancel,
		enabled: true,
		rnd:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	w.wg.Add(1)
	go w.loop()
	return w
}

func (w *writer) Record(trace *Trace) {
	if !w.enabled || trace == nil {
		return
	}
	if !w.shouldSample(trace.Namespace) {
		return
	}

	select {
	case w.ch <- trace:
	default:
		if w.logger != nil {
			w.logger.Warn("decision trace queue full; dropping", zap.String("namespace", trace.Namespace))
		}
	}
}

func (w *writer) Close(ctx context.Context) error {
	if !w.enabled {
		return nil
	}

	w.cancel()
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *writer) shouldSample(namespace string) bool {
	rate := w.cfg.SampleDefaultRate
	if w.cfg.NamespaceRates != nil {
		if v, ok := w.cfg.NamespaceRates[namespace]; ok {
			rate = v
		}
	}
	if rate >= 1 {
		return true
	}
	if rate <= 0 {
		return false
	}
	w.randMu.Lock()
	defer w.randMu.Unlock()
	return w.rnd.Float64() < rate
}

func (w *writer) loop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]*Trace, 0, w.cfg.BatchSize)

	flush := func(reason string) {
		if len(batch) == 0 {
			return
		}
		rows, err := w.buildRows(batch)
		batch = batch[:0]
		if err != nil {
			if w.logger != nil {
				w.logger.Error("build decision rows", zap.Error(err))
			}
			return
		}
		if w.logger != nil {
			w.logger.Debug("flushing decision traces", zap.String("reason", reason), zap.Int("rows", len(rows)))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := w.store.InsertDecisionTraces(ctx, rows); err != nil && !errors.Is(err, context.Canceled) {
			if w.logger != nil {
				w.logger.Error("insert decision traces", zap.Error(err))
			}
		}
	}

	for {
		select {
		case <-w.ctx.Done():
			flush("shutdown")
			return
		case trace := <-w.ch:
			if trace == nil {
				continue
			}
			batch = append(batch, trace)
			if len(batch) >= w.cfg.BatchSize {
				flush("batch")
			}
		case <-ticker.C:
			flush("interval")
		}
	}
}

func (w *writer) buildRows(traces []*Trace) ([]store.DecisionTraceInsert, error) {
	rows := make([]store.DecisionTraceInsert, 0, len(traces))
	for _, tr := range traces {
		row, err := toDecisionTraceInsert(tr)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func toDecisionTraceInsert(tr *Trace) (store.DecisionTraceInsert, error) {
	orgID, err := uuid.Parse(tr.OrgID)
	if err != nil {
		return store.DecisionTraceInsert{}, err
	}
	cfgJSON, err := json.Marshal(tr.Config)
	if err != nil {
		return store.DecisionTraceInsert{}, err
	}
	candidatesJSON, err := json.Marshal(tr.Candidates)
	if err != nil {
		return store.DecisionTraceInsert{}, err
	}
	finalsJSON, err := json.Marshal(tr.FinalItems)
	if err != nil {
		return store.DecisionTraceInsert{}, err
	}

	var constraintsJSON []byte
	if tr.Constraints != nil {
		if constraintsJSON, err = json.Marshal(tr.Constraints); err != nil {
			return store.DecisionTraceInsert{}, err
		}
	}

	var banditJSON []byte
	if tr.Bandit != nil {
		if banditJSON, err = json.Marshal(tr.Bandit); err != nil {
			return store.DecisionTraceInsert{}, err
		}
	}

	var mmrJSON []byte
	if len(tr.MMR) > 0 {
		if mmrJSON, err = json.Marshal(tr.MMR); err != nil {
			return store.DecisionTraceInsert{}, err
		}
	}

	var capsJSON []byte
	if len(tr.Caps) > 0 {
		if capsJSON, err = json.Marshal(tr.Caps); err != nil {
			return store.DecisionTraceInsert{}, err
		}
	}

	var extrasJSON []byte
	if len(tr.Extras) > 0 {
		if extrasJSON, err = json.Marshal(tr.Extras); err != nil {
			return store.DecisionTraceInsert{}, err
		}
	}

	var kPtr *int
	if tr.K > 0 {
		k := tr.K
		kPtr = &k
	}

	var surfacePtr *string
	if tr.Surface != "" {
		s := tr.Surface
		surfacePtr = &s
	}

	var requestIDPtr *string
	if tr.RequestID != "" {
		r := tr.RequestID
		requestIDPtr = &r
	}

	var userHashPtr *string
	if tr.UserHash != "" {
		uh := tr.UserHash
		userHashPtr = &uh
	}

	ts := tr.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	return store.DecisionTraceInsert{
		DecisionID:      tr.DecisionID,
		OrgID:           orgID,
		Timestamp:       ts,
		Namespace:       tr.Namespace,
		Surface:         surfacePtr,
		RequestID:       requestIDPtr,
		UserHash:        userHashPtr,
		K:               kPtr,
		ConstraintsJSON: constraintsJSON,
		EffectiveConfig: cfgJSON,
		BanditJSON:      banditJSON,
		CandidatesPre:   candidatesJSON,
		FinalItems:      finalsJSON,
		MMRInfo:         mmrJSON,
		Caps:            capsJSON,
		Extras:          extrasJSON,
	}, nil
}

type noopRecorder struct{}

func (noopRecorder) Record(*Trace) {}

func (noopRecorder) Close(context.Context) error { return nil }
