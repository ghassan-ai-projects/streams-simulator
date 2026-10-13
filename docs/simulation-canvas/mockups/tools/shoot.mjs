import { chromium } from "../../../../canvas/node_modules/playwright/index.mjs";
const [, , dir, out, ...pages] = process.argv;
const browser = await chromium.launch({ channel: "chrome" });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1 });
for (const name of pages) {
  for (const theme of ["light", "dark"]) {
    await page.goto(`file://${dir}/${name}.html`);
    await page.evaluate((t) => { document.documentElement.dataset.theme = t; }, theme);
    await page.screenshot({ path: `${out}/${name}-${theme}.png` });
  }
}
await browser.close();
