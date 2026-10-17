import { screen } from "../../primitives";
import { Boxes } from "lucide-react";
import { ReactFlowProvider } from "@xyflow/react";
import { useEffect, useMemo, useState } from "react";
import { isDirty } from "../../../app/editor";
import type { Draft, EditorSession } from "../../../app/editor";
import type { NavigationStore } from "../../../app/shell";
import { compositionPlant, shortKind } from "../../../model/editor";
import { connect } from "../../../model/editor";
import { plantView } from "../../../model/graph";
import { plantTitle } from "../../../model/plant";
import { failureLine } from "../../../model/shared/words";
import type { PlantView } from "../../../model/graph";
import { Navigator, Outline, PlantGraph, useCanvasAspect } from "../../graph";
import { useStore } from "../../primitives";
import { EditorBar } from "./EditorBar";
import styles from "./Editor.module.css";
import { EditorPanel } from "./EditorPanel";
import { PendingLeave } from "./LeaveDialog";
import { FocusRequest } from "./panelProps";

interface EditorScreenProps {
  readonly editor: EditorSession;
  readonly navigation: NavigationStore;
  readonly composition: string;
}

function usePlant(draft: Draft, aspect: number | undefined) {
  const plant = useMemo(() => compositionPlant(draft.doc, shortKind), [draft.doc]);
  const view = useMemo(() => plantView(plant, draft.specs, { open: null, aspect }), [plant, draft.specs, aspect]);
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

function NoEquipment() {
  return (
    <div className={styles.empty}>
      <Boxes size={28} aria-hidden="true" />
      <h2>No equipment yet</h2>
      <p>Equipment you add from the Build panel appears here.</p>
    </div>
  );
}

interface GraphProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
  readonly view: PlantView;
  readonly flagged: ReadonlySet<string>;
  readonly aspect: number | undefined;
}

function EditorGraph({ draft, editor, view, flagged, aspect }: GraphProps) {
  return (
    <PlantGraph
      view={view}
      viewKey={`edit:${draft.path}`}
      fitKey={`${draft.path}:${String(view.nodes.length)}:${String(aspect)}`}
      panCue
      selected={draft.selected}
      flagged={flagged}
      focusZoom={1}
      onSelect={(id) => {
        editor.select(id);
      }}
      onConnect={(from, to) => {
        editor.edit((doc) => connect(doc, from, to));
      }}
    />
  );
}

function Editing({ draft, editor, navigation }: { readonly draft: Draft } & Omit<EditorScreenProps, "composition">) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const { plant, view, flagged } = usePlant(draft, aspect);
  const [request] = useState(() => new FocusRequest());
  useLeaveGuard(isDirty(draft));
  const select = (id: string | null) => {
    editor.select(id);
  };
  return (
    <ReactFlowProvider>
      <div className={screen.screen}>
        <div className={screen.side}>
          <Outline
            plant={plant}
            hall={null}
            selected={draft.selected}
            onSelect={select}
            flagged={flagged}
            showEmptyAreas
          />
          {plant.instances.length > 0 ? <Navigator /> : null}
        </div>
        <section className={screen.center} aria-label={`Editing ${plantTitle(plant)}`}>
          <EditorBar draft={draft} editor={editor} navigation={navigation} request={request} />
          <div ref={ref} className={screen.graphHost}>
            {plant.instances.length === 0 ? (
              <NoEquipment />
            ) : (
              <EditorGraph draft={draft} editor={editor} view={view} flagged={flagged} aspect={aspect} />
            )}
          </div>
        </section>
        <EditorPanel draft={draft} plant={plant} editor={editor} request={request} />
      </div>
      <PendingLeave title={plantTitle(plant)} editor={editor} navigation={navigation} />
    </ReactFlowProvider>
  );
}

export function EditorScreen({ editor, navigation, composition }: EditorScreenProps) {
  const state = useStore(editor);
  useEffect(() => {
    const disposeGuard = navigation.guardUnsaved(() => editor.unsavedChanges());
    void editor.open(composition);
    return () => {
      disposeGuard();
      editor.close();
    };
  }, [editor, navigation, composition]);
  if (state.phase === "ready" && state.path === composition) {
    return <Editing draft={state} editor={editor} navigation={navigation} />;
  }
  const failed = state.phase === "failed" && state.path === composition;
  return (
    <p className={failed ? screen.failure : screen.loading} role={failed ? "alert" : "status"}>
      {state.phase === "failed" && failed
        ? `The composition could not be opened: ${failureLine(state.failure.diagnostic)}`
        : "Opening the composition…"}
    </p>
  );
}
