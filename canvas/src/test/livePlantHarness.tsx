import { act, render } from "@testing-library/react";
import { Profiler } from "react";
import { createLaunchStore } from "../app/launch";
import { createPlantMaker, createPlantSession } from "../app/plant";
import type { PlantMaker } from "../app/plant";
import type { ReviewSession } from "../app/review";
import { createReviewSession } from "../app/review";
import { createNavigationStore } from "../app/shell";
import { createProvenanceSession, createWorkspaceStore } from "../app/workspace";
import { PlantScreen } from "../ui/plant";
import { ConnectedPlant } from "../ui/review";
import { useStore } from "../ui/primitives";
import { factoryBridge, factoryJournal, factoryPath } from "./factoryFixture";
import { fakeClock } from "./fakeClock";
import { flushPromises } from "./deferred";
import { memoryLocation } from "./location";
import type { FakeBridge } from "./fakeBridge";
import type { PlantSession } from "../app/plant";
import type { ProvenanceSession, WorkspaceStore } from "../app/workspace";
import type { NavigationStore } from "../app/shell";
import type { LaunchStore } from "../app/launch";

export const watchedRun = "w-9";

interface Parts {
  readonly plants: PlantSession;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly review: ReviewSession;
  readonly provenance: ProvenanceSession;
  readonly maker: PlantMaker;
}

function Screen({ parts }: { readonly parts: Parts }) {
  const { plants, workspace, navigation, launch, review, provenance, maker } = parts;
  const state = useStore(plants);
  const ready = state.phase === "ready" && state.path === factoryPath ? state : null;
  return (
    <ConnectedPlant review={review} plant={ready} run={watchedRun} connect={launch.connect} now={() => 0}>
      {(session) => (
        <PlantScreen
          plants={plants}
          workspace={workspace}
          navigation={navigation}
          launch={launch}
          provenance={provenance}
          maker={maker}
          domain={factoryPath}
          hall={null}
          startLaunching={false}
          onShowDetails={() => undefined}
          {...(session === null ? {} : { connected: session })}
        />
      )}
    </ConnectedPlant>
  );
}

interface Profile {
  commits: number;
  actualMs: number;
  baseMs: number;
  phases: Record<string, number>;
}

export interface LivePlant {
  readonly profile: Profile;
  readonly bridge: FakeBridge;
  readonly review: ReviewSession;
  readonly parts: Parts;
  readonly timers: ReturnType<typeof fakeClock>;
  feed(fromS: number, toS: number): void;
  send(entries: readonly object[]): void;
}

function renderProfiled(parts: Parts, profile: Profile): void {
  render(
    <Profiler
      id="plant"
      onRender={(_id, phase, actual, base) => {
        profile.phases[phase] = (profile.phases[phase] ?? 0) + 1;
        profile.commits += 1;
        profile.actualMs += actual;
        profile.baseMs += base;
      }}
    >
      <Screen parts={parts} />
    </Profiler>,
  );
}

export async function mountLivePlant(open: (review: ReviewSession, worldId: string) => void): Promise<LivePlant> {
  const bridge = factoryBridge();
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  const workspace = createWorkspaceStore(bridge.workspace);
  const parts: Parts = {
    plants: createPlantSession(bridge.plant),
    workspace,
    provenance: createProvenanceSession(bridge.provenance, workspace),
    maker: createPlantMaker(bridge.plant),
    navigation: createNavigationStore(memoryLocation(`#/plant?domain=${encodeURIComponent(factoryPath)}`)),
    launch: createLaunchStore(bridge.launch, timers, timers),
    review,
  };
  const profile: Profile = { commits: 0, actualMs: 0, baseMs: 0, phases: {} };
  await act(async () => {
    await parts.workspace.open();
  });
  renderProfiled(parts, profile);
  await act(async () => {
    await flushPromises();
  });
  open(review, watchedRun);
  const feed = (fromS: number, toS: number) => {
    act(() => {
      bridge.stream(watchedRun, factoryJournal(fromS, toS), null);
      timers.advance(250);
    });
  };
  const send = (entries: readonly object[]) => {
    act(() => {
      bridge.stream(watchedRun, entries, null);
      timers.advance(250);
    });
  };
  return { bridge, review, parts, timers, feed, send, profile };
}
