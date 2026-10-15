import type { EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import { useStore } from "../useStore";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./Editor.module.css";
import leave from "./LeaveDialog.module.css";

interface Props {
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
}

function LeaveButtons({ editor, navigation }: Props) {
  const stay = useFocusOnMount<HTMLButtonElement>();
  const saveAndLeave = async () => {
    await editor.save();
    if (editor.unsavedChanges()) navigation.stay();
    else navigation.leave();
  };
  const discardAndLeave = () => {
    editor.revert();
    navigation.leave();
  };
  return (
    <div className={leave.actions}>
      <button
        type="button"
        ref={stay}
        className={styles.action}
        onClick={() => {
          navigation.stay();
        }}
      >
        Stay
      </button>
      <button type="button" className={styles.danger} onClick={discardAndLeave}>
        Discard and leave
      </button>
      <button
        type="button"
        className={styles.save}
        onClick={() => {
          void saveAndLeave();
        }}
      >
        Save and leave
      </button>
    </div>
  );
}

export function LeaveDialog({ title, editor, navigation }: Props & { readonly title: string }) {
  return (
    <div className={leave.scrim}>
      <dialog
        open
        className={leave.dialog}
        aria-modal="true"
        aria-labelledby="leave-title"
        aria-describedby="leave-text"
        onKeyDown={(event) => {
          if (event.key === "Escape") navigation.stay();
        }}
      >
        <h2 id="leave-title" className={leave.title}>
          Leave with unsaved changes?
        </h2>
        <p id="leave-text" className={leave.text}>
          {`${title} has changes that are not saved. Save them and leave, discard them, or stay and keep editing.`}
        </p>
        <LeaveButtons editor={editor} navigation={navigation} />
      </dialog>
    </div>
  );
}

export function PendingLeave({ title, editor, navigation }: Props & { readonly title: string }) {
  const leaving = useStore(navigation.leaving);
  if (leaving === null) return null;
  return <LeaveDialog title={title} editor={editor} navigation={navigation} />;
}
