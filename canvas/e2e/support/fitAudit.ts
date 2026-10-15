import type { Locator } from "@playwright/test";

export async function headingInside(frame: Locator, name: string): Promise<boolean> {
  const title = await frame.getByText(name).boundingBox();
  const outer = await frame.boundingBox();
  return title !== null && outer !== null && title.x + title.width <= outer.x + outer.width - 8;
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
