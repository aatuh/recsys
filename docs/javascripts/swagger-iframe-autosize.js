(function () {
  const seen = new WeakSet();

  function isSwaggerFrame(iframe) {
    try {
      const doc = iframe.contentDocument;
      if (!doc) return false;
      return !!doc.querySelector(".swagger-ui");
    } catch {
      return false; // cross-origin or not ready
    }
  }

  function resizeToContent(iframe) {
    try {
      const doc = iframe.contentDocument;
      if (!doc) return;

      // Ensure layout has settled
      const h = Math.max(
        doc.documentElement.scrollHeight,
        doc.body ? doc.body.scrollHeight : 0,
      );

      // A little padding to avoid 1px scrollbars from rounding
      iframe.style.height = h + 8 + "px";
      iframe.style.width = "100%";
      iframe.style.border = "0";
      iframe.style.overflow = "hidden";
    } catch {
      // If cross-origin, you can't introspect height; fallback:
      iframe.style.height = "85vh";
    }
  }

  function attach(iframe) {
    if (seen.has(iframe)) return;
    seen.add(iframe);

    const onLoad = () => {
      // first sizing
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
    };

    // if already loaded, run immediately; otherwise wait
    if (iframe.contentDocument?.readyState === "complete") onLoad();
    iframe.addEventListener("load", onLoad);
    window.addEventListener("resize", () => resizeToContent(iframe));
  }

  function init() {
    // Only attach to frames that actually host Swagger UI
    document.querySelectorAll("iframe").forEach((iframe) => {
      // Attach first (so load handler is there)
      attach(iframe);

      // If same-origin and Swagger UI already there, size now
      if (isSwaggerFrame(iframe)) resizeToContent(iframe);
    });
  }

  // Normal load
  document.addEventListener("DOMContentLoaded", init);

  // Material instant navigation: re-run on page change
  if (window.document$ && typeof window.document$.subscribe === "function") {
    window.document$.subscribe(() => setTimeout(init, 0));
  }
})();
