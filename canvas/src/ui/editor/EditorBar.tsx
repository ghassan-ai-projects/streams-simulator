import { ChevronRight, PenLine, Save, Undo2 } from "lucide-react";
import type { Draft, EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import { domainOf, titleOf } from "../../model/composition";
import { ZoomBar } from "../plant/GraphControls";
import plantStyles from "../plant/PlantScreen.module.css";
import { cx } from "../cx";
import shared from "../shared.module.css";
import { useScreen } from "../useScreen";
import styles from "./Editor.module.css";

interface BarProps {
  readonly draft: Draft;
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
}

interface Status {
  readonly text: string;
  readonly tone: "busy" | "warn" | "ok" | "";
}

function outcomeStatus(draft: Draft): Status {
  if (draft.outcome?.kind === "refused") {
    const count = draft.issues?.listed.length ?? 0;
    return { text: `Saved, not composed: ${String(count)} ${count === 1 ? "issue" : "issues"}`, tone: "warn" };
  }
  if (draft.outcome?.kind === "composed") return { text: `Saved. ${draft.outcome.summary}`, tone: "ok" };
  return { text: "No changes", tone: "" };
}

function statusOf(draft: Draft): Status {
  if (draft.saving) return { text: "Saving and composing…", tone: "busy" };
  if (draft.failure !== null) return { text: `Not saved: ${draft.failure.message}`, tone: "warn" };
  if (draft.doc !== draft.saved) return { text: "Unsaved changes", tone: "busy" };
  return outcomeStatus(draft);
}

function Title({ draft, navigation }: Omit<BarProps, "editor">) {
  const title = titleOf(draft.doc);
  const heading = useScreen(`${title} · editor`);
  return (
    <nav className={cx(plantStyles.crumbs, styles.crumbs)} aria-label="Breadcrumb">
      <button
        type="button"
        className={shared.crumb}
        onClick={() => {
          navigation.go({ screen: "home" });
        }}
      >
        Workspace
      </button>
      <ChevronRight size={14} aria-hidden="true" />
      <h1 aria-current="location" className={plantStyles.here} ref={heading} tabIndex={-1}>
        {title}
      </h1>
    </nav>
  );
}

function OpenPlant({ draft, navigation }: Omit<BarProps, "editor">) {
  if (draft.outcome?.kind !== "composed" || draft.doc !== draft.saved) return null;
  return (
    <button
      type="button"
      className={styles.action}
      onClick={() => {
        navigation.go({ screen: "plant", domain: domainOf(draft.doc), hall: null });
      }}
    >
      Open plant
    </button>
  );
}

export function EditorBar({ draft, editor, navigation }: BarProps) {
  const status = statusOf(draft);
  const clean = draft.doc === draft.saved;
  return (
    <div className={plantStyles.bar}>
      <PenLine size={14} aria-hidden="true" />
      <Title draft={draft} navigation={navigation} />
      <span role="status" className={styles[`status-${status.tone || "idle"}`]} title={status.text}>
        {status.text}
      </span>
      <OpenPlant draft={draft} navigation={navigation} />
      <button
        type="button"
        className={styles.action}
        aria-disabled={clean || draft.saving}
        onClick={() => {
          if (!clean && !draft.saving) editor.revert();
        }}
      >
        <Undo2 size={14} aria-hidden="true" />
        Revert
      </button>
      <button
        type="button"
        className={styles.save}
        aria-disabled={draft.saving}
        onClick={() => {
          void editor.save();
        }}
      >
        <Save size={14} aria-hidden="true" />
        Save
      </button>
      <ZoomBar />
    </div>
  );
}
