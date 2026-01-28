"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { appDefaults } from "@foo/config";
import { useFooActions, useFooList } from "@foo/hooks";
import { Button, Card, InputField, Pill } from "@api-boilerplate-core/ui";

const toIsoDate = (value: string) => {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) return value;
  return parsed.toISOString().replace("T", " ").replace("Z", " UTC");
};

export function FoosClient() {
  const [orgId, setOrgId] = useState(appDefaults.orgId);
  const [namespace, setNamespace] = useState(appDefaults.namespace);
  const [search, setSearch] = useState("");
  const [name, setName] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const filters = useMemo(() => {
    const trimmedSearch = search.trim();
    return {
      orgId,
      namespace,
      ...(trimmedSearch ? { search: trimmedSearch } : {}),
    };
  }, [orgId, namespace, search]);

  const { items, meta, status, error, reload } = useFooList(filters);
  const {
    createFoo,
    removeFoo,
    status: actionStatus,
    error: actionError,
    clearError,
  } = useFooActions({
    onSuccess: () => reload(),
  });

  const handleCreate = async (event: React.FormEvent) => {
    event.preventDefault();
    clearError();
    setFormError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setFormError("Name is required.");
      return;
    }
    const created = await createFoo({ orgId, namespace, name: trimmed });
    if (created) {
      setName("");
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-card">
        <form className="space-y-4" onSubmit={handleCreate}>
          <div className="grid gap-4 md:grid-cols-3">
            <InputField
              label="Org ID"
              value={orgId}
              onChange={(event) => setOrgId(event.target.value)}
              placeholder="org-demo"
              requiredLabel
            />
            <InputField
              label="Namespace"
              value={namespace}
              onChange={(event) => setNamespace(event.target.value)}
              placeholder="default"
              requiredLabel
            />
            <InputField
              label="Search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search by name"
            />
          </div>
          <div className="grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
            <InputField
              label="New foo name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Enter a name"
              {...(formError ? { error: formError } : {})}
              requiredLabel
            />
            <div className="flex items-center gap-3">
              <Button type="submit" disabled={actionStatus === "saving"}>
                {actionStatus === "saving" ? "Saving..." : "Create foo"}
              </Button>
              <Button type="button" variant="ghost" size="md" onClick={reload}>
                Reload
              </Button>
            </div>
          </div>
          {actionError ? (
            <p className="text-sm text-red-600">{actionError}</p>
          ) : null}
        </form>
      </Card>

      <Card>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm font-semibold text-foreground">Results</p>
            <p className="text-xs text-muted">
              {meta
                ? `${meta.count} of ${meta.total} total`
                : "Fetching data..."}
            </p>
          </div>
          <Pill tone={status === "error" ? "muted" : "success"}>
            {status === "loading"
              ? "Loading"
              : status === "error"
              ? "Error"
              : "Ready"}
          </Pill>
        </div>
        {status === "error" ? (
          <p className="mt-4 text-sm text-red-600">{error}</p>
        ) : null}
        <div className="mt-6 grid gap-4">
          {items.length === 0 && status !== "loading" ? (
            <p className="text-sm text-muted">
              No foos yet. Create the first one above.
            </p>
          ) : null}
          {items.map((item) => (
            <div
              key={item.id}
              className="rounded-2xl border border-border bg-surface px-4 py-3"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <Link
                    className="text-sm font-semibold text-foreground hover:text-primary"
                    href={`/foos/${item.id}`}
                  >
                    {item.name}
                  </Link>
                  <p className="text-xs text-muted">ID: {item.id}</p>
                </div>
                <Button
                  type="button"
                  size="md"
                  variant="ghost"
                  onClick={() => removeFoo(item.id)}
                >
                  Delete
                </Button>
              </div>
              <div className="mt-3 grid gap-2 text-xs text-muted sm:grid-cols-3">
                <div>
                  <span className="text-muted-strong">Org:</span> {item.orgId}
                </div>
                <div>
                  <span className="text-muted-strong">Namespace:</span>{" "}
                  {item.namespace}
                </div>
                <div>
                  <span className="text-muted-strong">Updated:</span>{" "}
                  {toIsoDate(item.updatedAt)}
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
