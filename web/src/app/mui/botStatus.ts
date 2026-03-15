import type { Theme } from "@mui/material/styles";
import type { SystemStyleObject } from "@mui/system";

export function botStatusChipSx(isOnline: boolean): SystemStyleObject<Theme> {
  if (isOnline) {
    return {
      fontWeight: 700,
    };
  }

  return {
    fontWeight: 800,
    color: "#fff7f7",
    bgcolor: "#b42318",
    borderColor: "#f87171",
    boxShadow: "0 0 0 1px rgba(248, 113, 113, 0.26), 0 0 16px rgba(220, 38, 38, 0.26)",
  };
}

export function botStatusTextSx(isOnline: boolean): SystemStyleObject<Theme> {
  if (isOnline) {
    return {
      fontWeight: 700,
    };
  }

  return {
    fontWeight: 800,
    color: "#ef9a9a",
    textShadow: "0 0 10px rgba(255, 107, 107, 0.18)",
  };
}
