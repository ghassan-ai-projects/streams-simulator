import type { Audit } from "./audit";

export function textClashes(atLeast = 1): Audit {
  interface Box {
    readonly text: string;
    readonly owner: Text;
    readonly box: DOMRect;
  }
  const boxesOf = (node: HTMLElement): Box[] => {
    const walker = document.createTreeWalker(node, NodeFilter.SHOW_TEXT);
    const boxes: Box[] = [];
    while (walker.nextNode()) {
      const owner = walker.currentNode as Text;
      const text = owner.textContent.trim();
      const parent = owner.parentElement;
      if (text === "" || parent?.closest(".sr-only") !== null) continue;
      if (getComputedStyle(parent).visibility === "hidden") continue;
      const range = document.createRange();
      range.selectNodeContents(owner);
      for (const box of range.getClientRects()) {
        if (box.width > 0 && box.height > 0) boxes.push({ text, owner, box });
      }
    }
    return boxes;
  };
  const meet = (a: DOMRect, b: DOMRect) =>
    Math.min(a.right, b.right) - Math.max(a.left, b.left) > 1 &&
    Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top) > 1;
  const nodes = [...document.querySelectorAll<HTMLElement>(".react-flow__node-instance")];
  const examined = nodes.reduce((sum, node) => sum + boxesOf(node).length, 0);
  const problems = nodes.flatMap((node) => {
    const boxes = boxesOf(node);
    return boxes.flatMap((a, index) =>
      boxes
        .slice(index + 1)
        .filter((b) => a.owner !== b.owner && meet(a.box, b.box))
        .map((b) => `${node.dataset.id ?? ""}: ${a.text} over ${b.text}`),
    );
  });
  return { considered: examined, problems, met: examined >= atLeast };
}
