import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("the factory opens from the workspace as a typed plant graph", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "Open Factory utilities" }).click();
  const graph = page.getByRole("region", { name: "Plant factory-utilities" });
  await expect(graph.getByText("23 instances · 59 links")).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
  await expect(page.locator(".react-flow__edge").first()).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});

test("selecting an instance in the outline shows its kind in the inspector", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT-1" }).click();
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await expect(inspector.getByRole("heading", { name: "CT-1" })).toBeVisible();
  await expect(inspector.getByRole("region", { name: "Channels" })).toContainText("basin temp · °C");
  await expect(inspector.getByRole("region", { name: "Links in" })).toContainText("CH-1 heat out → heat in");
  await expect(inspector.getByRole("region", { name: "Faults" })).toContainText("fan failure");
  await page.getByRole("button", { name: "Workspace", exact: true }).click();
  await expect(page.getByRole("heading", { name: "lab-workspace" })).toBeVisible();
});

test("a shared campus link navigates to its declared receiver in another hall", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Fcampus-3.domain.json`);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CAL-1" }).click();
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await inspector.getByText("Browse outgoing links (18)").click();
  await inspector.getByRole("button", { name: /Select receiver h02\/M-1; CAL-1 shift out/ }).click();
  await expect(page.getByRole("heading", { name: "h02", level: 1 })).toBeVisible();
  await expect(inspector.getByRole("heading", { name: "h02/M-1", level: 2 })).toBeVisible();
});

test("the graph works from the keyboard: arrows move, Enter selects, Escape clears, ? lists the keys", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page).toHaveTitle("Factory utilities · plant · Streams Simulator");
  const graph = page.getByRole("application", { name: "Plant graph" });
  await graph.focus();
  const zoom = await page.getByText(/^Zoom \d+ %$/).textContent();
  await page.keyboard.press("ControlOrMeta+Equal");
  await expect(page.getByText(/^Zoom \d+ %$/)).toHaveText(zoom ?? "");
  await page.keyboard.press("ArrowRight");
  await page.keyboard.press("Enter");
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await expect(inspector.getByRole("heading", { level: 2 })).toBeVisible();
  await expect(page.locator("#graph-said")).toContainText("selected");
  await page.keyboard.press("Escape");
  await expect(inspector).toContainText("Select an instance");
  await page.keyboard.press("?");
  await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeHidden();
});

function overlapsAndSpills(): string[] {
  const overlap = (a: DOMRect, b: DOMRect) =>
    Math.min(a.right, b.right) - Math.max(a.left, b.left) > 1 &&
    Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top) > 1;
  const outside = (box: DOMRect, frame: DOMRect) =>
    box.left < frame.left - 1 || box.right > frame.right + 1 || box.bottom > frame.bottom + 1;
  const shown = (element: HTMLElement, box: DOMRect) =>
    box.width > 0 && box.height > 0 && getComputedStyle(element).visibility !== "hidden";
  const spills = (node: HTMLElement, frame: DOMRect) =>
    [...node.querySelectorAll<HTMLElement>("*")].some((child) => {
      const box = child.getBoundingClientRect();
      return child.closest(".react-flow__handle") === null && shown(child, box) && outside(box, frame);
    });
  const nodes = [...document.querySelectorAll<HTMLElement>(".react-flow__node:not(.react-flow__node-frame)")].map(
    (node) => ({
      id: node.dataset.id ?? "",
      box: node.getBoundingClientRect(),
      node,
    }),
  );
  return nodes.flatMap((a, index) => [
    ...nodes.slice(index + 1).flatMap((b) => (overlap(a.box, b.box) ? [`${a.id} overlaps ${b.id}`] : [])),
    ...(spills(a.node, a.box) ? [`${a.id} spills`] : []),
  ]);
}

function settledZoom(): Promise<string> {
  const read = () =>
    getComputedStyle(document.querySelector("[aria-roledescription='plant graph']") ?? document.body).getPropertyValue(
      "--zoom",
    );
  const before = read();
  return new Promise((resolve) => {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        resolve(before === read() ? before : "");
      });
    });
  });
}

test("no node overlaps another or spills its text at any zoom tier", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await graph.focus();
  await page.keyboard.press("0");
  const tiers = new Set<string>();
  let zoom = "";
  for (let step = 0; step < 8; step += 1) {
    const previous = zoom;
    await expect.poll(async () => (zoom = await page.evaluate(settledZoom)) !== "" && zoom !== previous).toBe(true);
    tiers.add((await graph.getAttribute("class"))?.match(/tier-(\w+)/)?.[1] ?? "");
    expect(await page.evaluate(overlapsAndSpills)).toEqual([]);
    await page.keyboard.press("+");
  }
  expect([...tiers]).toEqual(expect.arrayContaining(["overview", "compact", "full"]));
});

function activeRingOnScreen(): number {
  const graph = document.querySelector<HTMLElement>("[aria-roledescription='plant graph']");
  const ring = graph?.querySelector<HTMLElement>("[role=group][class*='active']");
  if (!graph || !ring) return 0;
  const zoom = Number(graph.style.getPropertyValue("--zoom"));
  return parseFloat(getComputedStyle(ring).outlineWidth) * zoom;
}

test("the keyboard ring stays 2 px on screen when zoomed out", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await page.getByRole("application", { name: "Plant graph" }).focus();
  await page.keyboard.press("ArrowRight");
  await expect.poll(() => page.evaluate(activeRingOnScreen)).toBeGreaterThanOrEqual(1.9);
  await expect
    .poll(async () => {
      await page.keyboard.press("-");
      return page.getByRole("application", { name: "Plant graph" }).getAttribute("class");
    })
    .toMatch(/tier-map/);
  await expect.poll(() => page.evaluate(activeRingOnScreen)).toBeGreaterThanOrEqual(1.9);
});

test("Enter opens a hall and the arrows keep working inside it; Escape returns to the hall", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Fcampus-3.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  const said = page.locator("#graph-said");
  await expect(page.getByRole("group", { name: /^h01,/ })).toBeVisible();
  await graph.focus();
  await expect
    .poll(async () => {
      await page.keyboard.press("ArrowDown");
      return said.textContent();
    })
    .toMatch(/^h0/);
  const hall = ((await said.textContent()) ?? "").split(",")[0] ?? "";
  await page.keyboard.press("Enter");
  await expect(page).toHaveTitle(new RegExp(`^${hall} · `));
  await expect(graph).toBeFocused();
  await page.keyboard.press("ArrowRight");
  await expect(said).toContainText("inputs");
  await page.keyboard.press("Escape");
  await expect(page).toHaveTitle(/· plant ·/);
  await expect(graph).toBeFocused();
  await expect(said).toContainText(`${hall},`);
});

function insideGraph(id: string): boolean {
  const graph = document.querySelector("[aria-roledescription='plant graph']")?.getBoundingClientRect();
  const node = document.querySelector(`.react-flow__node[data-id='${id}']`)?.getBoundingClientRect();
  if (!graph || !node) return false;
  return node.left >= graph.left && node.right <= graph.right && node.top >= graph.top && node.bottom <= graph.bottom;
}

function dimContrast(): number {
  const luminance = (rgb: string) => {
    const [r = 0, g = 0, b = 0] = (rgb.match(/\d+/g) ?? []).map(Number).map((c) => {
      const s = c / 255;
      return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
    });
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  };
  const path = document.querySelector(".react-flow__edge.wire-off .react-flow__edge-path");
  const graph = document.querySelector("[aria-roledescription='plant graph']");
  if (!path || !graph) return 0;
  const [a, b] = [luminance(getComputedStyle(path).stroke), luminance(getComputedStyle(graph).backgroundColor)];
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
}

test("a selection made outside the graph comes into view, and arrow travel keeps the active node on screen", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await graph.focus();
  await expect
    .poll(async () => {
      await page.keyboard.press("+");
      return graph.getAttribute("class");
    })
    .toMatch(/tier-full/);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "MB-PROC" }).click();
  await expect.poll(() => page.evaluate(insideGraph, "MB-PROC")).toBe(true);
  await graph.focus();
  for (const key of ["ArrowLeft", "ArrowUp", "ArrowLeft", "ArrowUp"]) {
    await page.keyboard.press(key);
    const said = (await page.locator("#graph-said").textContent()) ?? "";
    await expect.poll(() => page.evaluate(insideGraph, said.split(",")[0] ?? "")).toBe(true);
  }
  await page.keyboard.press("0");
  await expect.poll(() => page.evaluate(dimContrast)).toBeGreaterThanOrEqual(3);
});
