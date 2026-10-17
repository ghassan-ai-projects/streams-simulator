import { screen, shared } from "../../primitives";
import { PenLine, RefreshCw, Save } from "lucide-react";
import type { Draft, EditorSession } from "../../../app/editor";
import { canSave, checked, isDirty, isSaving, saveStatusOf } from "../../../app/editor";
import type { NavigationStore } from "../../../app/shell";
import { domainOf, equipmentOf, titleOf } from "../../../model/editor";
import { ZoomBar } from "../../graph";
import { Breadcrumb } from "../../primitives";
import { useScreen } from "../../primitives";
import bar from "./EditorBar.module.css";
import { History } from "./EditorHistory";
import { Revert } from "./EditorRevert";
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
