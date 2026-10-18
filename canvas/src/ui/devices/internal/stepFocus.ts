import { createContext, useContext, useEffect, useRef } from "react";

interface StepFocus {
  readonly pending: boolean;
  readonly done: () => void;
}

export const StepFocusContext = createContext<StepFocus>({ pending: false, done: () => undefined });

export function useStepHeading() {
  const heading = useRef<HTMLHeadingElement>(null);
  const { pending, done } = useContext(StepFocusContext);
  useEffect(() => {
    if (!pending) return;
    heading.current?.focus();
    done();
  }, [pending, done]);
  return heading;
}

export function useFieldFocus(request: { readonly field: string; readonly at: number } | null) {
  const host = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (request === null) return;
    const fields = host.current?.querySelectorAll<HTMLElement>("[data-field]") ?? [];
    for (const element of fields) {
      if (element.dataset.field === request.field) {
        element.focus();
        return;
      }
    }
  }, [request]);
  return host;
}
