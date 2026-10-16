import { execFileSync, spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { chromium } from "@playwright/test";
import type { Browser, BrowserContext, Page } from "@playwright/test";

/**
 * Developer evidence, not a gate: capture the main screens at two desktop widths in both themes so a reviewer can
 * compare them before and after a structural change. Run with `node scripts/capture-screens.ts <output-directory>`.
 */

const screens: readonly { readonly name: string; readonly hash: string; readonly marker: string }[] = [
  { name: "workspace", hash: "#/", marker: "Factory utilities" },
  { name: "plant", hash: "#/plant?domain=domains%2Ffactory-utilities.domain.json", marker: "instances ·" },
  { name: "editor", hash: "#/edit?composition=compositions%2Ffactory-utilities.composition.json", marker: "Save" },
  { name: "review", hash: "#/run?world=w-1", marker: "observations" },
  { name: "launch", hash: "#/plant?domain=domains%2Ffactory-utilities.domain.json&run=1", marker: "Start the run" },
];

const widths = [1440, 1920] as const;
const themes = ["light", "dark"] as const;
const ready = /canvas: http:\/\/(\S+)\ncanvas: pairing code (\S+)\n/;
const repository = resolve(import.meta.dirname, "..", "..");

interface Served {
  readonly url: string;
  readonly code: string;
  readonly child: ChildProcess;
}

function binaryPath(): string {
  const path = process.env.STREAMSIM_BIN ?? join(repository, "bin", "streams-simulator");
  if (!existsSync(path)) throw new Error(`no streamsim binary at ${path}: run make build`);
  return path;
}

function buildWorkspace(root: string): string {
  const dir = join(root, "workspace");
  for (const folder of ["domains", "compositions", "runs"]) mkdirSync(join(dir, folder), { recursive: true });
  const copy = (from: string, to: string): void => {
    copyFileSync(join(repository, from), join(dir, to));
  };
  copy("examples/factory-utilities.domain.json", "domains/factory-utilities.domain.json");
  copy("examples/compositions/factory-utilities.composition.json", "compositions/factory-utilities.composition.json");
  execFileSync(
    binaryPath(),
    [
      "run",
      "-domains-dir",
      join(dir, "domains"),
      "-domain",
      "factory-utilities",
      "-seed",
      "7",
      "-duration",
      "1800",
      "-out",
      join(dir, "runs", "w-1"),
    ],
    { cwd: repository, stdio: ["ignore", "ignore", "pipe"] },
  );
  return dir;
}

function serve(dir: string): Promise<Served> {
  const child = spawn(
    binaryPath(),
    ["canvas", "--workspace", dir, "--ui-dir", join(repository, "canvas", "dist"), "--addr", "127.0.0.1:0"],
    { cwd: repository, stdio: ["ignore", "ignore", "pipe"] },
  );
  return new Promise((settle, fail) => {
    let output = "";
    child.stderr.on("data", (chunk: Buffer) => {
      output += chunk.toString();
      const found = ready.exec(output);
      if (found?.[1] !== undefined && found[2] !== undefined)
        settle({ url: `http://${found[1]}`, code: found[2], child });
    });
    child.on("exit", (status) => {
      fail(new Error(`the canvas bridge exited with ${String(status)}: ${output}`));
    });
  });
}

async function pairingCookie(url: string, code: string): Promise<{ name: string; value: string; domain: string }> {
  const answer = await fetch(`${url}/api/session`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
  });
  const [pair = ""] = (answer.headers.get("set-cookie") ?? "").split(";");
  const [name = "", ...value] = pair.split("=");
  if (!answer.ok) throw new Error(`pairing failed with ${String(answer.status)}`);
  return { name, value: value.join("="), domain: new URL(url).hostname };
}

async function paintTheme(page: Page, expected: string): Promise<void> {
  for (let press = 0; press < 3; press += 1) {
    const theme = await page.evaluate(() => document.documentElement.dataset.theme ?? "");
    if (theme === expected) return;
    await page.getByRole("button", { name: /theme/i }).click();
  }
}

async function settle(page: Page, marker: string): Promise<void> {
  await page.getByText(marker, { exact: false }).first().waitFor({ timeout: 15_000 });
  await page.evaluate(async () => {
    await document.fonts.ready;
  });
}

async function shoot(context: BrowserContext, app: Served, out: string): Promise<void> {
  for (const width of widths) {
    const page = await context.newPage();
    await page.setViewportSize({ width, height: width === 1920 ? 1080 : 900 });
    for (const theme of themes) {
      for (const screen of screens) {
        await page.goto(`${app.url}/${screen.hash}`);
        await settle(page, screen.marker);
        await paintTheme(page, theme);
        await page.screenshot({ path: join(out, `${screen.name}-${String(width)}-${theme}.png`) });
      }
    }
    await page.close();
  }
}

async function main(): Promise<void> {
  const out = resolve(process.argv[2] ?? "/tmp/canvas-capture");
  mkdirSync(out, { recursive: true });
  const root = mkdtempSync(join(tmpdir(), "canvas-capture-"));
  const app = await serve(buildWorkspace(root));
  const browser: Browser = await chromium.launch({ channel: "chrome" });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const cookie = await pairingCookie(app.url, app.code);
  await context.addCookies([{ ...cookie, path: "/", httpOnly: true, secure: false, sameSite: "Strict" }]);
  try {
    await shoot(context, app, out);
  } finally {
    await browser.close();
    app.child.kill();
    rmSync(root, { recursive: true, force: true });
  }
  process.stdout.write(`captured ${String(screens.length * widths.length * themes.length)} screens into ${out}\n`);
}

await main();
