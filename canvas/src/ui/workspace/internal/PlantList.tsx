import { shared, cx } from "../../primitives";
import { Factory, Pencil, Play, Plus } from "lucide-react";
import type { ReactNode, RefObject } from "react";
import { useRef, useState } from "react";
import type { PlantMaker } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { Lineage, PlantRoute, Workspace } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import type { Route } from "../../../model/shell";
import type { PlantEntry } from "../../../model/workspace";
import { plantFacts, plantOf, plantRoute, plantsOf } from "../../../model/workspace";
import { NewPlantForm } from "./NewPlant";
import { SectionHeading } from "./SectionHeading";
import workspaceStyles from "./WorkspaceScreen.module.css";
import rowStyles from "./ListSection.module.css";
import badgeStyles from "./Badge.module.css";

interface ListProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly examples: readonly ExamplePlant[];
  readonly lineage: Lineage | null;
}

function files(plant: PlantEntry): string {
  return [plant.composition, plant.domain].filter((path) => path !== null).join("\n");
}

function Action({
  label,
  plant,
  run = false,
  children,
  onClick,
}: {
  readonly label: string;
  readonly plant: PlantEntry;
  readonly run?: boolean;
  readonly children: ReactNode;
  readonly onClick: () => void;
}) {
  return (
    <button
      type="button"
      className={run ? cx(workspaceStyles.run, shared.secondary) : shared.secondary}
      aria-label={`${label} ${plant.title}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}

interface RowProps {
  readonly plant: PlantEntry;
  readonly navigation: NavigationStore;
  readonly route: PlantRoute | null;
}

function Actions({ plant, navigation, route }: RowProps) {
  const { domain, composition } = plant;
  const action = route?.action.kind === "behaviour" ? route.action : null;
  const go = (route: Route) => () => {
    navigation.go(route);
  };
  return (
    <span className={rowStyles.actions}>
      {domain !== null && (
        <Action label="Open" plant={plant} onClick={go({ screen: "plant", domain, hall: null })}>
          Open
        </Action>
      )}
      {composition !== null && (
        <Action label="Edit" plant={plant} onClick={go({ screen: "edit", composition })}>
          <Pencil size={13} aria-hidden="true" />
          Edit
        </Action>
      )}
      {action !== null && (
        <Action
          label="Open the Behaviour step of"
          plant={plant}
          onClick={go({ screen: "device", path: action.device, step: "behaviour" })}
        >
          Behaviour
        </Action>
      )}
      {domain !== null && (
        <Action label="Run" plant={plant} run onClick={go({ screen: "plant", domain, hall: null, launch: true })}>
          <Play size={13} aria-hidden="true" />
          Run
        </Action>
      )}
    </span>
  );
}

function PlantRow({ plant, navigation, route }: RowProps) {
  return (
    <li className={rowStyles.item} title={files(plant)}>
      <Factory size={16} aria-hidden="true" className={rowStyles.icon} />
      <span className={rowStyles.lines}>
        <span className={rowStyles.line}>
          <span className={rowStyles.name} title={plant.title}>
            {plant.title}
          </span>
          <span className={rowStyles.id}>{plant.id}</span>
          {plant.composition === null && <span className={badgeStyles.badge}>read-only</span>}
        </span>
        <span className={workspaceStyles.meta}>{plantFacts(plant)}</span>
        {route !== null && (
          <span className={workspaceStyles.meta} title={route.sentence}>
            {route.headline}
          </span>
        )}
      </span>
      <Actions plant={plant} navigation={navigation} route={route} />
    </li>
  );
}

function Heading({
  count,
  creating,
  canCreate,
  trigger,
  onToggle,
}: {
  readonly count: number;
  readonly creating: boolean;
  readonly canCreate: boolean;
  readonly trigger: RefObject<HTMLButtonElement | null>;
  readonly onToggle: () => void;
}) {
  return (
    <SectionHeading id="plants-title" title="Plants" count={count}>
      <button
        type="button"
        ref={trigger}
        className={shared.primary}
        aria-expanded={creating}
        aria-disabled={!canCreate}
        title={canCreate ? undefined : "A new plant starts from an existing one."}
        onClick={onToggle}
      >
        <Plus size={14} aria-hidden="true" />
        New plant
      </button>
    </SectionHeading>
  );
}

function routeOf(plant: PlantEntry, workspace: Workspace, lineage: Lineage | null): PlantRoute | null {
  const devices = workspace.kinds.map((kind) => kind.path);
  return plantRoute(plantOf(lineage, plant.domain ?? ""), devices);
}

export function PlantList({ workspace, navigation, maker, examples, lineage }: ListProps) {
  const plants = plantsOf(workspace);
  const [creating, setCreating] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const bases = workspace.compositions;
  const close = () => {
    setCreating(false);
    trigger.current?.focus();
  };
  return (
    <section className={rowStyles.section} aria-labelledby="plants-title">
      <Heading
        count={plants.length}
        creating={creating}
        canCreate={bases.length + examples.length > 0}
        trigger={trigger}
        onToggle={() => {
          if (bases.length + examples.length > 0) setCreating(!creating);
        }}
      />
      {creating && (
        <NewPlantForm
          bases={bases}
          examples={examples}
          taken={plants.map((plant) => plant.id)}
          maker={maker}
          navigation={navigation}
          onCancel={close}
        />
      )}
      <ul className={rowStyles.rows}>
        {plants.map((plant) => (
          <PlantRow key={plant.id} plant={plant} navigation={navigation} route={routeOf(plant, workspace, lineage)} />
        ))}
      </ul>
    </section>
  );
}
