// Initialize Mermaid diagrams for MkDocs Material.
// This is intentionally tiny and dependency-free.

function renderMermaid() {
  const mermaid = window.mermaid;
  if (!mermaid) return;

  mermaid.initialize({ startOnLoad: false });

  const nodes = document.querySelectorAll('.mermaid');
  if (!nodes.length) return;

  if (typeof mermaid.run === 'function') {
    mermaid.run({ nodes });
    return;
  }

  if (typeof mermaid.init === 'function') {
    mermaid.init(undefined, nodes);
  }
}

if (window.document$?.subscribe) {
  window.document$.subscribe(renderMermaid);
} else {
  window.addEventListener('load', renderMermaid);
}
