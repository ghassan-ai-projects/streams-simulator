import { useEffect } from "react";
import type { DeviceSession } from "../../../app/devices";
import { canSave } from "../../../app/devices";

export function useDeviceKeys(devices: DeviceSession) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const command = (event.ctrlKey || event.metaKey) && !event.altKey && event.key.toLowerCase() === "s";
      if (!command) return;
      event.preventDefault();
      if (canSave(devices.snapshot())) void devices.save();
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [devices]);
}
