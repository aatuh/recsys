import { MarkdownPage } from "@api-boilerplate-core/content";
import { getRequestLocale } from "@foo/i18n/locale.server";

export default async function DocsPage({
  params,
}: {
  params: { slug?: string[] };
}) {
  const locale = await getRequestLocale();
  const slug = params.slug?.length ? `docs/${params.slug.join("/")}` : "docs";
  return <MarkdownPage slug={slug} locale={locale} notFoundOnMissing />;
}
