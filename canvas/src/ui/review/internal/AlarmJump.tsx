import { Bell } from "lucide-react";
import type { EventMark } from "../../../model/review";
import { StepButton } from "./PlayControls";

interface AlarmJumpProps {
  readonly marks: readonly EventMark[];
  readonly positionNs: number;
  readonly onEvent: (mark: EventMark) => void;
}

export function AlarmJump({ marks, positionNs, onEvent }: AlarmJumpProps) {
  const alarms = marks.filter((mark) => mark.tone === "alarm");
  const alarm = alarms.find((mark) => mark.atNs > positionNs) ?? alarms[0];
  const label = alarm === undefined ? "Jump to alarm: none" : `Jump to alarm: ${alarm.label}`;
  return (
    <StepButton
      label={label}
      disabled={alarm === undefined}
      onActivate={() => {
        if (alarm !== undefined) onEvent(alarm);
      }}
    >
      <Bell size={14} aria-hidden="true" />
    </StepButton>
  );
}
