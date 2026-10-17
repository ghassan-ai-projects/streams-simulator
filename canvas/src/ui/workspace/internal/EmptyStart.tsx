import { Plus } from "lucide-react";
import { useRef, useState } from "react";
import type { NavigationStore } from "../../../app/shell";
import type { PlantMaker } from "../../../app/plant";
import type { ExamplePlant } from "../../../model/plant";
import { shared } from "../../primitives";
import { NewPlantForm } from "./NewPlant";
import styles from "./EmptyStart.module.css";

interface BlankStartProps {
  readonly examples: readonly ExamplePlant[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
}

export function BlankPlantStart({ examples, maker, navigation }: BlankStartProps) {
  const [creating, setCreating] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const available = examples.length > 0;
  const close = () => {
    setCreating(false);
    trigger.current?.focus();
  };
  return (
    <div className={styles.blankStart}>
      <button
        type="button"
        ref={trigger}
        className={shared.primary}
        aria-expanded={creating}
        aria-disabled={!available}
        title={available ? undefined : "The plant frame is still loading."}
        onClick={() => {
          if (available) setCreating(!creating);
        }}
      >
        <Plus size={14} aria-hidden="true" />
        New plant
      </button>
      {creating && (
        <NewPlantForm
          blank
          bases={[]}
          examples={examples}
          taken={[]}
          maker={maker}
          navigation={navigation}
          onCancel={close}
        />
      )}
    </div>
  );
}
