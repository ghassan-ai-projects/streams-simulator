import type { Audit } from "./audit";

/** What a user can see painting under the run plan's footer, clipped by every scrolling ancestor. */
export function underRunFooter(atLeast = 1): Audit {
  const fields =
    "aside button, aside p, aside small, aside select, aside input, aside legend, aside summary, aside label";
  const start = [...document.querySelectorAll<HTMLElement>("aside button")].find(
    (button) => button.textContent.trim() === "Start the run",
  );
  const footer = start?.parentElement ?? null;
  if (footer === null) return { considered: 0, problems: ["no run plan footer"], met: false };
  const clipped = (node: HTMLElement): DOMRect | null => {
    let rect = node.getBoundingClientRect();
    let parent = node.parentElement;
    while (parent !== null) {
      const style = getComputedStyle(parent);
      if ([style.overflowX, style.overflowY].some((value) => value !== "visible")) {
        const box = parent.getBoundingClientRect();
        const left = Math.max(rect.left, box.left);
        const top = Math.max(rect.top, box.top);
        const right = Math.min(rect.right, box.right);
        const bottom = Math.min(rect.bottom, box.bottom);
        if (right - left <= 0 || bottom - top <= 0) return null;
        rect = new DOMRect(left, top, right - left, bottom - top);
      }
      parent = parent.parentElement;
    }
    return rect;
  };
  const box = footer.getBoundingClientRect();
  const found: string[] = [];
  const nodes = [...document.querySelectorAll<HTMLElement>(fields)].filter((node) => !footer.contains(node));
  nodes.forEach((node) => {
    const rect = clipped(node);
    if (rect === null || rect.bottom - box.top <= 1 || box.bottom - rect.top <= 1) return;
    found.push(`${node.tagName.toLowerCase()}: ${node.textContent.trim().slice(0, 48)}`);
  });
  return { considered: nodes.length, problems: found, met: nodes.length >= atLeast };
}
