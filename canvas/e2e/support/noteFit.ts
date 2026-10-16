import type { Audit } from "./audit";

export function noteProblems(atLeast = 1): Audit {
  const visible = (element: HTMLElement) =>
    element.closest(".sr-only") === null &&
    element.getClientRects().length > 0 &&
    getComputedStyle(element).visibility !== "hidden";
  const problems: string[] = [];
  const groups = [...document.querySelectorAll<HTMLElement>(".react-flow__node-instance [role=group]")];
  groups.forEach((node) => {
    const own = node.getBoundingClientRect();
    const name = node.getAttribute("aria-label") ?? "";
    node.querySelectorAll<HTMLElement>("[class*='note'], [class*='note'] *").forEach((element) => {
      const cut = element.scrollWidth > element.clientWidth + 1 || element.scrollHeight > element.clientHeight + 1;
      if (visible(element) && cut) problems.push(`${name}: cut ${element.textContent.slice(0, 40)}`);
    });
    node.querySelectorAll<HTMLElement>("*").forEach((element) => {
      const spills = element.getBoundingClientRect().bottom > own.bottom + 0.5;
      if (visible(element) && spills) problems.push(`${name}: spills ${element.textContent.slice(0, 40)}`);
    });
  });
  return { considered: groups.length, problems: problems, met: groups.length >= atLeast };
}
