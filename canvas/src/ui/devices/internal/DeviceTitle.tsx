import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import { Breadcrumb, useScreen } from "../../primitives";

interface TitleProps {
  readonly path: string;
  readonly tag: string;
  readonly navigation: NavigationStore;
}

export function DeviceTitle({ path, tag, navigation }: TitleProps) {
  const name = stemOf(path);
  const heading = useScreen(`${name} · device`);
  const trail = [
    {
      label: "Workspace",
      onGo: () => {
        navigation.go({ screen: "workspace" });
      },
    },
  ];
  return <Breadcrumb trail={trail} here={name} heading={heading} tag={tag} />;
}
