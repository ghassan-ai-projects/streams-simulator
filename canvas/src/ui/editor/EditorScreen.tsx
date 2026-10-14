import { ReactFlowProvider } from "@xyflow/react";
import { useEffect, useMemo, useState } from "react";
import type { Draft, EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import { compositionPlant } from "../../model/composition";
import { connect } from "../../model/compositionEdits";
import { plantView } from "../../model/plantView";
import { Navigator } from "../plant/GraphControls";
import { Outline } from "../plant/Outline";
import { PlantGraph } from "../plant/PlantGraph";
import plantStyles from "../plant/PlantScreen.module.css";
import { useStore } from "../useStore";
import { EditorBar } from "./EditorBar";
import styles from "./Editor.module.css";
import { EditorPanel } from "./EditorPanel";
import { FocusRequest } from "./panelProps";

interface EditorScreenProps {
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
  readonly composition: string;
}

function usePlant(draft: Draft) {
  const titles = useMemo(() => new Map(draft.library.map((entry) => [entry.component, entry.title])), [draft.library]);
  const plant = useMemo(
    () => compositionPlant(draft.doc, (component) => titles.get(component) ?? component),
    [draft.doc, titles],
  );
  const view = useMemo(() => plantView(plant, draft.specs, null), [plant, draft.specs]);
  const flagged = useMemo(() => new Set(draft.issues?.instances.keys() ?? []), [draft.issues]);
  return { plant, view, flagged };
}

function useLeaveGuard(dirty: boolean) {
  useEffect(() => {
    if (!dirty) return undefined;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => {
      window.removeEventListener("beforeunload", warn);
    };
  }, [dirty]);
}

function Editing({ draft, editor, navigation }: { readonly draft: Draft } & Omit<EditorScreenProps, "composition">) {
  const { plant, view, flagged } = usePlant(draft);
  const [request] = useState(() => new FocusRequest());
  useLeaveGuard(draft.doc !== draft.saved);
  const select = (id: string | null) => {
    editor.select(id);
  };
  return (
    <ReactFlowProvider>
      <div className={plantStyles.screen}>
        <div className={plantStyles.side}>
          <Outline plant={plant} hall={null} selected={draft.selected} onSelect={select} />
          <Navigator />
        </div>
        <section className={plantStyles.center} aria-label={`Editing ${plant.title || plant.id}`}>
          <EditorBar draft={draft} editor={editor} navigation={navigation} />
          {plant.instances.length === 0 ? (
            <p className={styles.empty}>No equipment yet. Add the first piece from the Build panel.</p>
          ) : (
            <PlantGraph
              view={view}
              viewKey={`edit:${draft.path}`}
              fitKey={`${draft.path}:${String(plant.instances.length)}`}
              selected={draft.selected}
              flagged={flagged}
              minFitZoom={0.5}
              focusZoom={1}
              onSelect={select}
              onConnect={(from, to) => {
                editor.edit((doc) => connect(doc, from, to));
              }}
            />
          )}
        </section>
        <EditorPanel draft={draft} plant={plant} editor={editor} request={request} />
      </div>
    </ReactFlowProvider>
  );
}

export function EditorScreen({ editor, navigation, composition }: EditorScreenProps) {
  const state = useStore(editor);
  useEffect(() => {
    void editor.open(composition);
  }, [editor, composition]);
  if (state.phase === "ready" && state.path === composition) {
    return <Editing draft={state} editor={editor} navigation={navigation} />;
  }
  const failed = state.phase === "failed" && state.path === composition;
  return (
    <p className={failed ? plantStyles.failure : plantStyles.loading} role={failed ? "alert" : "status"}>
      {state.phase === "failed" && failed
        ? `The composition could not be opened: ${state.failure.message} (${state.failure.code})`
        : "Opening the composition…"}
    </p>
  );
}
