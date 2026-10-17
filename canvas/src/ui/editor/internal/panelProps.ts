import { useEffect, useRef } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import type { Plant } from "../../../model/plant";

type FocusTarget = "build" | "instance" | "name";

export class FocusRequest {
  private wanted: FocusTarget | null = null;

  ask(target: FocusTarget): void {
    this.wanted = target;
  }

  take(target: FocusTarget): boolean {
    if (this.wanted !== target) return false;
    this.wanted = null;
    return true;
  }
}

export interface PanelProps {
  readonly draft: Draft;
  readonly plant: Plant;
  readonly editor: EditorSession;
  readonly request: FocusRequest;
}

export function useRequestedFocus<T extends HTMLElement>(request: FocusRequest, target: FocusTarget) {
  const element = useRef<T>(null);
  useEffect(() => {
    if (request.take(target)) element.current?.focus();
  });
  return element;
}
