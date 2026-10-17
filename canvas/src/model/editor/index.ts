export {
  areaNodesOf,
  componentOf,
  compositionPlant,
  connectionsOf,
  domainOf,
  equipmentOf,
  linkEnd,
  parseComposition,
  serialise,
  shortKind,
  suggestedName,
  titleOf,
} from "./internal/composition";
export { endpointOf } from "../plant";
export { addArea, areaInUse, areaNameProblem, removeArea, renameArea } from "./internal/areaEdits";
export {
  addEquipment,
  connect,
  disconnect,
  isTemplate,
  moveInstance,
  removeInstance,
  renameInstance,
} from "./internal/compositionEdits";
export { groupIssues, issueTargets, profileWeightOf, removeProfileWeight } from "./internal/issueTargets";
export { instanceNameFits, instanceNameRule } from "./internal/instanceNameRule";
export type { Composition, Connection, IssueGroup, IssueTargets, LinkEnd, ListedIssue, NewEquipment } from "./contract";
