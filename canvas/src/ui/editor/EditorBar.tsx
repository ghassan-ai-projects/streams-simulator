import { PenLine, Redo, RefreshCw, Save, Undo, Undo2 } from "lucide-react";
import { useState } from "react";
import type { Draft, EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import { domainOf, equipmentOf, titleOf } from "../../model/composition";
import { ZoomBar } from "../plant/GraphControls";
import plantStyles from "../plant/PlantScreen.module.css";
import { Breadcrumb } from "../shell/Breadcrumb";
import { useScreen } from "../useScreen";
import styles from "./Editor.module.css";
import type { FocusRequest } from "./panelProps";
import { useEditorKeys } from "./useEditorKeys";

interface BarProps {
  readonly draft: Draft;
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
}

interface Status {
  readonly text: string;
  readonly detail: string | null;
  readonly tone: "busy" | "warn" | "ok" | "";
}

function outcomeStatus(draft: Draft): Status {
  if (draft.outcome?.kind === "refused") {
    const count = draft.issues?.listed.length ?? 0;
    const counted = `${String(count)} ${count === 1 ? "issue" : "issues"}`;
    const why = draft.outcome.message ?? `${counted}, listed in Build`;
    return { text: `Needs fixing: ${why}. Your changes are saved.`, detail: null, tone: "warn" };
  }
  if (draft.outcome?.kind === "composed") {
    return { text: "Ready to run. Saved.", detail: draft.outcome.summary, tone: "ok" };
  }
  return { text: "No changes", detail: null, tone: "" };
}

function statusOf(draft: Draft): Status {
  if (draft.saving) return { text: "Saving and composing…", detail: null, tone: "busy" };
  if (draft.failure !== null) return { text: `Not saved: ${draft.failure.message}`, detail: null, tone: "warn" };
  if (draft.doc !== draft.saved) return { text: "Unsaved changes", detail: null, tone: "busy" };
  return outcomeStatus(draft);
}

function Title({ draft, navigation }: Omit<BarProps, "editor">) {
  const title = titleOf(draft.doc);
  const heading = useScreen(`${title} · editor`);
  const trail = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "home" });
      },
    },
  ];
  return <Breadcrumb trail={trail} here={title} heading={heading} tag="Editing" />;
}

function OpenPlant({ draft, navigation }: Omit<BarProps, "editor">) {
  if (draft.domainEtag === "" && draft.outcome?.kind !== "composed") return null;
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

function ShowIssues({
  draft,
  editor,
  request,
}: {
  readonly draft: Draft;
  readonly editor: EditorStore;
  readonly request: FocusRequest;
}) {
  if ((draft.issues?.listed.length ?? 0) === 0) return null;
  return (
    <button
      type="button"
      className={styles.action}
      onClick={() => {
        request.ask("build");
        editor.select(null);
      }}
    >
      Show issues
    </button>
  );
}

function Reload({ draft, editor }: Omit<BarProps, "navigation">) {
  if (draft.failure?.code !== "conflict") return null;
  return (
    <button
      type="button"
      className={styles.action}
      onClick={() => {
        void editor.reload();
      }}
    >
      <RefreshCw size={14} aria-hidden="true" />
      Reload from disk
    </button>
  );
}

function useHistoryAnnouncement(draft: Draft): string {
  const counts = { past: draft.past.length, future: draft.future.length };
  const [seen, setSeen] = useState(counts);
  const [said, setSaid] = useState("");
  if (seen.past !== counts.past || seen.future !== counts.future) {
    setSeen(counts);
    const left = `${String(counts.past)} more to undo`;
    if (counts.future === seen.future + 1) setSaid(`Undone. ${left}.`);
    else if (counts.future === seen.future - 1) setSaid(`Redone. ${left}.`);
    else setSaid("");
  }
  return said;
}

function History({ draft, editor }: Omit<BarProps, "navigation">) {
  const mac = navigator.platform.toLowerCase().includes("mac");
  const key = mac ? "⌘" : "Ctrl+";
  const said = useHistoryAnnouncement(draft);
  return (
    <span className={styles.history}>
      <span className="sr-only" aria-live="polite">
        {said}
      </span>
      <button
        type="button"
        className={styles.icon}
        aria-label="Undo"
        aria-keyshortcuts="Control+Z Meta+Z"
        title={`Undo (${key}Z)`}
        aria-disabled={draft.past.length === 0}
        onClick={() => {
          editor.undo();
        }}
      >
        <Undo size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={styles.icon}
        aria-label="Redo"
        aria-keyshortcuts="Control+Shift+Z Meta+Shift+Z"
        title={`Redo (${key}Shift+Z)`}
        aria-disabled={draft.future.length === 0}
        onClick={() => {
          editor.redo();
        }}
      >
        <Redo size={14} aria-hidden="true" />
      </button>
    </span>
  );
}

function Revert({ disabled, editor }: { readonly disabled: boolean; readonly editor: EditorStore }) {
  const [confirming, setConfirming] = useState(false);
  return (
    <>
      <span className="sr-only" aria-live="polite">
        {confirming ? "Press Discard changes again to drop every unsaved change" : ""}
      </span>
      <RevertButton disabled={disabled} editor={editor} confirming={confirming} setConfirming={setConfirming} />
    </>
  );
}

function RevertButton({
  disabled,
  editor,
  confirming,
  setConfirming,
}: {
  readonly disabled: boolean;
  readonly editor: EditorStore;
  readonly confirming: boolean;
  readonly setConfirming: (confirming: boolean) => void;
}) {
  return (
    <button
      type="button"
      className={confirming ? styles.danger : styles.action}
      aria-disabled={disabled}
      onBlur={() => {
        setConfirming(false);
      }}
      onClick={() => {
        if (disabled) return;
        if (confirming) editor.revert();
        setConfirming(!confirming);
      }}
    >
      <Undo2 size={14} aria-hidden="true" />
      {confirming ? "Discard changes" : "Revert"}
    </button>
  );
}

export function EditorBar({ draft, editor, navigation, request }: BarProps & { readonly request: FocusRequest }) {
  useEditorKeys(editor);
  const status = statusOf(draft);
  const clean = draft.doc === draft.saved;
  return (
    <div className={plantStyles.bar}>
      <PenLine size={14} aria-hidden="true" />
      <Title draft={draft} navigation={navigation} />
      <OpenPlant draft={draft} navigation={navigation} />
      <ShowIssues draft={draft} editor={editor} request={request} />
      <Reload draft={draft} editor={editor} />
      {equipmentOf(draft.doc).length > 0 && <ZoomBar />}
      <span role="status" className={styles[`status-${status.tone || "idle"}`]} title={status.detail ?? status.text}>
        {status.text}
      </span>
      <History draft={draft} editor={editor} />
      <Revert disabled={clean || draft.saving} editor={editor} />
      <button
        type="button"
        className={styles.save}
        aria-disabled={clean || draft.saving}
        title={clean ? "Nothing to save" : "Save and compose (Ctrl+S)"}
        onClick={() => {
          if (!clean) void editor.save();
        }}
      >
        <Save size={14} aria-hidden="true" />
        Save
      </button>
    </div>
  );
}
