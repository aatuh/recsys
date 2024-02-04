(function () {
  function fix() {
    // Nudge layout-dependent widgets
    window.dispatchEvent(new Event("resize"));

    document.querySelectorAll("iframe.swagger-ui-iframe").forEach((el) => {
      el.style.display = "block";
      el.style.width = "100%";
      el.style.minHeight = "75vh";
      el.style.border = "0";

      const id = el.getAttribute("id");
      if (!id) return;
      if (typeof window.update_swagger_ui_iframe_height !== "function") return;
      try {
        window.update_swagger_ui_iframe_height(id);
      } catch {
        // ignore
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
