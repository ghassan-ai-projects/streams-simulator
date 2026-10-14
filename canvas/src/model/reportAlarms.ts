export interface AlarmActivation {
  readonly atNs: number;
  readonly entity: string;
  readonly state: string;
  readonly priority: string;
}

type Json = Record<string, unknown>;
const object = (value: unknown): Json =>
  typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
const list = (value: unknown): readonly unknown[] => (Array.isArray(value) ? value : []);
const text = (value: unknown): string => (typeof value === "string" ? value : "");

export function parseAlarmActivations(entities: unknown): readonly AlarmActivation[] {
  return list(entities).flatMap((raw) => {
    const entity = object(raw);
    return list(entity.alarms).flatMap((rawAlarm) => {
      const alarm = object(rawAlarm);
      return list(alarm.activation_times_ns)
        .filter((at): at is number => typeof at === "number" && Number.isFinite(at))
        .map((atNs) => ({
          atNs,
          entity: text(entity.entity_id),
          state: text(alarm.state),
          priority: text(alarm.priority),
        }));
    });
  });
}
