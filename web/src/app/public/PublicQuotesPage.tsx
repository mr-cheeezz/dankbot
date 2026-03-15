import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Box,
  Card,
  CardContent,
  Chip,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import { defaultPublicQuotes, fetchPublicQuotes, type PublicQuote } from "./api";

export function PublicQuotesPage() {
  const [items, setItems] = useState<PublicQuote[]>(defaultPublicQuotes);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");

  useEffect(() => {
    const controller = new AbortController();

    fetchPublicQuotes(controller.signal)
      .then((nextItems) => {
        setItems(nextItems);
      })
      .catch(() => {
        setItems(defaultPublicQuotes);
      })
      .finally(() => {
        setLoading(false);
      });

    return () => controller.abort();
  }, []);

  const filteredItems = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (normalized === "") {
      return items;
    }

    return items.filter((entry) => {
      return `#${entry.id} ${entry.message}`.toLowerCase().includes(normalized);
    });
  }, [items, query]);

  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h4">Quotes</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 0.75 }}>
          Saved stream quotes in one place, so viewers can browse them without turning chat into an archive.
        </Typography>
      </Box>

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
                {loading
                  ? "loading saved quotes..."
                  : filteredItems.length === items.length
                    ? `${items.length} saved quotes`
                    : `${filteredItems.length} matching quotes`}
              </Typography>
            </Box>
            <TextField
              size="small"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search quotes"
              sx={{ width: { xs: "100%", md: 320 } }}
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
              <CardContent sx={{ p: 2.25 }}>
                <Stack direction="row" spacing={1.25} alignItems="flex-start">
                  <Chip
                    label={`#${quote.id}`}
                    color="primary"
                    variant="outlined"
                    sx={{ mt: 0.15, flexShrink: 0 }}
                  />
                  <Typography sx={{ fontSize: "1rem", lineHeight: 1.7 }}>{quote.message}</Typography>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      ) : (
        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Typography variant="h6">
              {loading ? "Loading quotes..." : "No quotes yet"}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
              {loading
                ? "The archive is still loading."
                : query.trim() === ""
                  ? "There are no saved quotes yet."
                  : "No quotes matched your search."}
            </Typography>
          </CardContent>
        </Card>
      )}
    </Stack>
  );
}
