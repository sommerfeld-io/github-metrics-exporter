package acceptance_test

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cucumber/godog"
)

// capturingHandler wraps an existing slog.Handler and records every log entry.
// It is goroutine-safe; the HTTP server handler runs in a separate goroutine.
type capturingHandler struct {
	mu      sync.Mutex
	records []slog.Record
	base    slog.Handler
}

func (h *capturingHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	h.records = append(h.records, r.Clone())
	h.mu.Unlock()
	return h.base.Handle(context.Background(), r)
}

func (h *capturingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *capturingHandler) WithGroup(_ string) slog.Handler      { return h }

// scrapeObservabilityState captures log entries written during a single scenario.
type scrapeObservabilityState struct {
	handler    *capturingHandler
	origLogger *slog.Logger
}

func (s *scrapeObservabilityState) installHandler(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	s.origLogger = slog.Default()
	s.handler = &capturingHandler{base: s.origLogger.Handler()}
	slog.SetDefault(slog.New(s.handler))
	return ctx, nil
}

func (s *scrapeObservabilityState) restoreHandler(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	slog.SetDefault(s.origLogger)
	return ctx, nil
}

func (s *scrapeObservabilityState) aLogEntryWithMessageIsWritten(msg string) error {
	s.handler.mu.Lock()
	defer s.handler.mu.Unlock()
	for _, r := range s.handler.records {
		if r.Message == msg {
			return nil
		}
	}
	return fmt.Errorf("expected log entry %q, not found among %d captured entries", msg, len(s.handler.records))
}

// InitializeMetricsScrapeScenario registers log-observability step definitions with GoDog.
// The Given and When steps are reused from InitializeRepositoryScenario; only the Then step
// and the Before/After hooks for log capture are registered here.
func InitializeMetricsScrapeScenario(ctx *godog.ScenarioContext) {
	s := &scrapeObservabilityState{}
	ctx.Before(s.installHandler)
	ctx.After(s.restoreHandler)
	ctx.Step(`^a log entry with message "([^"]*)" is written$`, s.aLogEntryWithMessageIsWritten)
}
