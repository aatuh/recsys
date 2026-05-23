(function () {
  "use strict";

  var consentKey = "recsys.analyticsConsent.v1";
  var accepted = "accepted";
  var declined = "declined";
  var scriptId = "recsys-ga4-script";
  var config = window.RECSYS_ANALYTICS_CONFIG || {};
  var measurementId = typeof config.measurementId === "string" ? config.measurementId.trim() : "";
  var enabled = /^G-[A-Z0-9]+$/i.test(measurementId);
  var locale = document.documentElement.lang && document.documentElement.lang.toLowerCase().startsWith("fi") ? "fi" : "en";
  var strings = {
    en: {
      bannerTitle: "Analytics cookies",
      bannerText:
        "RecSys uses Google Analytics only if you accept analytics. We use aggregated page and CTA events to improve the website and documentation.",
      manage: "Manage preferences",
      accept: "Accept all",
      decline: "Decline",
      close: "Close",
      privacy: "Privacy and cookies",
      prefsTitle: "Analytics preferences",
      prefsText:
        "Analytics is optional. We do not send names, emails, user IDs, request IDs, or contact message content to Google Analytics.",
      acceptedStatus: "Analytics is currently accepted.",
      declinedStatus: "Analytics is currently declined.",
      unsetStatus: "Analytics is currently unset.",
      unavailableStatus: "Analytics is not configured for this local build.",
      footer: "Cookie preferences",
    },
    fi: {
      bannerTitle: "Analytiikkaevästeet",
      bannerText:
        "RecSys käyttää Google Analyticsia vain, jos hyväksyt analytiikan. Käytämme koottuja sivu- ja CTA-tapahtumia sivuston ja dokumentaation parantamiseen.",
      manage: "Hallitse asetuksia",
      accept: "Hyväksy kaikki",
      decline: "Kieltäydy",
      close: "Sulje",
      privacy: "Tietosuoja ja evästeet",
      prefsTitle: "Analytiikka-asetukset",
      prefsText:
        "Analytiikka on vapaaehtoista. Emme lähetä nimiä, sähköpostiosoitteita, käyttäjätunnisteita, request ID:itä tai yhteydenottojen sisältöä Google Analyticsiin.",
      acceptedStatus: "Analytiikka on tällä hetkellä hyväksytty.",
      declinedStatus: "Analytiikka on tällä hetkellä kielletty.",
      unsetStatus: "Analytiikka-asetusta ei ole vielä valittu.",
      unavailableStatus: "Analytiikkaa ei ole määritetty tähän paikalliseen buildiin.",
      footer: "Evästeasetukset",
    },
  }[locale];

  function getConsent() {
    try {
      var value = window.localStorage.getItem(consentKey);
      return value === accepted || value === declined ? value : null;
    } catch (_error) {
      return null;
    }
  }

  function setConsent(value) {
    try {
      window.localStorage.setItem(consentKey, value);
    } catch (_error) {
      return;
    }
    if (value === accepted) {
      loadGoogleAnalytics();
    } else {
      removeGoogleAnalytics();
    }
    closeConsentUi();
  }

  function expireCookie(name, domain) {
    var parts = [name + "=", "Path=/", "Expires=Thu, 01 Jan 1970 00:00:00 GMT", "Max-Age=0", "SameSite=Lax"];
    if (domain) {
      parts.push("Domain=" + domain);
    }
    document.cookie = parts.join("; ");
  }

  function expireGoogleAnalyticsCookies() {
    var names = ["_ga", "_gid", "_gat", "_gac_" + measurementId];
    document.cookie.split(";").forEach(function (cookie) {
      var name = cookie.split("=")[0].trim();
      if (name.indexOf("_ga") === 0 || name.indexOf("_gid") === 0 || name.indexOf("_gat") === 0 || name.indexOf("_gac") === 0) {
        names.push(name);
      }
    });

    var host = window.location.hostname;
    var domains = ["", host];
    if (host && host.indexOf(".") !== -1) {
      domains.push("." + host);
    }
    names.forEach(function (name) {
      domains.forEach(function (domain) {
        expireCookie(name, domain);
      });
    });
  }

  function removeGoogleAnalytics() {
    var script = document.getElementById(scriptId);
    if (script) {
      script.remove();
    }
    window["ga-disable-" + measurementId] = true;
    expireGoogleAnalyticsCookies();
  }

  function gtag() {
    window.dataLayer = window.dataLayer || [];
    window.dataLayer.push(arguments);
  }

  function loadGoogleAnalytics() {
    if (!enabled || getConsent() !== accepted) {
      return;
    }
    window["ga-disable-" + measurementId] = false;
    window.dataLayer = window.dataLayer || [];
    window.gtag = window.gtag || gtag;
    window.gtag("js", new Date());
    window.gtag("config", measurementId, {
      allow_ad_personalization_signals: false,
      allow_google_signals: false,
      anonymize_ip: true,
      send_page_view: true,
    });

    if (document.getElementById(scriptId)) {
      return;
    }
    var script = document.createElement("script");
    script.id = scriptId;
    script.async = true;
    script.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(measurementId);
    document.head.appendChild(script);
  }

  function sanitize(value, fallback) {
    var text = typeof value === "string" ? value : fallback || "";
    return text.replace(/[^a-zA-Z0-9_./:-]/g, "_").slice(0, 120);
  }

  function targetPathFor(link) {
    if (link.dataset.analyticsTargetPath) {
      return sanitize(link.dataset.analyticsTargetPath, "");
    }
    try {
      var url = new URL(link.href, window.location.href);
      return url.origin === window.location.origin ? url.pathname : url.hostname;
    } catch (_error) {
      return "";
    }
  }

  function trackCtaClick(event) {
    if (getConsent() !== accepted || typeof window.gtag !== "function") {
      return;
    }
    if (!event.target || typeof event.target.closest !== "function") {
      return;
    }
    var link = event.target.closest("[data-analytics-event]");
    if (!link) {
      return;
    }
    var eventName = sanitize(link.dataset.analyticsEvent, "cta_click").replace(/[^a-zA-Z0-9_]/g, "_").slice(0, 40);
    window.gtag("event", eventName || "cta_click", {
      cta_id: sanitize(link.dataset.analyticsCtaId, "unknown"),
      cta_location: sanitize(link.dataset.analyticsCtaLocation, "unknown"),
      locale: locale,
      target_path: targetPathFor(link),
    });
  }

  function closeConsentUi() {
    var ui = document.getElementById("recsys-consent");
    if (ui) {
      ui.remove();
    }
  }

  function statusText() {
    var consent = getConsent();
    if (consent === accepted) {
      return strings.acceptedStatus;
    }
    if (consent === declined) {
      return strings.declinedStatus;
    }
    if (!enabled) {
      return strings.unavailableStatus;
    }
    return strings.unsetStatus;
  }

  function button(text, className, onClick) {
    var element = document.createElement("button");
    element.type = "button";
    element.className = className;
    element.textContent = text;
    element.addEventListener("click", onClick);
    return element;
  }

  function renderConsentUi(mode) {
    closeConsentUi();
    var panel = document.createElement("section");
    panel.id = "recsys-consent";
    panel.className = "recsys-consent";
    panel.setAttribute("role", "dialog");
    panel.setAttribute("aria-labelledby", "recsys-consent-title");

    var content = document.createElement("div");
    content.className = "recsys-consent__panel";
    var copy = document.createElement("div");
    var title = document.createElement("h2");
    title.id = "recsys-consent-title";
    title.textContent = mode === "preferences" ? strings.prefsTitle : strings.bannerTitle;
    var text = document.createElement("p");
    text.textContent = mode === "preferences" ? strings.prefsText : strings.bannerText;
    var status = document.createElement("p");
    status.className = "recsys-consent__status";
    status.textContent = statusText();
    var privacy = document.createElement("a");
    privacy.href = locale === "fi" ? "/fi/tietosuoja/" : "/privacy/";
    privacy.textContent = strings.privacy;
    copy.append(title, text, status, privacy);

    var actions = document.createElement("div");
    actions.className = "recsys-consent__actions";
    if (enabled) {
      actions.append(
        button(strings.accept, "recsys-consent__button recsys-consent__button--primary", function () {
          setConsent(accepted);
        }),
        button(strings.decline, "recsys-consent__button", function () {
          setConsent(declined);
        }),
      );
    }
    if (mode === "preferences") {
      actions.append(button(strings.close, "recsys-consent__button", closeConsentUi));
    } else {
      actions.append(
        button(strings.manage, "recsys-consent__button", function () {
          renderConsentUi("preferences");
        }),
      );
    }

    content.append(copy, actions);
    panel.appendChild(content);
    document.body.appendChild(panel);
    var firstAction = actions.querySelector("button");
    if (firstAction) {
      firstAction.focus({ preventScroll: true });
    }
  }

  function installPreferencesControl() {
    var controls = Array.prototype.slice.call(document.querySelectorAll("[data-analytics-preferences]"));
    if (controls.length === 0) {
      var footer = document.querySelector(".md-footer-meta__inner, .md-footer, footer");
      if (footer) {
        var control = document.createElement("button");
        control.type = "button";
        control.className = "recsys-cookie-preferences";
        control.setAttribute("data-analytics-preferences", "");
        control.textContent = strings.footer;
        footer.appendChild(control);
        controls.push(control);
      }
    }
    controls.forEach(function (control) {
      control.addEventListener("click", function () {
        renderConsentUi("preferences");
      });
    });
  }

  function installStyles() {
    if (document.getElementById("recsys-consent-style")) {
      return;
    }
    var style = document.createElement("style");
    style.id = "recsys-consent-style";
    style.textContent =
      ".recsys-consent{position:fixed;inset:auto 1rem 1rem;z-index:9999;max-width:980px;margin:0 auto;color:#172026}" +
      ".recsys-consent__panel{display:grid;gap:1rem;background:#fbfdfb;border:1px solid #d4ded9;border-radius:8px;box-shadow:0 24px 70px rgba(23,32,38,.18);padding:1rem}" +
      ".recsys-consent h2{font:700 1.05rem/1.2 system-ui,sans-serif;margin:0 0 .35rem}" +
      ".recsys-consent p{font:400 .94rem/1.45 system-ui,sans-serif;margin:.3rem 0;color:#5f6c72}" +
      ".recsys-consent a{color:#0f766e;text-decoration:underline;text-underline-offset:.2em}" +
      ".recsys-consent__status{font-weight:700;color:#172026!important}" +
      ".recsys-consent__actions{display:flex;flex-wrap:wrap;gap:.55rem;align-items:center}" +
      ".recsys-consent__button,.recsys-cookie-preferences{border:1px solid #d4ded9;border-radius:6px;background:#fff;color:#172026;cursor:pointer;font:700 .92rem/1 system-ui,sans-serif;min-height:2.5rem;padding:.72rem .85rem}" +
      ".recsys-consent__button--primary{background:#172026;border-color:#172026;color:#fff}" +
      ".recsys-consent__button:focus-visible,.recsys-cookie-preferences:focus-visible{outline:3px solid rgba(15,118,110,.45);outline-offset:3px}" +
      ".recsys-cookie-preferences{margin:.5rem;background:transparent;color:inherit}" +
      "@media (min-width:720px){.recsys-consent{left:1.2rem;right:1.2rem}.recsys-consent__panel{grid-template-columns:1fr auto;align-items:center;padding:1.1rem 1.2rem}.recsys-consent__actions{justify-content:flex-end}}";
    document.head.appendChild(style);
  }

  function boot() {
    installStyles();
    installPreferencesControl();
    if (!enabled) {
      return;
    }
    if (getConsent() === accepted) {
      loadGoogleAnalytics();
    } else if (getConsent() === declined) {
      removeGoogleAnalytics();
    } else {
      renderConsentUi("banner");
    }
    document.addEventListener("click", trackCtaClick, true);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
