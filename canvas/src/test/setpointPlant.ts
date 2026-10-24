import { parsePlant } from "../model/plant";
import { domain } from "./plantFixture";

const effectors = [
  { name: "start", args_schema: { required: ["chiller_id"], properties: { chiller_id: { type: "string" } } } },
  {
    name: "set_setpoint",
    args_schema: {
      required: ["chiller_id", "setpoint_c"],
      properties: { chiller_id: { type: "string" }, setpoint_c: { type: "number", minimum: 4, maximum: 12 } },
    },
  },
];

export const setpointPlant = parsePlant({
  ...domain,
  kinds: domain.kinds.map((kind) =>
    kind.name === "chiller" ? { ...kind, effectors, faults: [{ id: "chiller__fouling" }] } : kind,
  ),
});
