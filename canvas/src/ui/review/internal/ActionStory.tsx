import type { ActionStory as Story } from "../../../model/review";
import { humanise } from "../../../model/shared/words";
import { shared } from "../../primitives";
import { Section } from "./ReportParts";
import styles from "./ActionStory.module.css";

export interface StoryActions {
  readonly showBefore: (story: Story) => void;
  readonly showAfter: (story: Story) => void;
  readonly hide: () => void;
}

function Arguments({ story }: { readonly story: Story }) {
  if (story.args.length === 0) return <>No arguments</>;
  return (
    <ul className={styles.args}>
      {story.args.map((arg) => (
        <li key={arg.name}>
          <span className={styles.name}>{arg.name}</span>
          {` = ${arg.value}`}
        </li>
      ))}
    </ul>
  );
}

function sourceWords(story: Story): string | null {
  if (story.source === null) return null;
  return `${story.applied ? "Applied" : "Requested"} ${story.source}`;
}

function Facts({ story }: { readonly story: Story }) {
  const source = sourceWords(story);
  return (
    <dl className={styles.facts}>
      <dt>Request</dt>
      <dd>{`${humanise(story.effector)} on ${story.entity}`}</dd>
      <dt>Arguments</dt>
      <dd>
        <Arguments story={story} />
      </dd>
      <dt>Requested at</dt>
      <dd>{story.requestedAt}</dd>
      <dt>Outcome</dt>
      <dd>{story.outcome}</dd>
      {source !== null && (
        <>
          <dt>Source</dt>
          <dd>{source}</dd>
        </>
      )}
    </dl>
  );
}

export function ActionStory({ story, actions }: { readonly story: Story; readonly actions: StoryActions }) {
  return (
    <Section title="Action story" note={`${story.entity} · ${humanise(story.effector)}`}>
      <Facts story={story} />
      <div className={styles.actions}>
        <button
          type="button"
          className={shared.secondary}
          onClick={() => {
            actions.showBefore(story);
          }}
        >
          Show 1 min before
        </button>
        <button
          type="button"
          className={shared.secondary}
          onClick={() => {
            actions.showAfter(story);
          }}
        >
          Show 1 min after
        </button>
        <button type="button" className={shared.secondary} onClick={actions.hide}>
          Hide
        </button>
      </div>
      <p className={styles.note}>Moves the cursor only. This view does not say what the call changed.</p>
    </Section>
  );
}
