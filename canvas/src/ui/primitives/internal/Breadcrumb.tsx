import { ChevronRight } from "lucide-react";
import type { Ref } from "react";
import { createContext, useContext } from "react";
import { createPortal } from "react-dom";
import shared from "./shared.module.css";
import styles from "./Breadcrumb.module.css";

export interface Crumb {
  readonly label: string;
  readonly onGo: () => void;
}

export const CrumbSlot = createContext<HTMLElement | null>(null);

interface BreadcrumbProps {
  readonly trail: readonly Crumb[];
  readonly here: string;
  readonly heading: Ref<HTMLHeadingElement>;
  readonly tag?: string;
  readonly caption?: string;
}

function Here({ here, heading, caption }: Omit<BreadcrumbProps, "trail" | "tag">) {
  return (
    <li className={styles.hereItem}>
      {caption !== undefined && (
        <span id="crumb-caption" className={styles.caption} title={caption}>
          {caption}
        </span>
      )}
      <h1
        aria-current="location"
        aria-label={here}
        aria-describedby={caption === undefined ? undefined : "crumb-caption"}
        className={styles.here}
        ref={heading}
        tabIndex={-1}
        title={here}
      >
        {here}
      </h1>
    </li>
  );
}

function Trail({ trail, here, heading, tag, caption }: BreadcrumbProps) {
  return (
    <nav className={styles.crumbs} aria-label="Breadcrumb">
      <ol>
        {trail.map((crumb, index) => (
          <li key={String(index)} className={styles.step}>
            <button
              type="button"
              className={shared.crumb}
              title={crumb.label}
              aria-label={crumb.label}
              onClick={crumb.onGo}
            >
              {crumb.label}
            </button>
            <ChevronRight size={14} aria-hidden="true" />
          </li>
        ))}
        <Here here={here} heading={heading} {...(caption === undefined ? {} : { caption })} />
      </ol>
      {tag !== undefined && <span className={styles.tag}>{tag}</span>}
    </nav>
  );
}

export function Breadcrumb(props: BreadcrumbProps) {
  const slot = useContext(CrumbSlot);
  return slot === null ? <Trail {...props} /> : createPortal(<Trail {...props} />, slot);
}
