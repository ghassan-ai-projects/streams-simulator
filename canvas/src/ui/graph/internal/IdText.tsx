import { fitId, idCap, idRoom, tierOf } from "../../../model/graph";

export function IdText({
  label,
  zoom,
  words,
}: {
  readonly label: string;
  readonly zoom: number;
  readonly words: number | null;
}) {
  const tier = tierOf(zoom);
  if (tier !== "compact" && tier !== "overview") return label;
  const fitted = fitId(label, idRoom(zoom, words, tier), zoom, idCap[tier]);
  if (fitted.text === label) return <span style={{ fontSize: `${String(fitted.size)}px` }}>{label}</span>;
  return (
    <span style={{ fontSize: `${String(fitted.size)}px` }}>
      <span aria-hidden="true">{fitted.text}</span>
      <span className="sr-only">{label}</span>
    </span>
  );
}
