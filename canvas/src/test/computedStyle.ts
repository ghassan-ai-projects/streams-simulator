const html = "http://www.w3.org/1999/xhtml";

type StyleOf = (element: Element, pseudo?: string | null) => CSSStyleDeclaration;

function shapeOf(element: HTMLElement): string {
  const { tagName, hidden, style } = element;
  return [tagName, element.getAttribute("type"), element.hasAttribute("open"), hidden, style.display].join("|");
}

function inheritedVisibility(element: HTMLElement): string {
  for (let node: HTMLElement | null = element; node !== null; node = node.parentElement) {
    if (node.style.visibility !== "") return node.style.visibility;
  }
  return "visible";
}

function displayOf(real: StyleOf, known: Map<string, string>, element: HTMLElement): string {
  const shape = shapeOf(element);
  let display = known.get(shape);
  if (display === undefined) {
    display = real(element).display;
    known.set(shape, display);
  }
  return display;
}

function quickStyle(real: StyleOf, known: Map<string, string>, element: HTMLElement): CSSStyleDeclaration {
  const quick: Record<string, string> = {
    display: displayOf(real, known, element),
    visibility: inheritedVisibility(element),
  };
  let full: CSSStyleDeclaration | undefined;
  const fullStyle = (): CSSStyleDeclaration => (full ??= real(element));
  return new Proxy({} as CSSStyleDeclaration, {
    get(_target, property) {
      if (typeof property === "string" && property in quick) return quick[property];
      if (property === "getPropertyValue") return (name: string) => quick[name] ?? fullStyle().getPropertyValue(name);
      const value: unknown = Reflect.get(fullStyle(), property);
      return typeof value === "function" ? (value as (...args: unknown[]) => unknown).bind(fullStyle()) : value;
    },
  });
}

export function installQuickComputedStyle(): void {
  const real: StyleOf = window.getComputedStyle.bind(window);
  const known = new Map<string, string>();
  window.getComputedStyle = (element, pseudo) =>
    pseudo || !(element instanceof HTMLElement) || element.namespaceURI !== html
      ? real(element, pseudo)
      : quickStyle(real, known, element);
}
