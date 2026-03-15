import { Button, Paper, Stack, Typography } from "@mui/material";

export function DashboardNotFoundPage() {
  return (
    <Paper
      elevation={0}
      sx={{
        p: 3,
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.paper",
      }}
    >
      <Typography
        sx={{
          fontSize: "0.78rem",
          fontWeight: 800,
          letterSpacing: "0.12em",
          textTransform: "uppercase",
          color: "primary.main",
        }}
      >
        404 Dashboard Page
      </Typography>
      <Typography variant="h4" sx={{ mt: 1.1, fontWeight: 800 }}>
        This dashboard page was not found
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 1.5, maxWidth: 680, lineHeight: 1.75 }}>
        That route is not part of the moderator dashboard right now. Use the sidebar or jump back
        to the dashboard overview.
      </Typography>

      <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap sx={{ mt: 3 }}>
        <Button href="/dashboard" variant="contained" sx={{ minHeight: 42, px: 1.8 }}>
          Dashboard Overview
        </Button>
        <Button href="/" variant="outlined" sx={{ minHeight: 42, px: 1.8 }}>
          Public Site
        </Button>
      </Stack>
    </Paper>
  );
}
