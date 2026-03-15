import type { GiveawayEntry } from "./types";

export function isAutoGiveawayStatus(entry: Pick<GiveawayEntry, "type">) {
  return entry.type === "1v1";
}

export function resolveGiveawayStatus(
  entry: Pick<GiveawayEntry, "type" | "status">,
  currentBotModeKey: string,
): GiveawayEntry["status"] {
  if (entry.type === "1v1") {
    return currentBotModeKey === "1v1" ? "live" : "ready";
  }

  return entry.status;
}
