import { createContext, useContext, useEffect, useRef } from "react";

export type Pending = "none" | "fresh" | "interim";

interface StepFocus {
  readonly pending: Pending;
  readonly done: (next: Pending) => void;
}

export const StepFocusContext = createContext<StepFocus>({ pending: "none", done: () => undefined });

export type StepKind = "final" | "interim";

export function useStepHeading(kind: StepKind = "final") {
  const heading = useRef<HTMLHeadingElement>(null);
  const { pending, done } = useContext(StepFocusContext);
  useEffect(() => {
    if (pending === "none" || (kind === "interim" && pending === "interim")) return;
    heading.current?.focus();
    done(kind === "interim" ? "interim" : "none");
  }, [pending, done, kind]);
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
