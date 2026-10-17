import type { Page } from "@playwright/test";

export interface Sample {
  readonly at: number;
  readonly height: number;
}

export interface CanvasProbe {
  readonly samples: readonly Sample[];
  readonly elapsed: number;
}

interface Probe {
  shifts: number;
  samples: Sample[];
  elapsed: number;
}

declare global {
  interface Window {
    __layout?: Probe;
  }
}

function install(): void {
  const probe: Probe = { shifts: 0, samples: [], elapsed: -1 };
  window.__layout = probe;
  new PerformanceObserver((list) => {
    for (const entry of list.getEntries()) {
      const shift = entry as PerformanceEntry & { value: number; hadRecentInput: boolean };
      if (!shift.hadRecentInput) probe.shifts += shift.value;
    }
  }).observe({ type: "layout-shift", buffered: true });
  let started = -1;
  const frame = (now: number): void => {
    const graph = document.querySelector("[aria-roledescription='plant graph']");
    if (graph !== null) {
      started = started < 0 ? now : started;
      probe.elapsed = Math.round(now - started);
      const height = Math.round(graph.getBoundingClientRect().height);
      if (probe.samples.at(-1)?.height !== height) probe.samples.push({ at: probe.elapsed, height });
    }
    requestAnimationFrame(frame);
  };
  requestAnimationFrame(frame);
}

export async function watchLayout(page: Page): Promise<void> {
  await page.addInitScript(install);
}

export function layoutShift(page: Page): Promise<number> {
  return page.evaluate(() => window.__layout?.shifts ?? -1);
}

export function canvasProbe(page: Page): Promise<CanvasProbe> {
  return page.evaluate(() => ({
    samples: window.__layout?.samples ?? [],
    elapsed: window.__layout?.elapsed ?? -1,
  }));
}

export function heightAt(samples: readonly Sample[], ms: number): number | null {
  return samples.filter((sample) => sample.at <= ms).at(-1)?.height ?? null;
}

interface Box {
  readonly label: string;
  readonly left: number;
  readonly right: number;
  readonly top: number;
  readonly bottom: number;
}

export interface GraphBoxes {
  readonly host: Box | null;
  readonly notices: readonly Box[];
  readonly targets: readonly Box[];
}

function measure(): GraphBoxes {
  const boxes = (selector: string, exclude: string): Box[] =>
    [...document.querySelectorAll(selector)]
      .filter((element) => element.closest(exclude) === null)
      .map((element) => {
        const { left, right, top, bottom } = element.getBoundingClientRect();
        return { label: element.textContent.slice(0, 40), left, right, top, bottom };
      });
  return {
    host: boxes("[aria-roledescription='plant graph']", "never")[0] ?? null,
    notices: boxes("[role='alert'], [role='status'][title]", "[role='complementary']"),
    targets: boxes("[data-part='frame-title'], .react-flow__node", "never"),
  };
}

export function graphBoxes(page: Page): Promise<GraphBoxes> {
  return page.evaluate(measure);
}

function meets(a: Box, b: Box): boolean {
  return a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;
}

export function coveredBy(graph: GraphBoxes): string[] {
  const host = graph.host;
  const visible = graph.targets.filter((target) => host !== null && target.top >= host.top && target.top < host.bottom);
  return graph.notices.flatMap((notice) =>
    visible.filter((target) => meets(notice, target)).map((target) => `${notice.label} covers ${target.label}`),
  );
}
