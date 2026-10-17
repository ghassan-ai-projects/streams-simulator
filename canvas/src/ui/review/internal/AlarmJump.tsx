import { Bell } from "lucide-react";
import type { EventMark } from "../../../model/review";
import { nextAlarm } from "../../../model/review";
import { StepButton } from "./PlayControls";

interface AlarmJumpProps {
  readonly marks: readonly EventMark[];
  readonly positionNs: number;
  readonly onEvent: (mark: EventMark) => void;
}

export function AlarmJump({ marks, positionNs, onEvent }: AlarmJumpProps) {
  const alarm = nextAlarm(marks, positionNs);
  const label = alarm === null ? "Jump to alarm: none" : `Jump to alarm: ${alarm.label}`;
  return (
    <StepButton
      label={label}
      disabled={alarm === null}
      onActivate={() => {
        if (alarm !== null) onEvent(alarm);
      }}
    >
      <Bell size={14} aria-hidden="true" />
    </StepButton>
  );
}
