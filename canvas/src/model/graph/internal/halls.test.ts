import { describe, expect, it } from "vitest";
import { campus } from "../../../test/plantFixture";
import { parsePlant } from "../../plant";
import { hallMembers, hallOf, hallsOf } from "./halls";
import { plantView } from "./plantView";

describe("halls", () => {
  it("reads a hall from the id's prefix", () => {
    expect(hallOf("h02/M-1")).toBe("h02");
    expect(hallOf("CAL-1")).toBeNull();
    expect(hallsOf(parsePlant(campus(3)))).toEqual(["h01", "h02", "h03"]);
  });

  it("treats a single prefix as an ordinary plant", () => {
    const one = parsePlant({ instances: [{ kind: "k", names: ["h01/a", "h01/b"], area: "A" }] });
    expect(hallsOf(one)).toEqual([]);
    expect(plantView(one, new Map(), { open: null }).nodes).toHaveLength(2);
  });

  it("counts a hall's members by kind", () => {
    const plant = parsePlant(campus(3));
    const hall = hallsOf(plant)[0] ?? "";
    expect(hallMembers(plant, hall).reduce((sum, [, n]) => sum + n, 0)).toBe(
      plant.instances.filter((instance) => hallOf(instance.id) === hall).length,
    );
    expect(hallMembers(plant, "nowhere")).toEqual([]);
  });
});
