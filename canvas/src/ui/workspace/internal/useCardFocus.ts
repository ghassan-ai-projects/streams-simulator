import type { RefObject } from "react";
import { useEffect, useState } from "react";

type Pending = { readonly after: number } | { readonly at: number } | null;

export interface CardFocus {
  readonly afterSkip: (index: number) => void;
  readonly afterUndo: (index: number) => void;
}

function targetIn(root: HTMLElement | null, pending: NonNullable<Pending>): HTMLElement | null {
  const actions = Array.from(root?.querySelectorAll<HTMLElement>("[data-actions-of]") ?? []);
  const index = (element: HTMLElement) => Number(element.dataset.actionsOf);
  const found = actions.find((element) =>
    "after" in pending ? index(element) > pending.after : index(element) === pending.at,
  );
  return found?.querySelector<HTMLElement>("button") ?? null;
}

export function useCardFocus(list: RefObject<HTMLElement | null>, heading: RefObject<HTMLElement | null>): CardFocus {
  const [pending, setPending] = useState<Pending>(null);
  useEffect(() => {
    if (pending === null) return;
    (targetIn(list.current, pending) ?? heading.current)?.focus();
  }, [pending, list, heading]);
  return {
    afterSkip: (index) => {
      setPending({ after: index });
    },
    afterUndo: (index) => {
      setPending({ at: index });
    },
  };
}
