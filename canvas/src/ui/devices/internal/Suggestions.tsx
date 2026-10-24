import { suggestedRefs, suggestionNoteOf, choiceId } from "../../../app/devices";
import type { TwinDraft, TwinSession } from "../../../app/devices";
import { chosenWords, suggestedWords } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import prov from "./Provenance.module.css";

export function ChoiceTag({ suggested }: { readonly suggested: boolean }) {
  return <span className={suggested ? prov.tagSuggested : prov.tag}>{suggested ? suggestedWords : chosenWords}</span>;
}

export function SuggestionNote({ draft, twin }: { readonly draft: TwinDraft; readonly twin: TwinSession }) {
  const note = suggestionNoteOf(draft);
  if (note === null) return null;
  return (
    <section className={prov.review} aria-label="Suggested choices">
      <p className={prov.reviewText}>{note}</p>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          twin.review(suggestedRefs(draft).map(choiceId));
        }}
      >
        Keep the suggestions
      </button>
    </section>
  );
}
