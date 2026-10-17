import { compare, register } from "./ratchetShape.ts";

const phase = "P1";
const folder = "src/ui/graph/internal";
const oldStyle = `${folder}/RunScreen.module.css`;
const newStyle = `${folder}/ReviewScreen.module.css`;
const oldElsewhere = "src/ui/run/RunScreen.module.css";

export function runRenameFixtures(): string[] {
  return [...acceptedRename(), ...acceptedMoveAndRename(), ...raisedRename(), ...addedByRename()];
}

function allowance(global: number, important = 0) {
  return { global, important, phase };
}

function acceptedRename(): string[] {
  const base = register({ extracted: true, globalCaps: { [oldStyle]: allowance(2) } });
  const renamed = register({ extracted: true, globalCaps: { [newStyle]: allowance(2) } });
  const shrunk = register({ extracted: true, globalCaps: { [newStyle]: allowance(1) } });
  return [
    ...compare("a stylesheet allowance renamed within its module folder is accepted", base, renamed, []),
    ...compare("a rename within its module folder may shrink the allowance", base, shrunk, []),
  ];
}

function acceptedMoveAndRename(): string[] {
  const base = register({ globalCaps: { [oldElsewhere]: allowance(2) } });
  const renamed = register({ extracted: true, globalCaps: { [newStyle]: allowance(2) } });
  return compare("a move between modules that also renames carries the allowance unchanged", base, renamed, []);
}

function raisedRename(): string[] {
  const base = register({ extracted: true, globalCaps: { [oldStyle]: allowance(2) } });
  const raised = register({ extracted: true, globalCaps: { [newStyle]: allowance(3) } });
  const moveBase = register({ globalCaps: { [oldElsewhere]: allowance(2) } });
  const moveShrunk = register({ extracted: true, globalCaps: { [newStyle]: allowance(1) } });
  const rejected = [`new :global/!important allowance ${newStyle}`];
  return [
    ...compare("a rename that raises the allowance is rejected", base, raised, rejected),
    ...compare("a move and rename that changes the allowance is rejected", moveBase, moveShrunk, rejected),
  ];
}

function addedByRename(): string[] {
  const base = register({ extracted: true, globalCaps: { [oldStyle]: allowance(2) } });
  const other = `${folder}/Other.module.css`;
  const extra = register({
    extracted: true,
    globalCaps: { [newStyle]: allowance(2), [other]: allowance(0, 0) },
  });
  return compare("a rename that also adds a stylesheet to the register is rejected", base, extra, [
    `new :global/!important allowance ${newStyle}`,
    `new :global/!important allowance ${other}`,
  ]);
}
