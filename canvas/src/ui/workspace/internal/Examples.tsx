import { FlaskConical, Play } from "lucide-react";
import { useEffect, useState } from "react";
import type { NavigationStore } from "../../../app/shell";
import type { PlantMaker } from "../../../app/plant";
import type { ExamplePlant } from "../../../model/plant";
import { suggestedId } from "../../../model/workspace";
import { cx } from "../../primitives";
import home from "./HomeScreen.module.css";
import rowStyles from "./ListSection.module.css";
import styles from "./Examples.module.css";

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
    else setFailure(`${example.title} could not be set up: ${made.diagnostic.message}`);
  };
  return { busy, failure, start };
}

function TryButton({
  example,
  busy,
  onTry,
}: {
  readonly example: ExamplePlant;
  readonly busy: string;
  readonly onTry: (example: ExamplePlant) => Promise<void>;
}) {
  return (
    <button
      type="button"
      className={home.create}
      aria-disabled={busy !== ""}
      onClick={() => {
        if (busy === "") void onTry(example);
      }}
    >
      <Play size={13} aria-hidden="true" />
      {busy === example.id ? "Setting up…" : "Try it"}
      <span className="sr-only">{`: ${example.title}`}</span>
    </button>
  );
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
            <span className={rowStyles.lines}>
              <span className={rowStyles.name}>{example.title}</span>
              <span className={cx(home.meta, styles.description)}>{example.description}</span>
            </span>
            <TryButton example={example} busy={busy} onTry={start} />
          </li>
        ))}
      </ul>
      <p role="status" className="sr-only">
        {busy === "" ? "" : `Setting up ${props.examples.find((example) => example.id === busy)?.title ?? busy}…`}
      </p>
      {failure !== "" && (
        <p role="alert" className={home.failure}>
          {failure}
        </p>
      )}
    </section>
  );
}
