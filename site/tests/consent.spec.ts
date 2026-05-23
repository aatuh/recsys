import { expect, test, type Page } from "@playwright/test";

const consentKey = "recsys.analyticsConsent.v1";

async function collectGoogleTagRequests(page: Page): Promise<string[]> {
  const requests: string[] = [];
  await page.route("https://www.googletagmanager.com/**", async (route) => {
    requests.push(route.request().url());
    await route.fulfill({
      body: "window.__recsysGtagLoaded = true;",
      contentType: "application/javascript",
      status: 200,
    });
  });
  return requests;
}

test("does not load Google Analytics before consent", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/");

  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Accept all" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Decline" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Manage preferences" })).toBeVisible();
  await expect.poll(() => requests.length).toBe(0);
});

test("accepting analytics loads GA and tracks only approved CTA parameters", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/");

  await page.getByRole("button", { name: "Accept all" }).click();

  await expect
    .poll(() => requests.some((url) => url.includes("https://www.googletagmanager.com/gtag/js?id=G-TEST0000")))
    .toBe(true);
  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toHaveCount(0);
  await expect(page.evaluate((key) => window.localStorage.getItem(key), consentKey)).resolves.toBe("accepted");

  await page.evaluate(() => {
    const link = document.querySelector<HTMLAnchorElement>('[data-analytics-cta-id="home_start_evaluation"]');
    if (!link) {
      throw new Error("home_start_evaluation CTA is missing");
    }
    link.addEventListener("click", (event) => event.preventDefault(), { once: true });
    link.click();
  });

  await expect
    .poll(async () =>
      page.evaluate(() => {
        const dataLayer = (window as Window & { dataLayer?: unknown[] }).dataLayer ?? [];
        return dataLayer.some((entry) => {
          if (typeof entry !== "object" || entry === null || !("length" in entry)) {
            return false;
          }
          const values = Array.from(entry as ArrayLike<unknown>);
          return (
            values[0] === "event" &&
            values[1] === "cta_click" &&
            typeof values[2] === "object" &&
            values[2] !== null &&
            (values[2] as Record<string, string>).cta_id === "home_start_evaluation" &&
            (values[2] as Record<string, string>).cta_location === "hero" &&
            (values[2] as Record<string, string>).target_path === "/contact/" &&
            (values[2] as Record<string, string>).locale === "en"
          );
        });
      }),
    )
    .toBe(true);

  await page.reload();
  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toHaveCount(0);
});

test("declining analytics persists denial and preferences can be reopened", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/");

  await page.getByRole("button", { name: "Decline" }).click();
  await expect(page.evaluate((key) => window.localStorage.getItem(key), consentKey)).resolves.toBe("declined");
  await page.reload();

  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toHaveCount(0);
  await expect.poll(() => requests.length).toBe(0);

  await page.getByRole("button", { name: "Cookie preferences" }).click();
  await expect(page.getByRole("dialog", { name: "Analytics preferences" })).toBeVisible();
  await expect(page.getByText("Analytics is currently declined.")).toBeVisible();
});

test("preferences explain when analytics is not configured for a local build", async ({ page }) => {
  await page.route("**/assets/analytics-config.js", async (route) => {
    await route.fulfill({
      body: 'window.RECSYS_ANALYTICS_CONFIG = {"measurementId":""};',
      contentType: "application/javascript",
      status: 200,
    });
  });

  await page.goto("/privacy/");

  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toHaveCount(0);
  await page.getByRole("button", { name: "Cookie preferences" }).first().click();
  await expect(page.getByRole("dialog", { name: "Analytics preferences" })).toBeVisible();
  await expect(page.getByText("Analytics is not configured for this local build.")).toBeVisible();
  await expect(page.getByRole("button", { name: "Accept all" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Decline" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Close" })).toBeVisible();
});

test("revoking accepted analytics removes GA and keeps denial after reload", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/");

  await page.getByRole("button", { name: "Accept all" }).click();
  await expect
    .poll(() => requests.some((url) => url.includes("https://www.googletagmanager.com/gtag/js?id=G-TEST0000")))
    .toBe(true);
  await expect(page.locator("#recsys-ga4-script")).toHaveCount(1);

  await page.getByRole("button", { name: "Cookie preferences" }).click();
  await page.getByRole("button", { name: "Decline" }).click();
  await expect(page.evaluate((key) => window.localStorage.getItem(key), consentKey)).resolves.toBe("declined");
  await expect(page.locator("#recsys-ga4-script")).toHaveCount(0);

  const requestCount = requests.length;
  await page.reload();
  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toHaveCount(0);
  await expect.poll(() => requests.length).toBe(requestCount);
});

test("renders Finnish consent copy on Finnish pages", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/fi/");

  await expect(page.getByRole("dialog", { name: "Analytiikkaevästeet" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Hyväksy kaikki" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Kieltäydy" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Hallitse asetuksia" })).toBeVisible();
  await expect.poll(() => requests.length).toBe(0);
});

test("technical documentation uses the same consent gate", async ({ page }) => {
  const requests = await collectGoogleTagRequests(page);
  await page.goto("/documentation/technical/");

  await expect(page.getByRole("dialog", { name: "Analytics cookies" })).toBeVisible();
  await expect.poll(() => requests.length).toBe(0);

  await page.getByRole("button", { name: "Accept all" }).click();
  await expect
    .poll(() => requests.some((url) => url.includes("https://www.googletagmanager.com/gtag/js?id=G-TEST0000")))
    .toBe(true);
});
