import fs from 'node:fs';
const css = fs.readFileSync(process.argv[2], 'utf8');
const block = (sel) => { const i = css.indexOf(sel + '{'); const j = css.indexOf('}', i); return css.slice(i + sel.length + 1, j); };
const vars = (b) => Object.fromEntries([...b.matchAll(/--([\w-]+):([^;]+)/g)].map(m => [m[1], m[2].trim()]));
const dark = vars(block(':root')), light = { ...dark, ...vars(block(':root[data-theme=light]')) };
const hex = (h) => { h = h.replace('#', ''); return [0, 2, 4].map(i => parseInt(h.slice(i, i + 2), 16)); };
const lab = ([r, g, b]) => {
  const f = (c) => { c /= 255; return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; };
  const [R, G, B] = [f(r), f(g), f(b)];
  const X = (R * 0.4124 + G * 0.3576 + B * 0.1805) / 0.95047, Y = R * 0.2126 + G * 0.7152 + B * 0.0722, Z = (R * 0.0193 + G * 0.1192 + B * 0.9505) / 1.08883;
  const h = (t) => t > 0.008856 ? Math.cbrt(t) : 7.787 * t + 16 / 116;
  return [116 * h(Y) - 16, 500 * (h(X) - h(Y)), 200 * (h(Y) - h(Z))];
};
const de2000 = (l1, l2) => {
  const [L1, a1, b1] = l1, [L2, a2, b2] = l2, rad = Math.PI / 180;
  const C1 = Math.hypot(a1, b1), C2 = Math.hypot(a2, b2), Cb = (C1 + C2) / 2;
  const G = 0.5 * (1 - Math.sqrt(Cb ** 7 / (Cb ** 7 + 25 ** 7)));
  const a1p = a1 * (1 + G), a2p = a2 * (1 + G), C1p = Math.hypot(a1p, b1), C2p = Math.hypot(a2p, b2);
  const hp = (a, b) => { const x = Math.atan2(b, a) / rad; return x < 0 ? x + 360 : x; };
  const h1p = hp(a1p, b1), h2p = hp(a2p, b2);
  const dLp = L2 - L1, dCp = C2p - C1p;
  let dh = h2p - h1p; if (C1p * C2p === 0) dh = 0; else if (dh > 180) dh -= 360; else if (dh < -180) dh += 360;
  const dHp = 2 * Math.sqrt(C1p * C2p) * Math.sin(dh / 2 * rad);
  const Lbp = (L1 + L2) / 2, Cbp = (C1p + C2p) / 2;
  let hbp = h1p + h2p; if (C1p * C2p !== 0) { if (Math.abs(h1p - h2p) > 180) hbp += hbp < 360 ? 360 : -360; hbp /= 2; }
  const T = 1 - 0.17 * Math.cos((hbp - 30) * rad) + 0.24 * Math.cos(2 * hbp * rad) + 0.32 * Math.cos((3 * hbp + 6) * rad) - 0.2 * Math.cos((4 * hbp - 63) * rad);
  const dTh = 30 * Math.exp(-(((hbp - 275) / 25) ** 2)), Rc = 2 * Math.sqrt(Cbp ** 7 / (Cbp ** 7 + 25 ** 7));
  const Sl = 1 + 0.015 * (Lbp - 50) ** 2 / Math.sqrt(20 + (Lbp - 50) ** 2), Sc = 1 + 0.045 * Cbp, Sh = 1 + 0.015 * Cbp * T, Rt = -Math.sin(2 * dTh * rad) * Rc;
  return Math.sqrt((dLp / Sl) ** 2 + (dCp / Sc) ** 2 + (dHp / Sh) ** 2 + Rt * (dCp / Sc) * (dHp / Sh));
};
const ports = ['air', 'pressure', 'water', 'cw', 'temp', 'heat', 'power', 'signal', 'env'].map(p => 't-' + p);
const reserved = ['alarm', 'warn', 'truth', 'accent'];
const minPort = +process.argv[3] || 12, minRes = +process.argv[4] || 20;
let fail = 0;
for (const [name, t] of [['dark', dark], ['light', light]]) {
  const c = (k) => lab(hex(t[k]));
  for (let i = 0; i < ports.length; i++) for (let j = i + 1; j < ports.length; j++) {
    const d = de2000(c(ports[i]), c(ports[j])); if (d < minPort) { fail++; console.log(`FAIL ${name} ${ports[i]} ~ ${ports[j]} ΔE ${d.toFixed(1)}`); }
  }
  for (const p of ports) for (const r of reserved) {
    const d = de2000(c(p), c(r)); if (d < minRes) { fail++; console.log(`FAIL ${name} ${p} ~ ${r} ΔE ${d.toFixed(1)}`); }
  }
}
console.log(fail ? `${fail} failures` : 'palette separated'); process.exit(fail ? 1 : 0);
