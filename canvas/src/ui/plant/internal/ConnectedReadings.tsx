import { OriginMark } from "../../graph";
import type { Connected } from "./connected";

export function ConnectedReadings({ connected, entity }: { readonly connected: Connected; readonly entity: string }) {
  const live = connected.live(entity);
  return (
    <section aria-label="Live instance">
      <OriginMark origin={connected.origins.get(entity) ?? null} />
      {live.shown.length === 0 ? (
        <p>Waiting for readings…</p>
      ) : (
        <dl>
          {live.shown.map((reading) => (
            <div key={reading.channel}>
              <dt>{reading.channel}</dt>
              <dd>
                {reading.value} {reading.unit}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </section>
  );
}
