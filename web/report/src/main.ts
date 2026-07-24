import { renderComparison } from './comparison.ts';
import { renderFindings } from './findings.ts';
const dataset = JSON.parse(document.getElementById('report-data').textContent);
if (dataset.version !== '1' || !dataset.comparison) throw new Error('Unsupported report dataset');
document.getElementById('title').textContent = dataset.title;
document.getElementById('metadata').textContent = `${dataset.skillAlias} | ${dataset.generatedAt} | ${dataset.freshness}`;
renderComparison(document.getElementById('comparison'), dataset.comparison);
renderFindings(document.getElementById('findings'), dataset.comparison);
