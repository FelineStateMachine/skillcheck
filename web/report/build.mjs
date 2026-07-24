import { mkdir, readFile, writeFile } from 'node:fs/promises';
const files = ['accessibility.ts','workflow-svg.ts','comparison.ts','findings.ts','state.ts','main.ts'];
let bundle = '';
for (const file of files) bundle += (await readFile(new URL(`src/${file}`, import.meta.url), 'utf8')).replace(/^import .*;$/gm, '').replace(/^export /gm, '') + '\n';
await mkdir(new URL('../../internal/report/assets/', import.meta.url), { recursive:true });
await writeFile(new URL('../../internal/report/assets/report.js', import.meta.url), bundle);
await writeFile(new URL('../../internal/report/assets/report.css', import.meta.url), await readFile(new URL('styles/report.css', import.meta.url)));
