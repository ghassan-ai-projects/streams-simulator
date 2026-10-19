import { useId, useRef, useState } from "react";
import {
  chosenInstrumentWords,
  instrumentCount,
  instrumentWords,
  noInstrument,
  searchInstruments,
  unfitWords,
  unstatedMember,
} from "../../../model/devices";
import type { Instrument } from "../../../model/devices";
import { cx, panel, shared, useFocusOnMount } from "../../primitives";
import tw from "./Behaviour.module.css";
import type { Aria } from "./Field";
import notes from "./Notes.module.css";

interface ListingProps {
  readonly metric: string;
  readonly value: string;
  readonly instruments: readonly Instrument[];
  readonly reveal: (() => void) | null;
  readonly onPick: (ref: string) => void;
  readonly onDone: () => void;
}

interface SearchProps {
  readonly metric: string;
  readonly query: string;
  readonly onQuery: (query: string) => void;
}

interface ChoiceProps {
  readonly group: string;
  readonly checked: boolean;
  readonly entry: Instrument | null;
  readonly onPick: (ref: string) => void;
}

function Search({ metric, query, onQuery }: SearchProps) {
  const search = useFocusOnMount<HTMLInputElement>();
  return (
    <input
      ref={search}
      type="search"
      aria-label={`Find an instrument for ${metric}`}
      placeholder="Find an instrument: temperature, humidity…"
      value={query}
      onChange={(event) => {
        onQuery(event.target.value);
      }}
      onKeyDown={(event) => {
        if (event.key === "Enter") event.preventDefault();
      }}
    />
  );
}

function Choice({ group, checked, entry, onPick }: ChoiceProps) {
  const ref = entry?.ref ?? noInstrument;
  const reason = entry === null ? "" : unfitWords(entry);
  return (
    <label className={cx(notes.choice, reason !== "" && notes.choiceOff)}>
      <input
        type="radio"
        name={group}
        value={ref}
        checked={checked}
        disabled={reason !== ""}
        onChange={() => {
          onPick(ref);
        }}
      />
      <span className={notes.choiceText}>
        <span>{entry === null ? "No instrument: a plain reading" : instrumentWords(entry)}</span>
        {reason !== "" && <span className={notes.choiceNote}>{reason}</span>}
      </span>
    </label>
  );
}

interface ChoicesProps {
  readonly metric: string;
  readonly picked: string;
  readonly found: readonly Instrument[];
  readonly onPick: (ref: string) => void;
}

function Choices({ metric, picked, found, onPick }: ChoicesProps) {
  const group = useId();
  return (
    <div role="radiogroup" aria-label={`Instruments for ${metric}`} className={notes.choices}>
      <Choice group={group} checked={picked === noInstrument} entry={null} onPick={onPick} />
      {found.map((entry) => (
        <Choice key={entry.ref} group={group} checked={picked === entry.ref} entry={entry} onPick={onPick} />
      ))}
    </div>
  );
}

function Listing({ metric, value, instruments, reveal, onPick, onDone }: ListingProps) {
  const [query, setQuery] = useState("");
  const [highlight, setHighlight] = useState(value);
  const found = searchInstruments(instruments, query);
  const listed = highlight === noInstrument || found.some((entry) => entry.ref === highlight);
  const picked = listed ? highlight : noInstrument;
  const waiting = instruments.some((entry) => unstatedMember(entry) !== null);
  return (
    <div
      className={tw.listing}
      onKeyDown={(event) => {
        if (event.key === "Escape") onDone();
      }}
    >
      <Search metric={metric} query={query} onQuery={setQuery} />
      <Choices metric={metric} picked={picked} found={found} onPick={setHighlight} />
      <p className={panel.hint}>{instrumentCount(found.length, query)}</p>
      <div className={notes.choiceActions}>
        <button
          type="button"
          className={shared.primary}
          onClick={() => {
            onPick(picked);
            onDone();
          }}
        >
          Use this instrument
        </button>
        {waiting && reveal !== null && (
          <button type="button" className={shared.secondary} onClick={reveal}>
            Set it in the Interface step
          </button>
        )}
      </div>
    </div>
  );
}

interface ChooserProps extends Omit<ListingProps, "onDone"> {
  readonly metric: string;
}

function Chooser({ metric, value, instruments, reveal, onPick }: ChooserProps) {
  const [open, setOpen] = useState(false);
  const toggle = useRef<HTMLButtonElement>(null);
  const done = () => {
    setOpen(false);
    toggle.current?.focus();
  };
  return (
    <>
      <button
        ref={toggle}
        type="button"
        className={shared.secondary}
        aria-expanded={open}
        aria-label={`${open ? "Close the list for" : "Change the instrument of"} ${metric}`}
        onClick={() => {
          setOpen(!open);
        }}
      >
        {open ? "Close" : "Change"}
      </button>
      {open && (
        <Listing
          metric={metric}
          value={value}
          instruments={instruments}
          reveal={reveal}
          onPick={onPick}
          onDone={done}
        />
      )}
    </>
  );
}

interface PickerProps {
  readonly metric: string;
  readonly value: string;
  readonly instruments: readonly Instrument[] | null;
  readonly aria: Aria;
  readonly reveal: (() => void) | null;
  readonly onPick: (ref: string) => void;
}

export function InstrumentPicker({ metric, value, instruments, aria, reveal, onPick }: PickerProps) {
  const shown = instruments === null ? value : chosenInstrumentWords(value, instruments);
  return (
    <div className={tw.picker}>
      <textarea {...aria} className={cx(shared.sunkenField, notes.chosen)} readOnly rows={2} value={shown} />
      {instruments === null ? (
        <p className={panel.hint}>The instrument list could not be read, so only plain readings can be chosen.</p>
      ) : (
        <Chooser metric={metric} value={value} instruments={instruments} reveal={reveal} onPick={onPick} />
      )}
    </div>
  );
}
