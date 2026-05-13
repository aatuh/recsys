package mapper

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/services/adminsvc"
)

func TestAuditLogResponseRedactsDetailsWhenNotIncluded(t *testing.T) {
	t.Parallel()

	resp := AuditLogResponse(auditLogFixture(t), false)
	if len(resp.Entries) != 1 {
		t.Fatalf("entries len = %d", len(resp.Entries))
	}
	entry := resp.Entries[0]
	if entry.Before != nil || entry.After != nil || entry.Extra != nil {
		t.Fatalf("expected audit details to be redacted: %+v", entry)
	}
	if entry.Action != "config.update" || entry.ActorSub != "operator-1" {
		t.Fatalf("expected metadata to remain visible: %+v", entry)
	}
}

func TestAuditLogResponseIncludesDetailsWhenAllowed(t *testing.T) {
	t.Parallel()

	resp := AuditLogResponse(auditLogFixture(t), true)
	entry := resp.Entries[0]
	after, ok := entry.After.(map[string]any)
	if !ok {
		t.Fatalf("after state type = %T", entry.After)
	}
	if after["secret"] != "redacted-in-viewer-response" {
		t.Fatalf("after state = %#v", after)
	}
	if entry.Before == nil || entry.Extra == nil {
		t.Fatalf("expected before and extra details")
	}
}

func auditLogFixture(t *testing.T) adminsvc.AuditLog {
	t.Helper()
	raw := func(v any) json.RawMessage {
		t.Helper()
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal fixture: %v", err)
		}
		return data
	}
	return adminsvc.AuditLog{
		TenantID: "tenant-1",
		Entries: []adminsvc.AuditEntry{{
			ID:         7,
			OccurredAt: time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC),
			TenantID:   "tenant-1",
			ActorSub:   "operator-1",
			ActorType:  "jwt",
			Action:     "config.update",
			Before:     raw(map[string]any{"old": true}),
			After:      raw(map[string]any{"secret": "redacted-in-viewer-response"}),
			Extra:      raw(map[string]any{"request": "req-1"}),
		}},
	}
}
