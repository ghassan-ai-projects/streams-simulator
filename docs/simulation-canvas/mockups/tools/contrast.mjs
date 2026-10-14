import fs from 'node:fs';
const css = fs.readFileSync(process.argv[2], 'utf8');
const block = (sel) => { const i = css.indexOf(sel + '{'); const j = css.indexOf('}', i); return css.slice(i + sel.length + 1, j); };
const vars = (b) => Object.fromEntries([...b.matchAll(/--([\w-]+):([^;]+)/g)].map(m => [m[1], m[2].trim()]));
const dark = vars(block(':root')), light = { ...dark, ...vars(block(':root[data-theme=light]')) };
const hex = (h) => { h = h.replace('#', ''); if (h.length === 3) h = [...h].map(c => c + c).join(''); return [0, 2, 4].map(i => parseInt(h.slice(i, i + 2), 16)); };
const lin = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
const L = (rgb) => 0.2126 * lin(rgb[0]) + 0.7152 * lin(rgb[1]) + 0.0722 * lin(rgb[2]);
const ratio = (a, b) => { const x = L(a), y = L(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };
const col = (t, v) => { if (v.startsWith('#')) return hex(v); if (v.startsWith('--')) return col(t, t[v.slice(2)]); throw v; };
const mix = (fg, bg, a) => fg.map((c, i) => Math.round(c * a + bg[i] * (1 - a)));
const pairs = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
let fail = 0;
for (const [name, t] of [['dark', dark], ['light', light]]) for (const [fg, bg, min, alpha, tint] of pairs) {
  let b = col(t, bg); if (tint) b = mix(col(t, tint[0]), b, tint[1]);
  let f = col(t, fg); const a = typeof alpha === 'string' ? parseFloat(t[alpha.slice(2)]) : alpha; if (a) f = mix(f, b, a);
  const r = ratio(f, b); if (r < min) { fail++; console.log(`FAIL ${name} ${fg} on ${bg}${tint ? ' tint ' + tint : ''}: ${r.toFixed(2)} < ${min}`); }
}
console.log(fail ? `${fail} failures` : 'all pairs pass'); process.exit(fail ? 1 : 0);
