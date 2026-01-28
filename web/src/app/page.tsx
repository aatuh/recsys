import { Button, Card, Pill, SectionHeader } from "@api-boilerplate-core/ui";
import { appDefaults } from "@foo/config";

const highlights = [
  {
    title: "Typed from API to UI",
    body: "Generate TypeScript types from the Go swagger output and keep DTOs honest across layers.",
  },
  {
    title: "Thin HTTP, rich domain",
    body: "Adapters map API DTOs into domain models so your UI stays clean and portable.",
  },
  {
    title: "Production defaults",
    body: "Retries, RFC-7807 errors, consent-ready theme toggles, and content tooling included.",
  },
];

const stack = [
  { label: "Go", detail: "api-toolkit + migrations" },
  { label: "Next.js", detail: "App Router, Tailwind" },
  { label: "Type safety", detail: "openapi-typescript" },
];

export default function HomePage() {
  return (
    <div className="space-y-12">
      <section className="grid gap-8 lg:grid-cols-[1.2fr_0.8fr] lg:items-center">
        <div className="space-y-6">
          <Pill>Production-ready starter</Pill>
          <h1 className="text-4xl font-semibold leading-tight text-foreground sm:text-5xl">
            Ship a Go API and a Next.js UI without the glue work.
          </h1>
          <p className="max-w-xl text-base text-muted sm:text-lg">
            API Boilerplate packages a clean architecture, typed API client, and
            reusable web primitives so you can focus on your domain.
          </p>
          <div className="flex flex-wrap gap-3">
            <Button href="/foos">Explore Foo API</Button>
            <Button href="/docs" variant="secondary">
              Read docs
            </Button>
          </div>
        </div>
        <Card className="bg-card">
          <div className="space-y-4">
            <div className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
              Quickstart
            </div>
            <pre className="rounded-2xl bg-surface-muted p-4 text-xs text-muted-strong">
              {`make dev
cd api && make migrate-up
cd ../web && pnpm dev`}
            </pre>
            <div className="grid gap-3 sm:grid-cols-3">
              {stack.map((item) => (
                <div
                  key={item.label}
                  className="rounded-xl border border-border bg-surface px-3 py-2"
                >
                  <div className="text-xs font-semibold uppercase tracking-[0.15em] text-muted-strong">
                    {item.label}
                  </div>
                  <div className="text-sm text-muted">{item.detail}</div>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </section>

      <section className="space-y-6">
        <SectionHeader
          title="What you get"
          description="Core packages and patterns you can lift into any product."
        />
        <div className="grid gap-6 lg:grid-cols-3">
          {highlights.map((item) => (
            <Card key={item.title}>
              <div className="space-y-2">
                <h3 className="text-lg font-semibold text-foreground">
                  {item.title}
                </h3>
                <p className="text-sm text-muted">{item.body}</p>
              </div>
            </Card>
          ))}
        </div>
      </section>

      <section className="grid gap-6 lg:grid-cols-[1fr_1.2fr]">
        <Card>
          <div className="space-y-3">
            <div className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
              Demo defaults
            </div>
            <p className="text-sm text-muted">
              The Foo API expects an org and namespace. The UI uses the defaults
              below, but you can override them in `.env`.
            </p>
            <div className="rounded-2xl border border-border bg-surface-muted px-4 py-3 text-sm text-muted-strong">
              <div>Org ID: {appDefaults.orgId}</div>
              <div>Namespace: {appDefaults.namespace}</div>
            </div>
          </div>
        </Card>
        <Card className="bg-card">
          <div className="space-y-4">
            <div className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
              Next steps
            </div>
            <ul className="space-y-2 text-sm text-muted">
              <li>
                Map real domain models under `packages/services/foo/domain`.
              </li>
              <li>Replace Foo pages with your product flows.</li>
              <li>Extend `api/` with new services and generate types.</li>
            </ul>
            <Button href="/docs/architecture" variant="ghost" size="md">
              Explore architecture notes
            </Button>
          </div>
        </Card>
      </section>
    </div>
  );
}
