import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { PlantMaker } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import { suggestedId } from "../../../model/workspace";
import { failureLine, shared } from "../../primitives";
import styles from "./Provenance.module.css";

interface CopyProps {
  readonly example: string;
  readonly title: string;
  readonly taken: readonly string[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
  readonly label: string;
}

function useCopy({ example, title, taken, maker, navigation }: CopyProps) {
  const [name, setName] = useState(`${title} copy`);
  const [id, setId] = useState(suggestedId(`${example} copy`, taken));
  const [failure, setFailure] = useState("");
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    void maker.create({ kind: "example", id: example }, id.trim(), name.trim(), "copy").then((made) => {
      if (made.ok) navigation.go({ screen: "edit", composition: made.value.composition });
      else setFailure(failureLine(made.failure.diagnostic));
    });
  };
  return { name, setName, id, setId, failure, submit };
}

export function ExampleCopy(props: CopyProps) {
  const { name, setName, id, setId, failure, submit } = useCopy(props);
  const ready = name.trim() !== "" && id.trim() !== "";
  return (
    <form className={styles.copy} onSubmit={submit} aria-label="Editable copy of this example" noValidate>
      <label>
        Name
        <input
          value={name}
          onChange={(event) => {
            setName(event.target.value);
          }}
        />
      </label>
      <label>
        Id
        <input
          value={id}
          onChange={(event) => {
            setId(event.target.value);
          }}
        />
      </label>
      <button type="submit" className={shared.primary} aria-disabled={!ready}>
        {props.label}
      </button>
      {failure !== "" && (
        <span role="alert" className={styles.refusal}>
          {failure}
        </span>
      )}
    </form>
  );
}
