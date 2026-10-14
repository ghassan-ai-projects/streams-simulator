const textTypes = new Set(["text", "search", "number", "email", "url", "tel", "password", ""]);

export function inTextField(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  if (target.isContentEditable || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement) {
    return true;
  }
  return target instanceof HTMLInputElement && textTypes.has(target.type);
}

const activators = "button, a[href], summary, [role='button'], [role='checkbox'], [role='radio'], [role='tab'], input";

function scrolls(target: HTMLElement): boolean {
  const overflow = getComputedStyle(target).overflowY;
  return (overflow === "auto" || overflow === "scroll") && target.scrollHeight > target.clientHeight + 1;
}

export function ownsSpace(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  if (target instanceof HTMLInputElement && target.type === "range") return false;
  if (inTextField(target) || target.closest("[role='application']") !== null) return true;
  return target.matches(activators) || scrolls(target);
}

export function plainKey(event: KeyboardEvent | { ctrlKey: boolean; metaKey: boolean; altKey: boolean }): boolean {
  return !event.ctrlKey && !event.metaKey && !event.altKey;
}
