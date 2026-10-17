import { shared, cx } from "../../primitives";
import { Factory, Pencil, Play, Plus } from "lucide-react";
import type { ReactNode, RefObject } from "react";
import { useRef, useState } from "react";
import type { PlantMaker } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { Workspace } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import type { Route } from "../../../model/shell";
import type { PlantEntry } from "../../../model/workspace";
import { plantFacts, plantsOf } from "../../../model/workspace";
import { NewPlantForm } from "./NewPlant";
import styles from "./HomeScreen.module.css";

interface ListProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly maker: PlantMaker;
  readonly examples: readonly ExamplePlant[];
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
      className={run ? cx(styles.run, shared.button) : cx(styles.open, shared.button)}
      aria-label={`${label} ${plant.title}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}

function Actions({ plant, navigation }: { readonly plant: PlantEntry; readonly navigation: NavigationStore }) {
  const { domain, composition } = plant;
  const go = (route: Route) => () => {
    navigation.go(route);
  };
  return (
    <span className={styles.actions}>
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
      {domain !== null && (
        <Action label="Run" plant={plant} run onClick={go({ screen: "plant", domain, hall: null, launch: true })}>
          <Play size={13} aria-hidden="true" />
          Run
        </Action>
      )}
    </span>
  );
}

function PlantRow({ plant, navigation }: { readonly plant: PlantEntry; readonly navigation: NavigationStore }) {
  return (
    <li className={styles.item} title={files(plant)}>
      <Factory size={16} aria-hidden="true" className={styles.icon} />
      <span className={styles.lines}>
        <span className={styles.line}>
          <span className={styles.name} title={plant.title}>
            {plant.title}
          </span>
          <span className={styles.id}>{plant.id}</span>
          {plant.composition === null && <span className={styles.badge}>read-only</span>}
        </span>
        <span className={styles.meta}>{plantFacts(plant)}</span>
      </span>
      <Actions plant={plant} navigation={navigation} />
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
    <header className={styles.heading}>
      <h2 id="plants-title">Plants</h2>
      <span className={styles.count}>{count}</span>
      <button
        type="button"
        ref={trigger}
        className={styles.create}
        aria-expanded={creating}
        aria-disabled={!canCreate}
        title={canCreate ? undefined : "A new plant starts from an existing one."}
        onClick={onToggle}
      >
        <Plus size={14} aria-hidden="true" />
        New plant
      </button>
    </header>
  );
}

export function PlantList({ workspace, navigation, maker, examples }: ListProps) {
  const plants = plantsOf(workspace);
  const [creating, setCreating] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const bases = workspace.compositions;
  const close = () => {
    setCreating(false);
    trigger.current?.focus();
  };
  return (
    <section className={styles.section} aria-labelledby="plants-title">
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
      <ul className={styles.rows}>
        {plants.map((plant) => (
          <PlantRow key={plant.id} plant={plant} navigation={navigation} />
        ))}
      </ul>
    </section>
  );
}
