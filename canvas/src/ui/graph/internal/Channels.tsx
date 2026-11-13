import { channelRows, channelWords } from "../../../model/graph";
import type { ViewNode } from "../../../model/graph";
import styles from "./readout.module.css";

export function Channels({ view }: { readonly view: ViewNode }) {
  const { listed, more } = channelRows(view.channels.length);
  return (
    <dl className={styles.readout} data-part="channels">
      {view.channels.slice(0, listed).map((channel, index) => (
        <div key={channel}>
          <dt title={channelWords(channel)}>{channelWords(channel)}</dt>
          <dd>
            <span className={styles.unit}>{view.units[index] ?? ""}</span>
          </dd>
        </div>
      ))}
      {view.channels.length === 0 && <span className={styles.none}>no measured channels</span>}
      {more > 0 && (
        <div className={styles.more}>
          <dt>{`+${String(more)} more`}</dt>
          <dd>in the inspector</dd>
        </div>
      )}
    </dl>
  );
}
