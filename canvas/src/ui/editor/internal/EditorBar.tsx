import { screen, shared } from "../../primitives";
import { PenLine, Redo, RefreshCw, Save, Undo, Undo2 } from "lucide-react";
import { useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import { canSave, checked, isDirty, isSaving, saveStatusOf } from "../../../app/editor";
import type { NavigationStore } from "../../../app/shell";
import { domainOf, equipmentOf, titleOf } from "../../../model/editor";
import { ZoomBar } from "../../graph";
import { Breadcrumb } from "../../primitives";
import { useScreen } from "../../primitives";
import styles from "./Editor.module.css";
import bar from "./EditorBar.module.css";
import type { FocusRequest } from "./panelProps";
import { useEditorKeys } from "./useEditorKeys";

interface BarProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
  readonly navigation: NavigationStore;
}

function Title({ draft, navigation }: Omit<BarProps, "editor">) {
  const title = titleOf(draft.doc);
  const heading = useScreen(`${title} · editor`);
  const trail = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "workspace" });
      },
    },
  ];
  return <Breadcrumb trail={trail} here={title} heading={heading} tag="Editing" />;
}

function OpenPlant({ draft, navigation }: Omit<BarProps, "editor">) {
  if (!checked(draft)) return null;
  return (
    <button
      type="button"
      className={shared.secondary}
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
  readonly editor: EditorSession;
  readonly request: FocusRequest;
}) {
  if ((draft.issues?.listed.length ?? 0) === 0) return null;
  return (
    <button
      type="button"
      className={shared.secondary}
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
  if (!(draft.save.kind === "write-failed" && draft.save.failure.kind === "conflict")) return null;
  return (
    <button
      type="button"
      className={shared.secondary}
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
    <span className={bar.history}>
      <span className="sr-only" aria-live="polite">
        {said}
      </span>
      <button
        type="button"
        className={bar.icon}
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
        className={bar.icon}
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

function Revert({ disabled, editor }: { readonly disabled: boolean; readonly editor: EditorSession }) {
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
  readonly editor: EditorSession;
  readonly confirming: boolean;
  readonly setConfirming: (confirming: boolean) => void;
}) {
  return (
    <button
      type="button"
      className={confirming ? styles.danger : shared.secondary}
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
  const status = saveStatusOf(draft);
  const saving = isSaving(draft);
  const clean = !isDirty(draft);
  const savable = canSave(draft);
  return (
    <div className={screen.bar}>
      <PenLine size={14} aria-hidden="true" />
      <Title draft={draft} navigation={navigation} />
      <OpenPlant draft={draft} navigation={navigation} />
      <ShowIssues draft={draft} editor={editor} request={request} />
      <Reload draft={draft} editor={editor} />
      {equipmentOf(draft.doc).length > 0 && <ZoomBar />}
      <span role="status" className={bar[`status-${status.tone || "idle"}`]} title={status.detail ?? status.text}>
        {status.text}
      </span>
      <History draft={draft} editor={editor} />
      <Revert disabled={clean || saving} editor={editor} />
      <button
        type="button"
        className={shared.primary}
        aria-disabled={!savable || saving}
        title={savable ? "Save and compose (Ctrl+S)" : "Nothing to save"}
        onClick={() => {
          if (savable) void editor.save();
        }}
      >
        <Save size={14} aria-hidden="true" />
        Save
      </button>
    </div>
  );
}
