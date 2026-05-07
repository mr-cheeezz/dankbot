import { Box } from "@mui/material";

import { PublicPollOverlayPage } from "./PublicPollOverlayPage";
import { PublicPredictionOverlayPage } from "./PublicPredictionOverlayPage";

export function PublicOverlayPage() {
  return (
    <Box
      sx={{
        width: "100vw",
        height: "100vh",
        position: "relative",
        overflow: "hidden",
        backgroundColor: "transparent",
      }}
    >
      <PublicPollOverlayPage />
      <PublicPredictionOverlayPage />
    </Box>
  );
}
