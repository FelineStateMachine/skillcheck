import { workflowSVG } from './workflow-svg.ts';
export function renderComparison(root, comparison) {
  const sides = [['left', comparison.left], ['right', comparison.right]];
  root.innerHTML = sides.map(([name, cohort]) => `<section><h2>${name === 'left' ? 'Cohort A' : 'Cohort B'}</h2><p>${cohort.members.length} uses</p>${workflowSVG(cohort.graph, `${name} workflow`)}</section>`).join('');
}
