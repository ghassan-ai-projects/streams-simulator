import { useCallback, useMemo, useState } from "react";
import type { DeviceSession, Step } from "../../../app/devices";
import type { Selection } from "../../../model/devices";
import type { Pending } from "./stepFocus";

interface Focus {
  readonly field: string;
  readonly at: number;
}

export function useStepNavigation(devices: DeviceSession, opening: Step | undefined) {
  const [step, setStep] = useState<Step>(opening ?? "interface");
  const [focus, setFocus] = useState<Focus | null>(null);
  const [pending, setPending] = useState<Pending>("fresh");
  const done = useCallback((next: Pending) => {
    setPending(next);
  }, []);
  const pick = (next: Step) => {
    setFocus(null);
    setPending("fresh");
    setStep(next);
  };
  const reveal = (selection: Selection, field: string) => {
    devices.select(selection);
    setFocus((previous) => ({ field, at: (previous?.at ?? 0) + 1 }));
    setPending("none");
    setStep("interface");
  };
  const stepFocus = useMemo(() => ({ pending, done }), [pending, done]);
  return { step, focus, stepFocus, pick, reveal };
}
