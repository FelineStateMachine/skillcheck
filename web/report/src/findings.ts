export function renderFindings(root, comparison) {
  const differences = comparison.differences || [];
  root.innerHTML = differences.length ? differences.map(d => `<li><button data-difference="${d.id}">${d.action}: ${d.detail}</button></li>`).join('') : '<li>No supportable differences.</li>';
}
