import {
  composeConsentConfig,
  getVanillaConsentConfig,
  type ConsentConfig,
  type ConsentOverrides,
} from "@api-boilerplate-core/legal";
import { resolveLegalLocale, tokensByLocale } from "./index";

const consentOverridesByLocale: Record<string, ConsentOverrides> = {};

export function getCookieConsentConfig(locale?: string | null): ConsentConfig {
  const resolvedLocale = resolveLegalLocale(locale);
  const template =
    getVanillaConsentConfig(resolvedLocale) ?? getVanillaConsentConfig("en");
  if (!template) {
    throw new Error("[legal] Missing vanilla consent config.");
  }
  const overrides = consentOverridesByLocale[resolvedLocale];
  const tokens = tokensByLocale[resolvedLocale] ?? tokensByLocale["en"] ?? {};
  return composeConsentConfig(template, overrides, tokens);
}
