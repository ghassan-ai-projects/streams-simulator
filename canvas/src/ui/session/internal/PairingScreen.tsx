import { KeyRound, Lock, XCircle } from "lucide-react";
import type { SubmitEvent } from "react";
import type { SessionFailure } from "../../../app/session";
import type { Session, SessionState } from "../../../app/session";
import { pairingBlocked } from "../../../app/session";
import { PAIRING_CODE_LENGTH } from "../../../app/session";
import { cx, shared, useScreen } from "../../primitives";
import { useStore } from "../../primitives";
import styles from "./PairingScreen.module.css";

type Unpaired = Extract<SessionState, { phase: "unpaired" }>;

function pairingAlert(failure: SessionFailure): string {
  if (failure.kind === "code-rejected")
    return "That code did not match. Check the code in your terminal and try again.";
  if (failure.kind === "pairing-closed") {
    return failure.reason === "locked"
      ? "Pairing is locked after five wrong codes. Restart streamsim canvas for a new code."
      : "This code was already used. Restart streamsim canvas to pair another browser.";
  }
  return `${failure.diagnostic.message} (${failure.diagnostic.displayCode})`;
}

function grouped(code: string): string {
  return code.replace(/(.{4})(?=.)/g, "$1-");
}

function Status({ state }: { readonly state: Unpaired }) {
  if (state.failure === null) {
    return (
      <span className={styles.hint}>{`${String(state.code.length)} of ${String(PAIRING_CODE_LENGTH)} characters`}</span>
    );
  }
  const Icon = pairingBlocked(state.failure) ? Lock : XCircle;
  return (
    <span className={styles.error} role="alert">
      <Icon size={14} aria-hidden="true" />
      {pairingAlert(state.failure)}
    </span>
  );
}

function CodeField({ session, state }: { readonly session: Session; readonly state: Unpaired }) {
  const blocked = pairingBlocked(state.failure);
  return (
    <>
      <label className={styles.label} htmlFor="pairing-code">
        Pairing code
      </label>
      <input
        id="pairing-code"
        className={styles.code}
        value={grouped(state.code)}
        onChange={(event) => {
          session.type(event.target.value);
        }}
        disabled={blocked}
        aria-invalid={state.failure !== null && !blocked}
        aria-describedby="pairing-status"
        autoComplete="one-time-code"
        spellCheck={false}
        autoFocus
        placeholder="XXXX-XXXX-XXXX"
      />
      <div id="pairing-status" className={styles.status}>
        <Status state={state} />
      </div>
    </>
  );
}

export function PairingScreen({ session }: { readonly session: Session }) {
  const state = useStore(session);
  useScreen("Pair this browser");
  if (state.phase !== "unpaired") {
    return null;
  }
  const submit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();
    void session.pair();
  };
  return (
    <main className={styles.backdrop}>
      <form className={styles.card} onSubmit={submit} aria-labelledby="pair-title">
        <h1 id="pair-title" className={styles.title}>
          <KeyRound size={20} aria-hidden="true" /> Pair this browser
        </h1>
        <p className={styles.lead}>
          Type the 12-character code that <code>streamsim canvas</code> printed in your terminal. The code works once;
          this browser then stays paired for 12 hours.
        </p>
        <CodeField session={session} state={state} />
        <div className={styles.actions}>
          <button type="submit" className={cx(shared.primary, styles.pair)} disabled={!session.canPair()}>
            {state.busy ? "Pairing…" : "Pair"}
          </button>
          <span className={styles.hint}>Five wrong codes lock pairing until the bridge restarts.</span>
        </div>
      </form>
    </main>
  );
}
