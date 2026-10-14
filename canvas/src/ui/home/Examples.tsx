import { FlaskConical, Play } from "lucide-react";
import { useEffect, useState } from "react";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantMaker } from "../../app/plantMaker";
import type { ExamplePlant } from "../../model/library";
import { suggestedId } from "../../model/workspace";
import styles from "./HomeScreen.module.css";

export function useExamples(maker: PlantMaker): readonly ExamplePlant[] {
  const [examples, setExamples] = useState<readonly ExamplePlant[]>([]);
  useEffect(() => {
    let current = true;
    void maker.examples().then((found) => {
      if (current) setExamples(found);
    });
    return () => {
      current = false;
    };
  }, [maker]);
  return examples;
}

interface TryProps {
  readonly examples: readonly ExamplePlant[];
  readonly taken: readonly string[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
}

function useTry({ maker, navigation, taken }: Omit<TryProps, "examples">) {
  const [busy, setBusy] = useState("");
  const [failure, setFailure] = useState("");
  const start = async (example: ExamplePlant) => {
    setBusy(example.id);
    setFailure("");
    const made = await maker.tryExample(example, suggestedId(example.id, taken));
    setBusy("");
    if ("domain" in made) navigation.go({ screen: "plant", domain: made.domain, hall: null, launch: true });
    else setFailure(`${example.title} could not be set up: ${made.message}`);
  };
  return { busy, failure, start };
}

export function ExampleList(props: TryProps) {
  const { busy, failure, start } = useTry(props);
  if (props.examples.length === 0) return null;
  return (
    <section className={styles.examples} aria-labelledby="examples-title">
      <h2 id="examples-title">
        <FlaskConical size={16} aria-hidden="true" />
        Try an example
      </h2>
      <ul>
        {props.examples.map((example) => (
          <li key={example.id}>
            <span className={styles.lines}>
              <span className={styles.name}>{example.title}</span>
              <span className={styles.meta}>{example.description}</span>
            </span>
            <button
              type="button"
              className={styles.create}
              aria-label={`Try ${example.title}`}
              aria-disabled={busy !== ""}
              onClick={() => {
                if (busy === "") void start(example);
              }}
            >
              <Play size={13} aria-hidden="true" />
              {busy === example.id ? "Setting up…" : "Try it"}
            </button>
          </li>
        ))}
      </ul>
      {failure !== "" && (
        <p role="alert" className={styles.failure}>
          {failure}
        </p>
      )}
    </section>
  );
}
