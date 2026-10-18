import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { shared, useFocusOnMount } from "../../primitives";
import notes from "./Notes.module.css";

interface ConfirmProps {
  readonly words: string;
  readonly icon: ReactNode;
  readonly title: string;
  readonly question: string;
  readonly confirmWords: string;
  readonly needsConfirm: boolean;
  readonly disabled?: boolean;
  readonly onConfirm: () => void;
}

interface QuestionProps {
  readonly question: string;
  readonly confirmWords: string;
  readonly onAnswer: (confirmed: boolean) => void;
}

function Question({ question, confirmWords, onAnswer }: QuestionProps) {
  const keep = useFocusOnMount<HTMLButtonElement>();
  return (
    <span className={notes.confirm} role="group" aria-label={question}>
      {question}
      <button
        type="button"
        className={shared.primary}
        onClick={() => {
          onAnswer(true);
        }}
      >
        {confirmWords}
      </button>
      <button
        type="button"
        ref={keep}
        className={shared.secondary}
        onClick={() => {
          onAnswer(false);
        }}
      >
        Keep editing
      </button>
    </span>
  );
}

export function ConfirmButton(props: ConfirmProps) {
  const { words, icon, title, question, confirmWords, needsConfirm, disabled = false, onConfirm } = props;
  const [asking, setAsking] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const wasAsking = useRef(false);
  useEffect(() => {
    if (wasAsking.current && !asking) trigger.current?.focus();
    wasAsking.current = asking;
  }, [asking]);
  if (asking && needsConfirm) {
    return (
      <Question
        question={question}
        confirmWords={confirmWords}
        onAnswer={(confirmed) => {
          setAsking(false);
          if (confirmed) onConfirm();
        }}
      />
    );
  }
  return (
    <button
      type="button"
      ref={trigger}
      className={shared.secondary}
      aria-disabled={disabled}
      title={title}
      onClick={() => {
        if (disabled) return;
        if (needsConfirm) setAsking(true);
        else onConfirm();
      }}
    >
      {icon}
      {words}
    </button>
  );
}
