package exposure

import "testing"

func TestEventValidate(t *testing.T) {
	event := Event{
		SchemaVersion: "v1",
		Items: []Item{
			{ItemID: "item-1", Rank: 1},
		},
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid event, got %v", err)
	}
}

func TestEventValidateInvalidItem(t *testing.T) {
	event := Event{
		SchemaVersion: "v1",
		Items: []Item{
			{ItemID: "", Rank: 0},
		},
	}
	if err := event.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}
