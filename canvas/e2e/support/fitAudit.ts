import type { Locator } from "@playwright/test";

export async function headingInside(frame: Locator, name: string): Promise<boolean> {
  const outer = await frame.boundingBox();
  // The heading's box is the panel's content box even when nowrap text runs past it,
  // so the check measures the painted text runs, not the element box.
  const painted = await frame.getByText(name).evaluate((element) => {
    const range = document.createRange();
    range.selectNodeContents(element);
    return [...range.getClientRects()].reduce((right, rect) => Math.max(right, rect.right), 0);
  });
  return outer !== null && painted > 0 && painted <= outer.x + outer.width - 8;
}

export function nodesOutsideCanvas(): string[] {
  const canvas = document.querySelector(".react-flow")?.getBoundingClientRect();
  if (canvas === undefined) return ["no canvas"];
  const outside: string[] = [];
  document
    .querySelectorAll<HTMLElement>(".react-flow__node-instance, .react-flow__node-hall, .react-flow__node-frame")
    .forEach((node) => {
      const box = node.getBoundingClientRect();
      const cut =
        box.left < canvas.left - 0.5 ||
        box.right > canvas.right + 0.5 ||
        box.top < canvas.top - 0.5 ||
        box.bottom > canvas.bottom + 0.5;
      if (cut) outside.push(node.dataset.id ?? "node");
    });
  return outside;
}
