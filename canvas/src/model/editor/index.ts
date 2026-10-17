export {
  areaNodesOf,
  componentOf,
  compositionPlant,
  connectionsOf,
  connectionsTouching,
  domainOf,
  equipmentOf,
  linkEnd,
  parseComposition,
  serialise,
  shortKind,
  suggestedName,
  titleOf,
} from "./internal/composition";
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
export { groupIssues, issuesWords, issueTargets, profileWeightOf, removeProfileWeight } from "./internal/issueTargets";
export { addProblem, libraryRows, libraryWords, suggestedWords } from "./internal/addWords";
export { deleteQuestion, linkParts, linkWords, nameCheck } from "./internal/linkWords";
export { instanceNameRule } from "./internal/instanceNameRule";
export type { Composition, Connection, IssueGroup, IssueTargets, LinkEnd, ListedIssue, NewEquipment } from "./contract";
