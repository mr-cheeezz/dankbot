import ContentCopyRoundedIcon from "@mui/icons-material/ContentCopyRounded";
import FormatQuoteRoundedIcon from "@mui/icons-material/FormatQuoteRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import ShuffleRoundedIcon from "@mui/icons-material/ShuffleRounded";
import TagRoundedIcon from "@mui/icons-material/TagRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  IconButton,
  InputAdornment,
  Snackbar,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import Alert from "@mui/material/Alert";
import { useEffect, useMemo, useState } from "react";

import { defaultPublicQuotes, fetchPublicQuotes, type PublicQuote } from "./api";

function normalizeQuoteQuery(value: string) {
  return value.trim().toLowerCase();
}

function pickRandomQuote(items: PublicQuote[], excludeID?: number): PublicQuote | null {
  if (items.length === 0) {
    return null;
  }
  if (items.length === 1) {
    return items[0];
  }

  const pool =
    excludeID == null ? items : items.filter((entry) => entry.id !== excludeID);
  const target = pool.length > 0 ? pool : items;
  return target[Math.floor(Math.random() * target.length)] ?? target[0] ?? null;
}

export function PublicQuotesPage() {
  const [items, setItems] = useState<PublicQuote[]>(defaultPublicQuotes);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [featuredQuote, setFeaturedQuote] = useState<PublicQuote | null>(null);
  const [copiedMessage, setCopiedMessage] = useState("");

  useEffect(() => {
    const controller = new AbortController();

    fetchPublicQuotes(controller.signal)
      .then((nextItems) => {
        setItems(nextItems);
        setFeaturedQuote(pickRandomQuote(nextItems));
      })
      .catch(() => {
        setItems(defaultPublicQuotes);
        setFeaturedQuote(pickRandomQuote(defaultPublicQuotes));
      })
      .finally(() => {
        setLoading(false);
      });

    return () => controller.abort();
  }, []);

  const filteredItems = useMemo(() => {
    const normalized = normalizeQuoteQuery(query);
    if (normalized === "") {
      return items;
    }

    const numericTarget = normalized.replace(/^#/, "");

    return items.filter((entry) => {
      if (`${entry.id}` === numericTarget) {
        return true;
      }
      return `#${entry.id} ${entry.message}`.toLowerCase().includes(normalized);
    });
  }, [items, query]);

  const quoteCountLabel = loading
    ? "loading saved quotes..."
    : filteredItems.length === items.length
      ? `${items.length} saved quotes`
      : `${filteredItems.length} matching quotes`;

  const rotateFeaturedQuote = () => {
    const next = pickRandomQuote(items, featuredQuote?.id);
    if (next != null) {
      setFeaturedQuote(next);
    }
  };

  const copyQuote = async (quote: PublicQuote) => {
    try {
      await navigator.clipboard.writeText(`#${quote.id} ${quote.message}`);
      setCopiedMessage(`Copied quote #${quote.id}.`);
    } catch {
      setCopiedMessage("Could not copy that quote right now.");
    }
  };

  return (
    <>
      <Stack spacing={2.5}>
        <Card
          sx={{
            background:
              "linear-gradient(180deg, rgba(74,137,255,0.12) 0%, rgba(255,255,255,0.02) 100%)",
          }}
        >
          <CardContent sx={{ p: 2.75 }}>
            <Stack
              direction={{ xs: "column", lg: "row" }}
              spacing={2}
              justifyContent="space-between"
              alignItems={{ xs: "flex-start", lg: "center" }}
            >
              <Box>
                <Stack direction="row" spacing={1} alignItems="center">
                  <FormatQuoteRoundedIcon color="primary" />
                  <Typography variant="h4">Quotes</Typography>
                </Stack>
                <Typography variant="body1" color="text.secondary" sx={{ mt: 0.85, maxWidth: 760 }}>
                  Saved stream quotes in one place, so viewers can browse the best lines without
                  turning chat into a quote dump.
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.4 }}>
                  <Chip color="primary" variant="outlined" label={quoteCountLabel} />
                  <Chip
                    variant="outlined"
                    label={
                      loading
                        ? "Quote IDs ready soon"
                        : items.length > 0
                          ? `IDs #${Math.min(...items.map((entry) => entry.id))}-${Math.max(...items.map((entry) => entry.id))}`
                          : "No quote IDs yet"
                    }
                  />
                </Stack>
              </Box>

              <Button
                variant="contained"
                startIcon={<ShuffleRoundedIcon />}
                onClick={rotateFeaturedQuote}
                disabled={loading || items.length === 0}
              >
                Random quote
              </Button>
            </Stack>
          </CardContent>
        </Card>

        {featuredQuote != null ? (
          <Card>
            <CardContent sx={{ p: 2.5 }}>
              <Stack
                direction={{ xs: "column", md: "row" }}
                justifyContent="space-between"
                alignItems={{ xs: "flex-start", md: "center" }}
                spacing={1.5}
              >
                <Box>
                  <Typography variant="h6">Featured quote</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.4 }}>
                    A quick pull from the archive when you just want one good line.
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1}>
                  <Chip color="primary" variant="outlined" label={`#${featuredQuote.id}`} />
                  <Tooltip title="Copy quote" arrow>
                    <span>
                      <IconButton
                        onClick={() => void copyQuote(featuredQuote)}
                        aria-label={`Copy quote ${featuredQuote.id}`}
                      >
                        <ContentCopyRoundedIcon fontSize="small" />
                      </IconButton>
                    </span>
                  </Tooltip>
                </Stack>
              </Stack>

              <Box
                sx={{
                  mt: 1.75,
                  p: 2,
                  borderRadius: 2,
                  border: "1px solid",
                  borderColor: "divider",
                  backgroundColor: "rgba(255,255,255,0.03)",
                }}
              >
                <Typography
                  sx={{
                    fontSize: { xs: "1.05rem", md: "1.15rem" },
                    lineHeight: 1.75,
                    fontWeight: 700,
                  }}
                >
                  “{featuredQuote.message}”
                </Typography>
              </Box>
            </CardContent>
          </Card>
        ) : null}

        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Stack
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              alignItems={{ xs: "stretch", md: "center" }}
              spacing={1.5}
            >
              <Box>
                <Typography variant="h6">Quotes archive</Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.4 }}>
                  Search by quote text or jump straight to a quote number like <strong>#42</strong>.
                </Typography>
              </Box>
              <TextField
                size="small"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search quotes or enter #id"
                sx={{ width: { xs: "100%", md: 340 } }}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchRoundedIcon fontSize="small" />
                    </InputAdornment>
                  ),
                }}
              />
            </Stack>
          </CardContent>
        </Card>

        {filteredItems.length > 0 ? (
          <Stack spacing={1.5}>
            {filteredItems.map((quote) => (
              <Card key={quote.id}>
                <CardContent sx={{ p: 2.1 }}>
                  <Stack direction="row" spacing={1.5} alignItems="flex-start">
                    <Box
                      sx={{
                        width: 4,
                        alignSelf: "stretch",
                        borderRadius: 999,
                        background:
                          "linear-gradient(180deg, rgba(74,137,255,0.9) 0%, rgba(74,137,255,0.25) 100%)",
                        flexShrink: 0,
                      }}
                    />

                    <Stack spacing={1.1} sx={{ minWidth: 0, flex: 1 }}>
                      <Stack
                        direction={{ xs: "column", sm: "row" }}
                        justifyContent="space-between"
                        alignItems={{ xs: "flex-start", sm: "center" }}
                        spacing={1}
                      >
                        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                          <Chip
                            icon={<TagRoundedIcon />}
                            label={`#${quote.id}`}
                            color="primary"
                            variant="outlined"
                          />
                          <Chip
                            size="small"
                            variant="outlined"
                            label={`${quote.message.length} chars`}
                          />
                        </Stack>

                        <Tooltip title="Copy quote" arrow>
                          <span>
                            <IconButton
                              onClick={() => void copyQuote(quote)}
                              aria-label={`Copy quote ${quote.id}`}
                            >
                              <ContentCopyRoundedIcon fontSize="small" />
                            </IconButton>
                          </span>
                        </Tooltip>
                      </Stack>

                      <Typography sx={{ fontSize: "1rem", lineHeight: 1.75 }}>
                        {quote.message}
                      </Typography>
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Stack>
        ) : (
          <Card>
            <CardContent sx={{ p: 2.5 }}>
              <Typography variant="h6">
                {loading ? "Loading quotes..." : query.trim() === "" ? "No quotes yet" : "No matches"}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
                {loading
                  ? "The archive is still loading."
                  : query.trim() === ""
                    ? "There are no saved quotes yet."
                    : "No quotes matched that search. Try a quote number or a different phrase."}
              </Typography>
            </CardContent>
          </Card>
        )}
      </Stack>

      <Snackbar
        open={copiedMessage !== ""}
        autoHideDuration={2400}
        onClose={(_, reason) => {
          if (reason === "clickaway") {
            return;
          }
          setCopiedMessage("");
        }}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          severity={copiedMessage.startsWith("Copied") ? "success" : "error"}
          variant="filled"
          onClose={() => setCopiedMessage("")}
          sx={{ width: "100%" }}
        >
          {copiedMessage}
        </Alert>
      </Snackbar>
    </>
  );
}
