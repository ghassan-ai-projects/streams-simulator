import { Boxes } from "lucide-react";
import { ReactFlowProvider } from "@xyflow/react";
import { useEffect, useMemo, useState } from "react";
import type { Draft, EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import { compositionPlant, shortKind } from "../../model/composition";
import { connect } from "../../model/compositionEdits";
import type { PlantView } from "../../model/plantView";
import { plantView } from "../../model/plantView";
import { Navigator } from "../plant/GraphControls";
import { Outline } from "../plant/Outline";
import { useCanvasAspect } from "../plant/canvasAspect";
import { PlantGraph } from "../plant/PlantGraph";
import plantStyles from "../plant/PlantScreen.module.css";
import { useStore } from "../useStore";
import { EditorBar } from "./EditorBar";
import styles from "./Editor.module.css";
import { EditorPanel } from "./EditorPanel";
import { PendingLeave } from "./LeaveDialog";
import { FocusRequest } from "./panelProps";

interface EditorScreenProps {
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
  readonly composition: string;
}

function usePlant(draft: Draft, aspect: number | undefined) {
  const plant = useMemo(() => compositionPlant(draft.doc, shortKind), [draft.doc]);
  const view = useMemo(() => plantView(plant, draft.specs, null, false, aspect), [plant, draft.specs, aspect]);
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
      <p>
        Find a piece of equipment in the Build panel, give it a name and an area, and add it. Then link an output to an
        input, and save to compose the plant.
      </p>
    </div>
  );
}

interface GraphProps {
  readonly draft: Draft;
  readonly editor: EditorStore;
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
      selected={draft.selected}
      flagged={flagged}
      minFitZoom={0.25}
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
  useLeaveGuard(draft.doc !== draft.saved);
  const select = (id: string | null) => {
    editor.select(id);
  };
  return (
    <ReactFlowProvider>
      <div className={plantStyles.screen}>
        <div className={plantStyles.side}>
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
        <section className={plantStyles.center} aria-label={`Editing ${plant.title || plant.id}`}>
          <EditorBar draft={draft} editor={editor} navigation={navigation} request={request} />
          <div ref={ref} className={plantStyles.graphHost}>
            {plant.instances.length === 0 ? (
              <NoEquipment />
            ) : (
              <EditorGraph draft={draft} editor={editor} view={view} flagged={flagged} aspect={aspect} />
            )}
          </div>
        </section>
        <EditorPanel draft={draft} plant={plant} editor={editor} request={request} />
      </div>
      <PendingLeave title={plant.title || plant.id} editor={editor} navigation={navigation} />
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
