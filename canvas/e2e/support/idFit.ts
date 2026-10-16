import type { Audit } from "./audit";

export function idProblems(atLeast = 1): Audit {
  const graph = document.querySelector("[aria-roledescription='plant graph']");
  const zoom = Number(graph === null ? 1 : getComputedStyle(graph).getPropertyValue("--zoom"));
  const lineTops = (element: HTMLElement): Set<number> => {
    const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
    const tops = new Set<number>();
    while (walker.nextNode()) {
      const owner = walker.currentNode as Text;
      if (owner.parentElement?.closest(".sr-only") !== null) continue;
      const range = document.createRange();
      range.selectNodeContents(owner);
      for (const box of range.getClientRects()) if (box.width > 0) tops.add(Math.round(box.top));
    }
    return tops;
  };
  const problemsOf = (element: HTMLElement): string[] => {
    const name = element.getAttribute("title") ?? "";
    const inner = element.firstElementChild ?? element;
    const size = parseFloat(getComputedStyle(inner).fontSize) * zoom;
    const found: string[] = [];
    if (lineTops(element).size > 1) found.push(`${name}: broken over several lines`);
    if (size < 7.9) found.push(`${name}: ${size.toFixed(1)} px`);
    if (element.scrollWidth > element.clientWidth + 1) found.push(`${name}: overflows`);
    return found;
  };
  const ids = [...document.querySelectorAll<HTMLElement>(".react-flow__node-instance [class*='id_']")];
  const problems = ids.filter((element) => element.getClientRects().length > 0).flatMap(problemsOf);
  return { considered: ids.length, problems, met: ids.length >= atLeast };
}
