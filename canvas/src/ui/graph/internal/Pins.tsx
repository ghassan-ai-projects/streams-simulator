import { Handle, Position } from "@xyflow/react";
import { useRef } from "react";
import { pinWords } from "../../../model/graph";
import type { Pin } from "../../../model/plant";
import { familyNames, shapeOf } from "../../../model/plant";
import { cx } from "../../primitives";
import { useLabelFit } from "./labelFit";
import pins from "./pins.module.css";
import vendor from "./vendor.module.css";

function pinLook(pin: Pin): { readonly family: string; readonly shape: string; readonly typeName: string } {
  if (pin.family === null) {
    return { family: "none", shape: "circle", typeName: "untyped" };
  }
  return { family: pin.family, shape: shapeOf(pin.family), typeName: familyNames[pin.family] };
}

const connectionWords = {
  connected: "connected",
  open: "not connected",
  unknown: "link without a recorded port",
} as const;

function PinRow({
  pin,
  side,
  editable,
  zoom,
}: {
  readonly pin: Pin;
  readonly side: "in" | "out";
  readonly editable: boolean;
  readonly zoom: number;
}) {
  const look = pinLook(pin);
  const words = pinWords(pin.name, side);
  const label = useRef<HTMLSpanElement>(null);
  useLabelFit(label, zoom);
  return (
    <div className={cx(pins.pin, pins[side])} title={`${pin.name} · ${pin.type || look.typeName} · ${look.typeName}`}>
      <Handle
        type={side === "in" ? "target" : "source"}
        position={side === "in" ? Position.Left : Position.Right}
        id={pin.name}
        className={cx(pins.dot, vendor.dot, pins[look.shape], pins[pin.connection], `family-${look.family}`)}
        isConnectable={editable}
      />
      <span ref={label} className={pins.pinName} data-part="pin-name">
        <span className={pins.pinFull}>{words.full}</span>
        <span className={pins.pinShort}>{words.short}</span>
      </span>
      <span className="sr-only">{`, ${side === "in" ? "input" : "output"}, ${look.typeName}, ${connectionWords[pin.connection]}`}</span>
    </div>
  );
}

export function Pins({
  ins,
  outs,
  editable,
  zoom,
}: {
  readonly ins: readonly Pin[];
  readonly outs: readonly Pin[];
  readonly editable: boolean;
  readonly zoom: number;
}) {
  if (ins.length + outs.length === 0) {
    return (
      <div className={pins.none} title="The component file lists no ports, so none can be drawn">
        no port metadata
      </div>
    );
  }
  return (
    <div className={pins.pins}>
      <div>
        {ins.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="in" editable={editable} zoom={zoom} />
        ))}
      </div>
      <div>
        {outs.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="out" editable={editable} zoom={zoom} />
        ))}
      </div>
    </div>
  );
}
