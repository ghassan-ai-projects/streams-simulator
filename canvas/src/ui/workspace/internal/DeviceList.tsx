import { Cable, Cpu, Pencil, Plus } from "lucide-react";
import type { Ref } from "react";
import { useEffect, useRef, useState } from "react";
import type { DeviceDesk, DeviceRow } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { kindFacts, stemOf } from "../../../model/devices";
import type { Lineage, Workspace } from "../../../model/workspace";
import { deviceBadge } from "../../../model/workspace";
import { shared } from "../../primitives";
import badgeStyles from "./Badge.module.css";
import { NewDeviceForm } from "./NewDevice";
import { SectionHeading } from "./SectionHeading";
import rowStyles from "./ListSection.module.css";
import workspaceStyles from "./WorkspaceScreen.module.css";

interface DeviceListProps {
  readonly workspace: Workspace;
  readonly navigation: NavigationStore;
  readonly desk: DeviceDesk;
  readonly lineage: Lineage | null;
}

interface Shown {
  readonly row: DeviceRow;
  readonly loaded: boolean;
}

function useDeviceRows(desk: DeviceDesk, workspace: Workspace): readonly DeviceRow[] | null {
  const [rows, setRows] = useState<readonly DeviceRow[] | null>(null);
  const { kinds, domains } = workspace;
  useEffect(() => {
    let live = true;
    void desk
      .rows(
        kinds,
        domains.map((domain) => domain.path),
      )
      .then((next) => {
        if (live) setRows(next);
      });
    return () => {
      live = false;
    };
  }, [desk, kinds, domains]);
  return rows;
}

function shownRows(kinds: Workspace["kinds"], rows: readonly DeviceRow[] | null): readonly Shown[] {
  const shown = kinds.map((entry): Shown => {
    const found = rows?.find((row) => row.path === entry.path);
    if (found !== undefined) return { row: found, loaded: true };
    const name = entry.kind === "" ? stemOf(entry.path) : entry.kind;
    return { row: { path: entry.path, name, metrics: null, commands: null, status: "invalid" }, loaded: false };
  });
  return shown.sort((a, b) => a.row.name.localeCompare(b.row.name));
}

function EditDevice({ row, navigation }: { readonly row: DeviceRow; readonly navigation: NavigationStore }) {
  return (
    <span className={rowStyles.actions}>
      <button
        type="button"
        className={shared.secondary}
        aria-label={`Edit ${row.name}`}
        onClick={() => {
          navigation.go({ screen: "device", path: row.path });
        }}
      >
        <Pencil size={13} aria-hidden="true" />
        Edit
      </button>
    </span>
  );
}

function DeviceItem({
  item,
  navigation,
  lineage,
}: {
  readonly item: Shown;
  readonly navigation: NavigationStore;
  readonly lineage: Lineage | null;
}) {
  const { row, loaded } = item;
  const badge = deviceBadge(row.status, row.path, lineage);
  return (
    <li className={rowStyles.item} title={row.path}>
      <Cpu size={16} aria-hidden="true" className={rowStyles.icon} />
      <span className={rowStyles.lines}>
        <span className={rowStyles.line}>
          <span className={rowStyles.name} title={row.name}>
            {row.name}
          </span>
          {loaded && <span className={badgeStyles[badge.tone]}>{badge.words}</span>}
        </span>
        <span className={workspaceStyles.meta}>
          {loaded ? kindFacts(row.metrics, row.commands) : "Reading the file…"}
        </span>
      </span>
      <EditDevice row={row} navigation={navigation} />
    </li>
  );
}

function NoDevices() {
  return (
    <p className={rowStyles.note}>
      No devices yet. A device describes what a board reports (its readings) and what it accepts (its commands). Create
      one to define your own.
    </p>
  );
}

function useReturnToCapture(desk: DeviceDesk) {
  const button = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (desk.capture().comingBack()) button.current?.focus();
  }, [desk]);
  return button;
}

interface HeadingActionsProps extends Pick<DeviceListProps, "desk" | "navigation"> {
  readonly creating: boolean;
  readonly trigger: Ref<HTMLButtonElement>;
  readonly onToggle: () => void;
}

function HeadingActions({ creating, trigger, desk, navigation, onToggle }: HeadingActionsProps) {
  const captureButton = useReturnToCapture(desk);
  const startCapture = () => {
    desk.capture().clear();
    navigation.go({ screen: "capture" });
  };
  return (
    <span className={rowStyles.actions}>
      <button type="button" ref={captureButton} className={shared.secondary} onClick={startCapture}>
        <Cable size={14} aria-hidden="true" />
        Capture from a board
      </button>
      <button type="button" ref={trigger} className={shared.primary} aria-expanded={creating} onClick={onToggle}>
        <Plus size={14} aria-hidden="true" />
        Create device
      </button>
    </span>
  );
}

export function DeviceList({ workspace, navigation, desk, lineage }: DeviceListProps) {
  const rows = useDeviceRows(desk, workspace);
  const [creating, setCreating] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const taken = workspace.kinds.flatMap((entry) => [entry.kind, stemOf(entry.path)]).filter((name) => name !== "");
  const close = () => {
    setCreating(false);
    trigger.current?.focus();
  };
  return (
    <section className={rowStyles.section} aria-labelledby="devices-title">
      <SectionHeading id="devices-title" title="Devices" count={workspace.kinds.length}>
        <HeadingActions
          creating={creating}
          trigger={trigger}
          desk={desk}
          navigation={navigation}
          onToggle={() => {
            setCreating(!creating);
          }}
        />
      </SectionHeading>
      {creating && <NewDeviceForm taken={taken} desk={desk} navigation={navigation} onCancel={close} />}
      {workspace.kinds.length === 0 ? (
        <NoDevices />
      ) : (
        <ul className={rowStyles.rows}>
          {shownRows(workspace.kinds, rows).map((item) => (
            <DeviceItem key={item.row.path} item={item} navigation={navigation} lineage={lineage} />
          ))}
        </ul>
      )}
    </section>
  );
}
