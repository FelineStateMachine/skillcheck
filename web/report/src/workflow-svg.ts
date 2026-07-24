export function workflowSVG(graph, label) {
  const nodes = graph.nodes || [];
  const rows = nodes.map((node, i) => `<g id="node-${safe(node.id)}" transform="translate(24 ${32 + i * 48})"><rect width="300" height="34" rx="4"/><text x="12" y="22">${text(node.action)} (${node.count})</text></g>`).join('');
  return `<svg role="img" aria-label="${text(label)}" viewBox="0 0 348 ${Math.max(80, nodes.length * 48 + 32)}">${rows}</svg>`;
}
function safe(value) { return String(value).replace(/[^a-zA-Z0-9_-]/g, '-'); }
function text(value) { const node = document.createElement('span'); node.textContent = String(value); return node.innerHTML; }
