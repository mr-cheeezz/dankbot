import {
  Button,
  Paper,
  Stack,
  Typography,
} from "@mui/material";

export function NotFoundPage() {
  return (
    <Paper
      elevation={0}
      sx={{
        maxWidth: 760,
        mx: "auto",
        p: { xs: 3, md: 4 },
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.paper",
      }}
    >
      <Typography
        sx={{
          fontSize: "0.8rem",
          fontWeight: 800,
          letterSpacing: "0.12em",
          textTransform: "uppercase",
          color: "primary.main",
        }}
      >
        404 Not Found
      </Typography>
      <Typography variant="h3" sx={{ mt: 1.25, fontWeight: 800 }}>
        That page does not exist
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 1.5, maxWidth: 620, lineHeight: 1.75 }}>
        The link may be old, the page may have moved, or the URL may have been typed a little
        wrong. You can head back home or jump straight to the commands page.
      </Typography>

      <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap sx={{ mt: 3 }}>
        <Button href="/" variant="contained" sx={{ minHeight: 42, px: 1.8 }}>
          Back Home
        </Button>
        <Button href="/commands" variant="outlined" sx={{ minHeight: 42, px: 1.8 }}>
          Open Commands
        </Button>
      </Stack>
    </Paper>
  );
}
