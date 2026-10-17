import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { BaseChoice, PlantMaker } from "../../../app/plant";
import { baseChoices, newPlantReadiness, startingMode } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { Definition } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import { Field } from "./NewPlantField";
import { Actions } from "./NewPlantActions";
import { Choices } from "./NewPlantChoices";
import { useCreate, useIdProblem, useNaming } from "./useNewPlant";
import styles from "./NewPlant.module.css";

export interface NewPlantProps {
  readonly bases: readonly Definition[];
  readonly examples: readonly ExamplePlant[];
  readonly taken: readonly string[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
  readonly blank?: boolean;
}

export function NewPlantForm(props: NewPlantProps & { readonly onCancel: () => void }) {
  const { bases, examples, taken, blank = false, onCancel } = props;
  const { name, setName, id, setId, followAgain } = useNaming(taken);
  const choices = baseChoices(bases, examples);
  const [base, setBase] = useState<BaseChoice | null>(choices[0] ?? null);
  const [mode, setMode] = useState(startingMode({ blank }));
  const { problem, leave, reportId } = useIdProblem(props.maker, base, name, id, followAgain);
  const { failure, create } = useCreate(props, reportId);
  const reason = newPlantReadiness(name, id, base);
  const ready = reason === "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready && base !== null) void create(base, id.trim(), name.trim(), mode);
  };
  return (
    <form
      className={styles.newPlant}
      onSubmit={submit}
      onKeyDown={(event) => {
        if (event.key === "Escape") onCancel();
      }}
      aria-label="New plant"
      noValidate
    >
      <Field first label="Name" placeholder="e.g. Cooling hall" value={name} onChange={setName} />
      <Field
        label="Id"
        hint="Lower case letters, digits and hyphens, starting with a letter."
        value={id}
        onChange={setId}
        onLeave={leave}
        selectOnFocus
        problem={problem}
      />
      <Choices choices={choices} base={base} setBase={setBase} blank={blank} mode={mode} setMode={setMode} />
      <Actions failure={failure} reason={reason} onCancel={onCancel} />
    </form>
  );
}
