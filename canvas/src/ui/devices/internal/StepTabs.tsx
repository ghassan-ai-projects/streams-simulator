import type { Step, StepStatus } from "../../../app/devices";
import tw from "./Behaviour.module.css";

interface TabsProps {
  readonly step: Step;
  readonly statuses: Readonly<Record<Step, StepStatus>>;
  readonly onPick: (step: Step) => void;
}

const steps: readonly (readonly [Step, string])[] = [
  ["interface", "Interface"],
  ["behaviour", "Behaviour"],
  ["check", "Check"],
];

const chips = { ok: tw.chipOk, warn: tw.chipWarn, busy: tw.chipBusy, "": tw.chip } as const;

export function StepTabs({ step, statuses, onPick }: TabsProps) {
  return (
    <nav className={tw.steps} aria-label="Device steps">
      <ol>
        {steps.map(([key, label], index) => (
          <li key={key}>
            <button
              type="button"
              className={tw.step}
              aria-current={step === key ? "step" : undefined}
              onClick={() => {
                onPick(key);
              }}
            >
              <span className={tw.stepNumber} aria-hidden="true">
                {String(index + 1)}
              </span>
              <span className={tw.stepName}>{label}</span>{" "}
              <span className={chips[statuses[key].tone]}>{statuses[key].text}</span>
            </button>
          </li>
        ))}
      </ol>
    </nav>
  );
}
