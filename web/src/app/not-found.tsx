import Link from "next/link";
import { Card } from "@api-boilerplate-core/ui";

export default function NotFound() {
  return (
    <Card className="mx-auto max-w-xl text-center">
      <div className="space-y-3">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
          404
        </p>
        <h1 className="text-2xl font-semibold text-foreground">
          Page not found
        </h1>
        <p className="text-sm text-muted">
          The page you are looking for does not exist or has moved.
        </p>
        <Link className="text-sm font-semibold text-primary" href="/">
          Go back home
        </Link>
      </div>
    </Card>
  );
}
