import type { ReactNode } from "react";
import rowStyles from "./ListSection.module.css";

interface SectionHeadingProps {
  readonly id: string;
  readonly title: string;
  readonly count: number;
  readonly children?: ReactNode;
}

export function SectionHeading({ id, title, count, children }: SectionHeadingProps) {
  return (
    <header className={rowStyles.heading}>
      <h2 id={id}>{title}</h2>
      <span className={rowStyles.count}>{count}</span>
      {children}
    </header>
  );
}
