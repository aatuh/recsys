// Initialize Mermaid diagrams for MkDocs Material.
// This is intentionally tiny and dependency-free.

window.addEventListener('load', () => {
  if (window.mermaid) {
    window.mermaid.initialize({ startOnLoad: true });
  }
});
