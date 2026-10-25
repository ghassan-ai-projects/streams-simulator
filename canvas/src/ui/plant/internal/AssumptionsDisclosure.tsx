import { useState } from "react";
import type { AssumptionsSession } from "../../../app/workspace";
import { AssumptionsPanel } from "./AssumptionsPanel";
import { Disclosure } from "./Disclosure";

interface DisclosureProps {
  readonly session: AssumptionsSession;
  readonly path: string;
  readonly twin: string | null;
}

export function AssumptionsDisclosure({ session, path, twin }: DisclosureProps) {
  const [open, setOpen] = useState(false);
  return (
    <Disclosure
      label="Model assumptions"
      open={open}
      onToggle={() => {
        setOpen(!open);
      }}
    >
      <AssumptionsPanel session={session} path={path} twin={twin} />
    </Disclosure>
  );
}
