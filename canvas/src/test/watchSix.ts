import type { ReviewSession } from "../app/review";

const channels = [
  { entity: "CT-1", channel: "cooling_tower.basin_temp" },
  { entity: "CT-1", channel: "cooling_tower.fan_power" },
  { entity: "CH-1", channel: "chiller.power" },
  { entity: "CH-1", channel: "chiller.chw_supply_temp" },
  { entity: "P-1", channel: "chw_pump.flow" },
  { entity: "AC-1", channel: "compressor_lead.power" },
] as const;

export function watchSix(review: ReviewSession): void {
  channels.forEach((watched) => {
    review.watchList.toggle("factory-utilities", watched);
  });
  review.watchList.show();
}
