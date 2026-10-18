import type { ChangeEvent, SyntheticEvent } from "react";
import { useId, useState } from "react";
import { captureSizeProblem, captureTextProblem } from "../../../model/devices";
import { shared } from "../../primitives";
import styles from "./Capture.module.css";

interface TextProps {
  readonly busy: boolean;
  readonly onRead: (content: string) => void;
}

interface FieldsProps {
  readonly id: string;
  readonly text: string;
  readonly problem: string;
  readonly onType: (next: string) => void;
  readonly onLoad: (event: ChangeEvent<HTMLInputElement>) => void;
}

const unreadable = "That file could not be read.";

function useAnnouncementText() {
  const [text, setText] = useState("");
  const [fileProblem, setFileProblem] = useState("");
  const load = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    if (file === null) return;
    const tooLarge = captureSizeProblem(file.size);
    setFileProblem(tooLarge);
    if (tooLarge !== "") return;
    file
      .text()
      .then(setText)
      .catch(() => {
        setFileProblem(unreadable);
      });
  };
  const type = (next: string) => {
    setFileProblem("");
    setText(next);
  };
  return { text, type, load, problem: captureTextProblem(text) || fileProblem };
}

function Fields({ id, text, problem, onType, onLoad }: FieldsProps) {
  return (
    <>
      <div className={styles.field}>
        <label htmlFor={`${id}-text`}>Announcement text</label>
        <textarea
          id={`${id}-text`}
          className={styles.textarea}
          value={text}
          spellCheck={false}
          aria-invalid={problem === "" ? undefined : true}
          aria-describedby={problem === "" ? undefined : `${id}-problem`}
          onChange={(event) => {
            onType(event.target.value);
          }}
        />
      </div>
      <div className={styles.field}>
        <label htmlFor={`${id}-file`}>Or load it from a file</label>
        <input id={`${id}-file`} type="file" className={styles.file} onChange={onLoad} />
      </div>
    </>
  );
}

export function CaptureText({ busy, onRead }: TextProps) {
  const { text, type, load, problem } = useAnnouncementText();
  const id = useId();
  const blocked = busy || text.trim() === "" || problem !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!blocked) onRead(text);
  };
  return (
    <form className={styles.form} aria-label="Paste an announcement" noValidate onSubmit={submit}>
      <Fields id={id} text={text} problem={problem} onType={type} onLoad={load} />
      {problem !== "" && (
        <p id={`${id}-problem`} role="alert" className={styles.problem}>
          {problem}
        </p>
      )}
      <span className={styles.actions}>
        <button type="submit" className={shared.primary} aria-disabled={blocked}>
          Read the announcement
        </button>
      </span>
    </form>
  );
}
