export function formatStreamerTitle(name: string): string {
  const trimmed = name.trim();
  if (trimmed === "") {
    return "DankBot";
  }

  const normalized = trimmed
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();

  if (normalized === "") {
    return "DankBot";
  }

  return normalized
    .split(" ")
    .map((part) => {
      const lower = part.toLowerCase();
      return lower.charAt(0).toUpperCase() + lower.slice(1);
    })
    .join(" ");
}

export function buildSiteTitle(streamerName: string): string {
  const formatted = formatStreamerTitle(streamerName);
  if (formatted === "DankBot") {
    return "DankBot";
  }

  return `DankBot - ${formatted}`;
}
