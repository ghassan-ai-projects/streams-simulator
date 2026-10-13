import { chromium } from "../../../../canvas/node_modules/playwright/index.mjs";
const [, , dir, ...pages] = process.argv;
const browser = await chromium.launch({ channel: "chrome" });
let total = 0;
for (const size of [[1440, 900], [1920, 1080]]) {
  const page = await browser.newPage({ viewport: { width: size[0], height: size[1] } });
  for (const name of pages) {
    await page.goto(`file://${dir}/${name}.html`);
    const issues = await page.evaluate(() => {
      const out = [];
      const label = (el) => (el.getAttribute("aria-label") || el.textContent || "").trim().slice(0, 50);
      for (const g of document.querySelectorAll(".graph")) {
        const gr = g.getBoundingClientRect();
        const nodes = [...g.querySelectorAll(".node:not([aria-hidden])")].map((n) => [n, n.getBoundingClientRect()]);
        for (const [n, r] of nodes) if (r.bottom > gr.bottom + 0.5 || r.right > gr.right + 0.5 || r.top < gr.top || r.left < gr.left) out.push(`node outside graph: ${label(n.querySelector(".nh") || n)}`);
        for (let i = 0; i < nodes.length; i++) for (let j = i + 1; j < nodes.length; j++) {
          const [a, ra] = nodes[i], [b, rb] = nodes[j];
          if (ra.left < rb.right + 8 && rb.left < ra.right + 8 && ra.top < rb.bottom + 8 && rb.top < ra.bottom + 8) out.push(`nodes overlap: ${label(a.querySelector(".nh"))} / ${label(b.querySelector(".nh"))}`);
        }
      }
      for (const g of document.querySelectorAll(".graph")) {
        const blocks = [...g.querySelectorAll(".node:not([aria-hidden]), .frame .ft")].map((n) => [n, n.getBoundingClientRect()]);
        for (const o of g.querySelectorAll(":scope > .badge, :scope > .tip, :scope > .portal, :scope > .sh")) {
          const r = o.getBoundingClientRect();
          for (const [n, b] of blocks) if (r.left < b.right && b.left < r.right && r.top < b.bottom && b.top < r.bottom) out.push(`overlay on content: "${o.textContent.trim().slice(0, 30)}" over "${n.textContent.trim().slice(0, 20)}"`);
        }
        for (const t of g.querySelectorAll("svg text")) {
          const r = t.getBoundingClientRect(), s = t.closest("svg").getBoundingClientRect();
          if (r.right > s.right + 1 || r.left < s.left - 1) out.push(`svg text outside its chart: "${t.textContent.slice(0, 30)}"`);
        }
      }
      for (const t of document.querySelectorAll("aside svg text, section svg text")) {
        const r = t.getBoundingClientRect(), s = t.closest("svg").getBoundingClientRect();
        if (r.right > s.right + 1 || r.left < s.left - 1 || r.top < s.top - 1 || r.bottom > s.bottom + 1) out.push(`svg text outside its chart: "${t.textContent.slice(0, 30)}"`);
      }
      for (const el of document.querySelectorAll("body *")) {
        if ((el.closest("svg") && !el.closest("text")) || el.classList.contains("k") || el.classList.contains("sr-only") || el.closest("[aria-hidden=true]")) continue;
        const cs = getComputedStyle(el);
        if (cs.display === "none" || cs.visibility === "hidden") continue;
        const own = [...el.childNodes].some((c) => c.nodeType === 3 && c.textContent.trim());
        if (!own) continue;
        if (el.scrollWidth > el.clientWidth + 1 && cs.overflow !== "visible" && !(cs.textOverflow === "ellipsis" && el.closest("[title]"))) out.push(`clipped text: "${el.textContent.trim().slice(0, 40)}"`);
        const r = el.getRange ? null : el.getBoundingClientRect();
        let p = el.parentElement;
        while (p && p !== document.body) {
          const ps = getComputedStyle(p);
          if (ps.overflow !== "visible" || ps.overflowX !== "visible") {
            const pr = p.getBoundingClientRect();
            if (r.width && (r.right > pr.right + 1 || (r.bottom > pr.bottom + 1 && !p.classList.contains("scroll"))) && (!p.classList.contains("graph") || el.classList.contains("tip"))) out.push(`cut by container: "${el.textContent.trim().slice(0, 40)}"`);
            break;
          }
          p = p.parentElement;
        }
        if (cs.whiteSpace !== "nowrap" && el.getClientRects().length === 1 && cs.height !== "auto") {
          const lh = parseFloat(cs.lineHeight) || parseFloat(cs.fontSize) * 1.4;
          if (el.clientHeight && el.scrollHeight > el.clientHeight + 1 && el.scrollHeight > lh * 1.5 && cs.overflow !== "visible") out.push(`wrapped in a fixed box: "${el.textContent.trim().slice(0, 40)}"`);
        }
        const fs = parseFloat(cs.fontSize);
        if (fs < 11) out.push(`font ${fs}px: "${el.textContent.trim().slice(0, 30)}"`);
      }
      return [...new Set(out)];
    });
    if (issues.length) { total += issues.length; console.log(`== ${name} ${size.join("x")}`); for (const i of issues) console.log("  " + i); }
  }
  await page.close();
}
await browser.close();
console.log(total ? `${total} layout issues` : "layout clean");
process.exit(total ? 1 : 0);
