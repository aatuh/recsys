(function () {
  const seen = new WeakSet();
  let schemeObserverAttached = false;

  function isDarkMode() {
    const darkSchemeName = "slate";
    const schemeAttr = document.body?.getAttribute("data-md-color-scheme");
    const isMediaPrefersScheme =
      document.body?.getAttribute("data-md-color-media") ===
      "(prefers-color-scheme: dark)";

    if (!isMediaPrefersScheme) {
      return schemeAttr === darkSchemeName;
    }

    const computed = window
      .getComputedStyle(document.body)
      .getPropertyValue("color-scheme")
      .trim();
    return computed === "dark";
  }

  function resizeToContent(iframe) {
    try {
      const doc = iframe.contentDocument;
      if (!doc) return;

      // Ensure layout has settled
      const h = Math.max(
        doc.documentElement.scrollHeight,
        doc.body ? doc.body.scrollHeight : 0,
        doc.documentElement.offsetHeight,
        doc.body ? doc.body.offsetHeight : 0,
      );
      if (h <= 0) return;

      // A little padding to avoid 1px scrollbars from rounding
      iframe.style.height = h + 8 + "px";
      iframe.style.width = "100%";
      iframe.style.border = "0";
      iframe.style.display = "block";
      iframe.style.overflow = "hidden";
      iframe.setAttribute("scrolling", "no");
    } catch {
      // If cross-origin, you can't introspect height; fallback:
      iframe.style.height = "85vh";
    }
  }

  function syncTheme(iframe) {
    try {
      const win = iframe.contentWindow;
      if (!win) return;
      const dark = isDarkMode();
      window.__init_is_dark_mode = dark;
      if (dark) {
        if (typeof win.enable_dark_mode === "function") win.enable_dark_mode();
      } else {
        if (typeof win.disable_dark_mode === "function") win.disable_dark_mode();
      }
    } catch {
      // ignore
    }
  }

  function attachSchemeObserver() {
    if (schemeObserverAttached) return;
    if (!document.body || typeof MutationObserver !== "function") return;
    schemeObserverAttached = true;

    window.__init_is_dark_mode = isDarkMode();

    const observer = new MutationObserver(() => {
      window.__init_is_dark_mode = isDarkMode();
      document.querySelectorAll("iframe.swagger-ui-iframe").forEach((iframe) => {
        syncTheme(iframe);
        resizeToContent(iframe);
      });
    });
    observer.observe(document.body, { attributeFilter: ["data-md-color-scheme"] });
  }

  function attach(iframe) {
    if (seen.has(iframe)) return;
    seen.add(iframe);

    const onLoad = () => {
      // first sizing
      syncTheme(iframe);
      resizeToContent(iframe);

      // keep sizing as Swagger UI expands/collapses operations
      try {
        const doc = iframe.contentDocument;
        if (!doc || !doc.body) return;

        const mo = new MutationObserver(() => {
          // throttle via rAF
          requestAnimationFrame(() => resizeToContent(iframe));
        });

        mo.observe(doc.body, {
          subtree: true,
          childList: true,
          attributes: true,
        });
      } catch {
        // ignore
      }

      // Swagger UI may expand after initial load; retry a few times.
      [50, 250, 1000].forEach((ms) => setTimeout(() => resizeToContent(iframe), ms));
    };

    // if already loaded, run immediately; otherwise wait
    if (iframe.contentDocument?.readyState === "complete") onLoad();
    iframe.addEventListener("load", onLoad);
    window.addEventListener("resize", () => resizeToContent(iframe));
  }

  function init() {
    attachSchemeObserver();

    document.querySelectorAll("iframe.swagger-ui-iframe").forEach((iframe) => {
      attach(iframe);
      syncTheme(iframe);
      resizeToContent(iframe);
    });
  }

  // Provide a stable hook for mkdocs-swagger-ui-tag if instant navigation
  // prevents inline scripts from running.
  window.update_swagger_ui_iframe_height = function (id) {
    const iframe = document.getElementById(id);
    if (!iframe) return;
    resizeToContent(iframe);
  };

  // Normal load
  document.addEventListener("DOMContentLoaded", init);

  // Material instant navigation: re-run on page change
  if (window.document$ && typeof window.document$.subscribe === "function") {
    window.document$.subscribe(() => setTimeout(init, 0));
  }
})();
