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
              <dd title={reading.unitTitle === "" ? undefined : reading.unitTitle}>
                {reading.valueText} · {reading.stateWords}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </>
  );
}

export function ConnectedReadings({ connected, entity }: { readonly connected: Connected; readonly entity: string }) {
  const device = connected.device(entity);
  const interrupt = connected.live(entity).interrupt;
  return (
    <section aria-label="Live instance">
      <OriginMark origin={connected.origins.get(entity) ?? null} />
      {interrupt !== null && <p data-interrupt={interrupt.latched ? "latched" : interrupt.mode}>{interrupt.words}</p>}
      {device === null ? connected.observed(entity) : <DeviceReadings device={device} />}
    </section>
  );
}
