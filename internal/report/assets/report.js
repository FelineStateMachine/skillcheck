function activateOnKeyboard(event, activate) {
  if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); activate(); }
}

function workflowSVG(graph, label) {
  const nodes = graph.nodes || [];
  const rows = nodes.map((node, i) => `<g id="node-${safe(node.id)}" transform="translate(24 ${32 + i * 48})"><rect width="300" height="34" rx="4"/><text x="12" y="22">${text(node.action)} (${node.count})</text></g>`).join('');
  return `<svg role="img" aria-label="${text(label)}" viewBox="0 0 348 ${Math.max(80, nodes.length * 48 + 32)}">${rows}</svg>`;
}
function safe(value) { return String(value).replace(/[^a-zA-Z0-9_-]/g, '-'); }
function text(value) { const node = document.createElement('span'); node.textContent = String(value); return node.innerHTML; }


function renderComparison(root, comparison) {
  const sides = [['left', comparison.left], ['right', comparison.right]];
  root.innerHTML = sides.map(([name, cohort]) => `<section><h2>${name === 'left' ? 'Cohort A' : 'Cohort B'}</h2><p>${cohort.members.length} uses</p>${workflowSVG(cohort.graph, `${name} workflow`)}</section>`).join('');
}

function renderFindings(root, comparison) {
  const differences = comparison.differences || [];
  root.innerHTML = differences.length ? differences.map(d => `<li><button data-difference="${d.id}">${d.action}: ${d.detail}</button></li>`).join('') : '<li>No supportable differences.</li>';
}

const initialState = { selectedDifference: '', selectedSide: 'left' };
function reduce(state, action) {
  if (action.type === 'difference') return { ...state, selectedDifference: action.id };
  if (action.type === 'side') return { ...state, selectedSide: action.side };
  return state;
}



const dataset = JSON.parse(document.getElementById('report-data').textContent);
if (dataset.version !== '1' || !dataset.comparison) throw new Error('Unsupported report dataset');
document.getElementById('title').textContent = dataset.title;
document.getElementById('metadata').textContent = `${dataset.skillAlias} | ${dataset.generatedAt} | ${dataset.freshness}`;
renderComparison(document.getElementById('comparison'), dataset.comparison);
renderFindings(document.getElementById('findings'), dataset.comparison);

