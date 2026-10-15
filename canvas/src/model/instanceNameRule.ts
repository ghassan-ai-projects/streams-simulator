import composition from "../../../contracts/composition-v0.1.schema.json";

const rule = composition.properties.equipment.items.properties.names.items;
const pattern = new RegExp(rule.pattern);

export function instanceNameRule(): string {
  const text = rule["x-expected"].replace(/[.:_/-]+(?: [.:_/-]+)+/g, (run) => run.replace(/ /g, "\u00a0"));
  return `${text.charAt(0).toUpperCase()}${text.slice(1)}.`;
}

export function instanceNameFits(name: string): boolean {
  return pattern.test(name);
}
