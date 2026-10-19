import { OriginMark } from "../../graph";
import type { Connected, ConnectedDevice } from "./connected";

function DeviceReadings({ device }: { readonly device: ConnectedDevice }) {
  return (
    <>
      <p>{device.line}</p>
      {device.readings.length === 0 ? (
        <p>No reading at this time.</p>
      ) : (
        <dl>
          {device.readings.map((reading) => (
            <div key={reading.name}>
              <dt>{reading.name}</dt>
              <dd>
                {reading.valueText} · {reading.stateWords}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </>
  );
}

function ShownReadings({ connected, entity }: { readonly connected: Connected; readonly entity: string }) {
  const live = connected.live(entity);
  if (live.shown.length === 0) return <p>Waiting for readings…</p>;
  return (
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
  );
}

export function ConnectedReadings({ connected, entity }: { readonly connected: Connected; readonly entity: string }) {
  const device = connected.device(entity);
  const interrupt = connected.live(entity).interrupt;
  return (
    <section aria-label="Live instance">
      <OriginMark origin={connected.origins.get(entity) ?? null} />
      {interrupt !== null && <p data-interrupt={interrupt.latched ? "latched" : interrupt.mode}>{interrupt.words}</p>}
      {device === null ? <ShownReadings connected={connected} entity={entity} /> : <DeviceReadings device={device} />}
    </section>
  );
}
