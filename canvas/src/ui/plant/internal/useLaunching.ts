import { useCallback, useEffect, useRef, useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";

export function useConsume(navigation: NavigationStore, domain: string, hall: string | null) {
  return useCallback(() => {
    navigation.replace({ screen: "plant", domain, hall });
  }, [navigation, domain, hall]);
}

export function useShowRun(launch: LaunchStore, onShow: () => void) {
  return useCallback(() => {
    launch.reset();
    onShow();
  }, [launch, onShow]);
}

export function useLaunching({ asked }: { readonly asked: boolean }, onShow: () => void, consume: () => void) {
  const [launching, setLaunching] = useState(asked);
  const button = useRef<HTMLButtonElement>(null);
  const [seen, setSeen] = useState(asked);
  if (asked !== seen) {
    setSeen(asked);
    if (asked) setLaunching(true);
  }
  useEffect(() => {
    if (!asked) return;
    onShow();
    consume();
  }, [asked, onShow, consume]);
  const toggle = () => {
    if (!launching) onShow();
    setLaunching(!launching);
  };
  const close = () => {
    setLaunching(false);
    button.current?.focus();
  };
  return { launching, button, toggle, close };
}
