import { useRef, useState } from "react";
import {
  chosenInstrumentWords,
  instrumentCount,
  instrumentWords,
  noInstrument,
  searchInstruments,
} from "../../../model/devices";
import type { Instrument } from "../../../model/devices";
import { panel, shared, useFocusOnMount } from "../../primitives";
import tw from "./Behaviour.module.css";
import type { Aria } from "./Field";

interface ListingProps {
  readonly metric: string;
  readonly value: string;
  readonly instruments: readonly Instrument[];
  readonly onPick: (ref: string) => void;
  readonly onDone: () => void;
}

interface SearchProps {
  readonly metric: string;
  readonly query: string;
  readonly onQuery: (query: string) => void;
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

function Options({ found }: { readonly found: readonly Instrument[] }) {
  return (
    <>
      <option value={noInstrument}>No instrument: a plain reading</option>
      {found.map((entry) => (
        <option key={entry.ref} value={entry.ref}>
          {instrumentWords(entry)}
        </option>
      ))}
    </>
  );
}

function Listing({ metric, value, instruments, onPick, onDone }: ListingProps) {
  const [query, setQuery] = useState("");
  const [highlight, setHighlight] = useState(value);
  const found = searchInstruments(instruments, query);
  const listed = highlight === noInstrument || found.some((entry) => entry.ref === highlight);
  const picked = listed ? highlight : noInstrument;
  return (
    <div
      className={tw.listing}
      onKeyDown={(event) => {
        if (event.key === "Escape") onDone();
      }}
    >
      <Search metric={metric} query={query} onQuery={setQuery} />
      <select
        size={Math.min(6, found.length + 1)}
        aria-label={`Instruments for ${metric}`}
        value={picked}
        onChange={(event) => {
          setHighlight(event.target.value);
        }}
      >
        <Options found={found} />
      </select>
      <p className={panel.hint}>{instrumentCount(found.length, query)}</p>
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
    </div>
  );
}

interface ChooserProps extends Omit<ListingProps, "onDone"> {
  readonly metric: string;
}

function Chooser({ metric, value, instruments, onPick }: ChooserProps) {
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
      {open && <Listing metric={metric} value={value} instruments={instruments} onPick={onPick} onDone={done} />}
    </>
  );
}

interface PickerProps {
  readonly metric: string;
  readonly value: string;
  readonly instruments: readonly Instrument[] | null;
  readonly aria: Aria;
  readonly onPick: (ref: string) => void;
}

export function InstrumentPicker({ metric, value, instruments, aria, onPick }: PickerProps) {
  const shown = instruments === null ? value : chosenInstrumentWords(value, instruments);
  return (
    <div className={tw.picker}>
      <input {...aria} type="text" readOnly value={shown} />
      {instruments === null ? (
        <p className={panel.hint}>The instrument list could not be read, so only plain readings can be chosen.</p>
      ) : (
        <Chooser metric={metric} value={value} instruments={instruments} onPick={onPick} />
      )}
    </div>
  );
}
