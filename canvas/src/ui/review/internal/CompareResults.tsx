import type { AlarmLine, Comparison, KpiLine, SystemLine, Table } from "../../../model/review";
import { alarmSystemTable, alarmTable, kpiTable, windowedWords } from "../../../model/review";
import styles from "./Compare.module.css";

type Unit = (entity: string, state: string) => string;

function Frame({
  title,
  caption,
  head,
  children,
}: {
  readonly title: string;
  readonly caption: string;
  readonly head: readonly string[];
  readonly children: React.ReactNode;
}) {
  return (
    <section className={styles.tableBlock} aria-label={title}>
      <h4 className={styles.laneTitle}>{title}</h4>
      <div className={styles.scroll}>
        <table className={styles.table}>
          <caption className={styles.caption}>{caption}</caption>
          <thead>
            <tr>
              {head.map((name, index) => (
                <th key={name} scope="col" className={index < 2 ? undefined : styles.number}>
                  {name}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>{children}</tbody>
        </table>
      </div>
    </section>
  );
}

function KpiRow({ line }: { readonly line: KpiLine }) {
  return (
    <tr data-note={line.note === "" ? undefined : line.note}>
      <th scope="row" className={styles.plain}>
        <span className={styles.entity}>{line.entity}</span>
        {` ${line.kpi}`}
        {line.note !== "" && <span className={styles.tag}>{line.note}</span>}
      </th>
      <td className={styles.unit}>{line.unit}</td>
      <td className={styles.number}>{line.reference}</td>
      <td className={styles.number}>{line.thisRun}</td>
      <td className={styles.number}>{line.delta}</td>
      <td className={styles.number}>{line.relative}</td>
    </tr>
  );
}

function Empty({ columns, children }: { readonly columns: number; readonly children: string }) {
  return (
    <tr>
      <td colSpan={columns} className={styles.muted}>
        {children}
      </td>
    </tr>
  );
}

function Kpis({ table }: { readonly table: Table<KpiLine> }) {
  const head = ["KPI", "Unit", "Reference", "This run", "Change", "Relative change"];
  return (
    <Frame title="KPIs" caption={table.caption} head={head}>
      {table.rows.length === 0 && <Empty columns={6}>No KPI results in either run.</Empty>}
      {table.rows.map((line) => (
        <KpiRow key={line.key} line={line} />
      ))}
    </Frame>
  );
}

function AlarmRow({ line }: { readonly line: AlarmLine }) {
  return (
    <tr>
      <th scope="row" className={styles.plain}>
        <span className={styles.entity}>{line.entity}</span>
        {` ${line.alarm}`}
        {line.note !== "" && <span className={styles.tag}>{line.note}</span>}
      </th>
      <td className={styles.number}>{line.activations.reference}</td>
      <td className={styles.number}>{line.activations.thisRun}</td>
      <td className={styles.number}>{line.activations.delta}</td>
      <td className={styles.number}>{line.active.reference}</td>
      <td className={styles.number}>{line.active.thisRun}</td>
      <td className={styles.number}>{line.active.delta}</td>
    </tr>
  );
}

function Alarms({ table }: { readonly table: Table<AlarmLine> }) {
  const head = [
    "Alarm",
    "Raised: reference",
    "Raised: this run",
    "Change",
    "Active: reference",
    "Active: this run",
    "Change",
  ];
  return (
    <Frame title="Alarms" caption={table.caption} head={head}>
      {table.rows.length === 0 && <Empty columns={7}>No alarm in either run.</Empty>}
      {table.rows.map((line) => (
        <AlarmRow key={line.key} line={line} />
      ))}
    </Frame>
  );
}

function SystemRow({ line }: { readonly line: SystemLine }) {
  return (
    <tr>
      <th scope="row" className={styles.plain}>
        {line.measure}
      </th>
      <td className={styles.number}>{line.reference}</td>
      <td className={styles.number}>{line.thisRun}</td>
      <td className={styles.number}>{line.delta}</td>
    </tr>
  );
}

function AlarmSystem({ table }: { readonly table: Table<SystemLine> }) {
  return (
    <Frame title="Alarm system" caption={table.caption} head={["Measure", "Reference", "This run", "Change"]}>
      {table.rows.map((line) => (
        <SystemRow key={line.key} line={line} />
      ))}
    </Frame>
  );
}

export function CompareResults({ comparison, unit }: { readonly comparison: Comparison; readonly unit: Unit }) {
  return (
    <section className={styles.block} aria-label="Whole-run results">
      <h3 className={styles.blockTitle}>Whole-run results</h3>
      <p className={styles.muted}>Change is this run minus the reference. Nothing here ranks the runs.</p>
      <Kpis table={kpiTable(comparison, unit)} />
      <Alarms table={alarmTable(comparison)} />
      <AlarmSystem table={alarmSystemTable(comparison)} />
      <p className={styles.windowed} data-part="windowed">
        {windowedWords(comparison)}
      </p>
    </section>
  );
}
