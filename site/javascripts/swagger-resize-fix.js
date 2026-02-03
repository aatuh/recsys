(function () {
  function fix() {
    // Nudge layout-dependent widgets
    window.dispatchEvent(new Event("resize"));

    // Optional: if your Swagger iframe still ends up tiny,
    // force a sane minimum height:
    document.querySelectorAll("iframe").forEach((el) => {
      const src = (el.getAttribute("src") || "").toLowerCase();
      if (src.includes("swagger") || src.includes("openapi")) {
        el.style.width = "100%";
        el.style.minHeight = "75vh";
      }
    });
  }

  // Normal full load
  document.addEventListener("DOMContentLoaded", fix);

  // Material instant navigation
  if (window.document$ && typeof window.document$.subscribe === "function") {
    window.document$.subscribe(fix);
  }
})();
