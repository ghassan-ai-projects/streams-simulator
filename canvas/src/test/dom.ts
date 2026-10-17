import type { Mock } from "vitest";
import { onTestFinished, vi } from "vitest";

export function stubScrollMetrics(node: HTMLElement, client: number, total: number, top: number): void {
  Object.defineProperty(node, "clientHeight", { value: client, configurable: true });
  Object.defineProperty(node, "scrollHeight", { value: total, configurable: true });
  Object.defineProperty(node, "scrollTop", { value: top, configurable: true });
}

export function stubScrollIntoView(): Mock {
  const stub = vi.fn();
  definePrototype(Element.prototype, "scrollIntoView", stub);
  return stub;
}

export function stubCheckVisibility(visible: (node: HTMLElement) => boolean): void {
  definePrototype(HTMLElement.prototype, "checkVisibility", function (this: HTMLElement) {
    return visible(this);
  });
}

export interface ResizeWatcher {
  readonly observed: Element[];
  disconnected: boolean;
  trigger(): void;
}

export function stubResizeObserver(): ResizeWatcher[] {
  const watchers: ResizeWatcher[] = [];
  class StubObserver implements ResizeWatcher {
    readonly observed: Element[] = [];
    disconnected = false;
    constructor(private readonly callback: () => void) {
      watchers.push(this);
    }
    observe(node: Element) {
      this.observed.push(node);
    }
    disconnect() {
      this.disconnected = true;
    }
    unobserve() {
      return undefined;
    }
    trigger() {
      this.callback();
    }
  }
  vi.stubGlobal("ResizeObserver", StubObserver);
  return watchers;
}

function definePrototype(target: object, name: string, value: unknown): void {
  const original = Object.getOwnPropertyDescriptor(target, name);
  Object.defineProperty(target, name, { configurable: true, writable: true, value });
  onTestFinished(() => {
    if (original === undefined) Reflect.deleteProperty(target, name);
    else Object.defineProperty(target, name, original);
  });
}

export function stubFlowDom(): ResizeWatcher[] {
  class StubMatrix {
    readonly m22 = 1;
  }
  vi.stubGlobal("DOMMatrixReadOnly", StubMatrix);
  vi.stubGlobal("matchMedia", (query: string) => ({ matches: query.includes("reduce"), media: query }));
  definePrototype(Element.prototype, "scrollIntoView", vi.fn());
  definePrototype(HTMLElement.prototype, "offsetWidth", 1200);
  definePrototype(HTMLElement.prototype, "offsetHeight", 800);
  return stubResizeObserver();
}
