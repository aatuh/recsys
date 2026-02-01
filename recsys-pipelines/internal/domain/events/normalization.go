package events

import "strings"

func NormalizeID(id string) string {
	return strings.TrimSpace(id)
}

func (e ExposureEvent) Normalized() ExposureEvent {
	n := e
	n.Tenant = NormalizeID(n.Tenant)
	n.Surface = NormalizeID(n.Surface)
	n.UserID = NormalizeID(n.UserID)
	n.SessionID = NormalizeID(n.SessionID)
	n.RequestID = NormalizeID(n.RequestID)
	n.ItemID = NormalizeID(n.ItemID)
	return n
}
