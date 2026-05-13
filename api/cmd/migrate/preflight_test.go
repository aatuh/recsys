package main

import "testing"

func TestQuoteTableIdentDefaultsAndQuotes(t *testing.T) {
	t.Parallel()

	got, err := quoteTableIdent("")
	if err != nil {
		t.Fatalf("quoteTableIdent() error = %v", err)
	}
	if got != `"schema_migrations"` {
		t.Fatalf("quoteTableIdent() = %q", got)
	}

	got, err = quoteTableIdent("public.schema_migrations")
	if err != nil {
		t.Fatalf("quoteTableIdent() schema error = %v", err)
	}
	if got != `"public"."schema_migrations"` {
		t.Fatalf("quoteTableIdent() schema = %q", got)
	}
}

func TestQuoteTableIdentRejectsSQLFragments(t *testing.T) {
	t.Parallel()

	if _, err := quoteTableIdent("schema_migrations; drop table tenants"); err == nil {
		t.Fatalf("expected invalid identifier error")
	}
}
