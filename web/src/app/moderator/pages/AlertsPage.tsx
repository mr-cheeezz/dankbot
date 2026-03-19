import {
  Box,
  Button,
  Checkbox,
  Chip,
  IconButton,
  InputAdornment,
  Paper,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import PollRoundedIcon from "@mui/icons-material/PollRounded";
import StarsRoundedIcon from "@mui/icons-material/StarsRounded";
import { useEffect, useMemo, useState } from "react";

import { useModerator } from "../ModeratorContext";
import type { AlertEntry } from "../types";

type AlertProvider = AlertEntry["provider"];
type AlertSectionTab = "basic" | "hypeSpam";
type HypeSpamRule = {
  id: string;
  minimumAmount: number;
  lineCount: number;
  emoteLine: string;
};
type HypeSpamTierLine = {
  id: string;
  label: string;
  messageCount: number;
  emoteLine: string;
};
type HypeSpamConfig = {
  enabled: boolean;
  enabledWhenOffline: boolean;
  enabledWhenOnline: boolean;
  rateLimitSeconds: number;
  minimumAmount?: number;
  rules?: HypeSpamRule[];
  singleLine?: string;
  tierLines?: HypeSpamTierLine[];
};
type HypeSpamKey = "bits" | "giftedSubs" | "subscriptions" | "resubscriptions";

type ProviderOption = {
  key: AlertProvider;
  label: string;
};

const providerLabels: Record<AlertProvider, string> = {
  twitch: "Twitch",
  streamlabs: "Streamlabs",
  streamelements: "StreamElements",
};

const maxHypeSpamLines = 6;
const minHypeSpamCooldownSeconds = 20;
const defaultSectionTabs: Record<string, AlertSectionTab> = {};
const initialHypeSpamConfigs: Record<HypeSpamKey, HypeSpamConfig> = {
  bits: {
    enabled: false,
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    minimumAmount: 100,
    rateLimitSeconds: 30,
    rules: [
      {
        id: "bits-1",
        minimumAmount: 100,
        lineCount: 1,
        emoteLine: "Cheer Cheer Cheer",
      },
      {
        id: "bits-2",
        minimumAmount: 500,
        lineCount: 2,
        emoteLine: "PogChamp Cheer PogChamp",
      },
      {
        id: "bits-3",
        minimumAmount: 1000,
        lineCount: 3,
        emoteLine: "Cheer PogChamp HYPE",
      },
    ],
  },
  giftedSubs: {
    enabled: false,
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    minimumAmount: 1,
    rateLimitSeconds: 45,
    singleLine: "bleedPurple POGGIES bleedPurple",
  },
  subscriptions: {
    enabled: false,
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    rateLimitSeconds: 30,
    tierLines: [
      {
        id: "subs-tier-1",
        label: "Tier 1",
        messageCount: 1,
        emoteLine: "POGGIES PogU POGGIES",
      },
      {
        id: "subs-tier-2",
        label: "Tier 2",
        messageCount: 3,
        emoteLine: "SubHype POGGIES SubHype",
      },
      {
        id: "subs-tier-3",
        label: "Tier 3",
        messageCount: 5,
        emoteLine: "PogChamp PogU POGGIES",
      },
      {
        id: "subs-prime",
        label: "Prime Gaming",
        messageCount: 1,
        emoteLine: "PrimeHype POGGIES PrimeHype",
      },
    ],
  },
  resubscriptions: {
    enabled: false,
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    rateLimitSeconds: 30,
    tierLines: [
      {
        id: "resubs-tier-1",
        label: "Tier 1",
        messageCount: 1,
        emoteLine: "POGGIES PogU POGGIES",
      },
      {
        id: "resubs-tier-2",
        label: "Tier 2",
        messageCount: 3,
        emoteLine: "SubHype POGGIES SubHype",
      },
      {
        id: "resubs-tier-3",
        label: "Tier 3",
        messageCount: 5,
        emoteLine: "PogChamp PogU POGGIES",
      },
      {
        id: "resubs-prime",
        label: "Prime Gaming",
        messageCount: 1,
        emoteLine: "PrimeHype POGGIES PrimeHype",
      },
    ],
  },
};

function providerIsAvailable(
  provider: AlertProvider,
  alerts: AlertEntry[],
): boolean {
  return alerts.some((entry) => entry.provider === provider);
}

function sectionSupportsHypeSpam(sectionTitle: string): HypeSpamKey | null {
  switch (sectionTitle) {
    case "Bits Alerts":
      return "bits";
    case "Gifted Subscription Alerts":
      return "giftedSubs";
    case "Subscription Alerts":
      return "subscriptions";
    case "Resubscription Alerts":
      return "resubscriptions";
    default:
      return null;
  }
}

function sanitizeEmoteLine(value: string): string {
  return value
    .split(/\s+/)
    .map((token) => token.trim())
    .filter((token) => /^[A-Za-z0-9_:-]+$/.test(token))
    .join(" ");
}

export function AlertsPage() {
  const {
    summary,
    alerts,
    filteredAlerts,
    toggleAlert,
    updateAlertTemplate,
    updateAlert,
  } = useModerator();
  const [provider, setProvider] = useState<AlertProvider>("twitch");
  const [sectionTabs, setSectionTabs] =
    useState<Record<string, AlertSectionTab>>(defaultSectionTabs);
  const [hypeSpamConfigs, setHypeSpamConfigs] = useState(
    initialHypeSpamConfigs,
  );
  const [pollSettings, setPollSettings] = useState({
    enabled: true,
    showPointBreakdown: true,
    mentionExtraVoting: true,
    minimumCalloutPoints: 1000,
    completionTemplate: "Channel points spent: {option_breakdown}",
  });
  const [predictionSettings, setPredictionSettings] = useState({
    enabled: true,
    showLockSummary: true,
    showOutcomeSummary: true,
    largeSpendThreshold: 50000,
    mentionTopPredictors: true,
  });

  const availableProviders = useMemo<ProviderOption[]>(() => {
    const options: ProviderOption[] = [
      { key: "twitch", label: providerLabels.twitch },
    ];
    const integrationMap = new Map(
      summary.integrations.map((entry) => [
        entry.id.trim().toLowerCase(),
        entry.status.trim().toLowerCase(),
      ]),
    );

    if (
      providerIsAvailable("streamlabs", alerts) &&
      integrationMap.has("streamlabs") &&
      (integrationMap.get("streamlabs") === "linked" ||
        integrationMap.get("streamlabs") === "configured")
    ) {
      options.push({ key: "streamlabs", label: providerLabels.streamlabs });
    }

    if (
      providerIsAvailable("streamelements", alerts) &&
      integrationMap.has("streamelements") &&
      (integrationMap.get("streamelements") === "linked" ||
        integrationMap.get("streamelements") === "configured")
    ) {
      options.push({
        key: "streamelements",
        label: providerLabels.streamelements,
      });
    }

    return options;
  }, [alerts, summary.integrations]);

  useEffect(() => {
    if (!availableProviders.some((entry) => entry.key === provider)) {
      setProvider(availableProviders[0]?.key ?? "twitch");
    }
  }, [availableProviders, provider]);

  useEffect(() => {
    if (!hypeSpamConfigs.giftedSubs.enabled) {
      return;
    }

    alerts
      .filter(
        (entry) =>
          entry.section === "Mass Gift Subscription Alerts" && entry.enabled,
      )
      .forEach((entry) => {
        updateAlert(entry.id, {
          enabled: false,
          status: "muted",
        });
      });
  }, [alerts, hypeSpamConfigs.giftedSubs.enabled, updateAlert]);

  const sections = useMemo(() => {
    const sectionMap = new Map<string, AlertEntry[]>();
    const order: string[] = [];

    filteredAlerts
      .filter((entry) => entry.provider === provider)
      .forEach((entry) => {
        if (!sectionMap.has(entry.section)) {
          sectionMap.set(entry.section, []);
          order.push(entry.section);
        }

        sectionMap.get(entry.section)?.push(entry);
      });

    return order.map((sectionTitle) => {
      const entries = sectionMap.get(sectionTitle) ?? [];
      return {
        title: sectionTitle,
        note: entries.find((entry) => entry.note)?.note ?? "",
        entries,
      };
    });
  }, [filteredAlerts, provider]);

  const updateHypeSpamConfig = (
    key: HypeSpamKey,
    next: Partial<HypeSpamConfig>,
  ) => {
    setHypeSpamConfigs((current) => ({
      ...current,
      [key]: {
        ...current[key],
        ...next,
      },
    }));
  };

  const updateHypeSpamRule = (
    key: HypeSpamKey,
    ruleId: string,
    next: Partial<HypeSpamRule>,
  ) => {
    setHypeSpamConfigs((current) => ({
      ...current,
      [key]: {
        ...current[key],
        rules: (current[key].rules ?? []).map((entry) =>
          entry.id === ruleId
            ? {
                ...entry,
                ...next,
              }
            : entry,
        ),
      },
    }));
  };

  const addHypeSpamRule = (key: HypeSpamKey) => {
    setHypeSpamConfigs((current) => ({
      ...current,
      [key]: {
        ...current[key],
        rules: [
          ...(current[key].rules ?? []),
          {
            id: `${key}-${Date.now().toString(36)}`,
            minimumAmount: current[key].minimumAmount ?? 1,
            lineCount: 1,
            emoteLine: "PogU PogU PogU",
          },
        ],
      },
    }));
  };

  const deleteHypeSpamRule = (key: HypeSpamKey, ruleId: string) => {
    setHypeSpamConfigs((current) => ({
      ...current,
      [key]: {
        ...current[key],
        rules: (current[key].rules ?? []).filter(
          (entry) => entry.id !== ruleId,
        ),
      },
    }));
  };

  return (
    <Paper
      elevation={0}
      sx={{
        overflow: "hidden",
        backgroundColor: "background.paper",
      }}
    >
      <Box
        sx={{
          px: 3,
          py: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <Typography variant="h5">Chat Alerts</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          Provider-based alert templates with cleaner grouping for Twitch,
          Streamlabs, and StreamElements.
        </Typography>
      </Box>

      <Tabs
        value={provider}
        onChange={(_, next: AlertProvider) => setProvider(next)}
        textColor="primary"
        indicatorColor="primary"
        sx={{
          px: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
          minHeight: 52,
          "& .MuiTabs-indicator": {
            height: 2,
          },
        }}
      >
        {availableProviders.map((entry) => (
          <Tab
            key={entry.key}
            value={entry.key}
            label={entry.label}
            disableRipple
          />
        ))}
      </Tabs>

      <Box sx={{ px: 3, py: 3 }}>
        {sections.length === 0 ? (
          <Paper
            elevation={0}
            sx={{
              p: 4,
              textAlign: "center",
              border: "1px dashed",
              borderColor: "divider",
              bgcolor: "background.default",
            }}
          >
            <Typography variant="h6">
              No alert cards match the current view.
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
              Try clearing the dashboard search or link another alerts provider
              first.
            </Typography>
          </Paper>
        ) : (
          <Box
            sx={{
              display: "grid",
              gap: 2.5,
              gridTemplateColumns: {
                xs: "1fr",
                xl: "repeat(2, minmax(0, 1fr))",
              },
            }}
          >
            {sections.map((section) => (
              <Paper
                key={section.title}
                elevation={0}
                sx={{
                  p: 2.5,
                  bgcolor: "background.default",
                  border: "1px solid",
                  borderColor: "divider",
                }}
              >
                <Stack spacing={2.5}>
                  <Stack spacing={0.75}>
                    <Stack
                      direction="row"
                      spacing={1}
                      alignItems="center"
                      justifyContent="space-between"
                    >
                      <Typography variant="h6" sx={{ fontSize: "1.22rem" }}>
                        {section.title}
                      </Typography>
                      <Chip
                        size="small"
                        label={providerLabels[provider]}
                        color={provider === "twitch" ? "primary" : "default"}
                        variant={provider === "twitch" ? "filled" : "outlined"}
                      />
                    </Stack>
                    {section.note !== "" ? (
                      <Box
                        sx={{
                          px: 2,
                          py: 1.5,
                          borderRadius: 1.75,
                          bgcolor: "background.paper",
                          border: "1px solid",
                          borderColor: "divider",
                        }}
                      >
                        <Typography variant="body2" color="text.secondary">
                          {section.note}
                        </Typography>
                      </Box>
                    ) : null}
                    {sectionSupportsHypeSpam(section.title) ? (
                      <Tabs
                        value={sectionTabs[section.title] ?? "basic"}
                        onChange={(_, next: AlertSectionTab) =>
                          setSectionTabs((current) => ({
                            ...current,
                            [section.title]: next,
                          }))
                        }
                        textColor="primary"
                        indicatorColor="primary"
                        sx={{
                          minHeight: 44,
                          mt: 0.25,
                          "& .MuiTabs-indicator": {
                            height: 2,
                          },
                        }}
                      >
                        <Tab value="basic" label="Basic Alert" disableRipple />
                        <Tab value="hypeSpam" label="Hype Spam" disableRipple />
                      </Tabs>
                    ) : null}
                  </Stack>

                  {(() => {
                    const hypeSpamKey = sectionSupportsHypeSpam(section.title);
                    const selectedTab = sectionTabs[section.title] ?? "basic";

                    if (hypeSpamKey != null && selectedTab === "hypeSpam") {
                      const config = hypeSpamConfigs[hypeSpamKey];
                      const basicAlertEnabled = section.entries.some(
                        (entry) => entry.enabled,
                      );
                      const amountLabel =
                        hypeSpamKey === "bits"
                          ? "Minimum bits amount"
                          : "Minimum gifted subs";
                      const amountUnit =
                        hypeSpamKey === "bits" ? "bits" : "gifts";
                      const usesRuleTable = hypeSpamKey === "bits";
                      const usesSingleLine = hypeSpamKey === "giftedSubs";
                      const usesTierLines =
                        hypeSpamKey === "subscriptions" ||
                        hypeSpamKey === "resubscriptions";
                      const effectiveHypeSpamEnabled =
                        config.enabled && !basicAlertEnabled;

                      return (
                        <Box
                          sx={{
                            display: "grid",
                            gap: 2,
                            gridTemplateColumns: {
                              xs: "1fr",
                              xl: "minmax(0, 0.88fr) minmax(0, 1.32fr)",
                            },
                          }}
                        >
                          <Paper
                            elevation={0}
                            sx={{
                              p: 2,
                              bgcolor: "background.paper",
                              border: "1px solid",
                              borderColor: "divider",
                            }}
                          >
                            <Stack spacing={2}>
                              {basicAlertEnabled ? (
                                <Box
                                  sx={{
                                    px: 2,
                                    py: 1.5,
                                    borderRadius: 1.75,
                                    bgcolor: "warning.main",
                                    color: "warning.contrastText",
                                  }}
                                >
                                  <Typography
                                    sx={{
                                      fontWeight: 700,
                                      fontSize: "0.95rem",
                                    }}
                                  >
                                    Basic alert is still enabled.
                                  </Typography>
                                  <Typography
                                    variant="body2"
                                    sx={{ color: "inherit", mt: 0.5 }}
                                  >
                                    Hype spam will stay inactive until you
                                    disable the matching basic alert rows in
                                    this section.
                                  </Typography>
                                </Box>
                              ) : null}

                              <Stack
                                direction="row"
                                alignItems="center"
                                justifyContent="space-between"
                              >
                                <Box>
                                  <Typography
                                    sx={{
                                      fontSize: "1.05rem",
                                      fontWeight: 600,
                                    }}
                                  >
                                    Enabled
                                  </Typography>
                                  <Typography
                                    variant="body2"
                                    color="text.secondary"
                                  >
                                    Hype spam is rate limited and only accepts
                                    emote-style lines.
                                  </Typography>
                                </Box>
                                <Switch
                                  checked={effectiveHypeSpamEnabled}
                                  onChange={() =>
                                    updateHypeSpamConfig(hypeSpamKey, {
                                      enabled: !config.enabled,
                                    })
                                  }
                                  color="primary"
                                  disabled={basicAlertEnabled}
                                />
                              </Stack>

                              <Stack
                                direction="row"
                                alignItems="center"
                                justifyContent="space-between"
                              >
                                <Box>
                                  <Typography
                                    sx={{ fontSize: "1rem", fontWeight: 500 }}
                                  >
                                    Enabled while stream offline
                                  </Typography>
                                  <Typography
                                    variant="body2"
                                    color="text.secondary"
                                  >
                                    Keep this off unless you really want offline
                                    hype moments.
                                  </Typography>
                                </Box>
                                <Switch
                                  checked={config.enabledWhenOffline}
                                  onChange={() =>
                                    updateHypeSpamConfig(hypeSpamKey, {
                                      enabledWhenOffline:
                                        !config.enabledWhenOffline,
                                    })
                                  }
                                  color="primary"
                                />
                              </Stack>

                              <Stack
                                direction="row"
                                alignItems="center"
                                justifyContent="space-between"
                              >
                                <Box>
                                  <Typography
                                    sx={{ fontSize: "1rem", fontWeight: 500 }}
                                  >
                                    Enabled while stream online
                                  </Typography>
                                  <Typography
                                    variant="body2"
                                    color="text.secondary"
                                  >
                                    Recommended for real hype moments, not
                                    constant spam.
                                  </Typography>
                                </Box>
                                <Switch
                                  checked={config.enabledWhenOnline}
                                  onChange={() =>
                                    updateHypeSpamConfig(hypeSpamKey, {
                                      enabledWhenOnline:
                                        !config.enabledWhenOnline,
                                    })
                                  }
                                  color="primary"
                                />
                              </Stack>

                              {config.minimumAmount != null ? (
                                <TextField
                                  fullWidth
                                  type="number"
                                  label={amountLabel}
                                  value={config.minimumAmount}
                                  onChange={(event) =>
                                    updateHypeSpamConfig(hypeSpamKey, {
                                      minimumAmount: Math.max(
                                        1,
                                        Number(event.target.value || "1"),
                                      ),
                                    })
                                  }
                                  helperText="The smallest amount that should trigger hype spam at all."
                                  InputProps={{
                                    endAdornment: (
                                      <InputAdornment position="end">
                                        {amountUnit}
                                      </InputAdornment>
                                    ),
                                  }}
                                  disabled={basicAlertEnabled}
                                />
                              ) : null}

                              <TextField
                                fullWidth
                                type="number"
                                label="Alert cooldown"
                                value={config.rateLimitSeconds}
                                onChange={(event) =>
                                  updateHypeSpamConfig(hypeSpamKey, {
                                    rateLimitSeconds: Math.max(
                                      minHypeSpamCooldownSeconds,
                                      Number(
                                        event.target.value ||
                                          String(minHypeSpamCooldownSeconds),
                                      ),
                                    ),
                                  })
                                }
                                helperText={`Minimum ${minHypeSpamCooldownSeconds}s between hype spam bursts. No abuse.`}
                                InputProps={{
                                  endAdornment: (
                                    <InputAdornment position="end">
                                      seconds
                                    </InputAdornment>
                                  ),
                                }}
                                disabled={basicAlertEnabled}
                              />

                              {hypeSpamKey === "giftedSubs" ? (
                                <Chip
                                  color={
                                    effectiveHypeSpamEnabled
                                      ? "warning"
                                      : "default"
                                  }
                                  variant={
                                    effectiveHypeSpamEnabled
                                      ? "filled"
                                      : "outlined"
                                  }
                                  label={
                                    effectiveHypeSpamEnabled
                                      ? "Mass gift alerts are automatically muted while gifted hype spam is on."
                                      : "Mass gift alerts stay available while gifted hype spam is off."
                                  }
                                  sx={{
                                    alignSelf: "flex-start",
                                    maxWidth: "100%",
                                  }}
                                />
                              ) : null}
                            </Stack>
                          </Paper>

                          <Paper
                            elevation={0}
                            sx={{
                              p: 2,
                              bgcolor: "background.paper",
                              border: "1px solid",
                              borderColor: "divider",
                            }}
                          >
                            <Stack spacing={1.75}>
                              <Stack
                                direction="row"
                                alignItems="center"
                                justifyContent="space-between"
                              >
                                <Box>
                                  <Typography
                                    sx={{
                                      fontSize: "1.05rem",
                                      fontWeight: 600,
                                    }}
                                  >
                                    {usesTierLines
                                      ? "Tier Hype Lines"
                                      : "Messages"}
                                  </Typography>
                                  <Typography
                                    variant="body2"
                                    color="text.secondary"
                                  >
                                    {usesSingleLine
                                      ? "One emote-only line per gifted sub. Keep it short and hype."
                                      : usesTierLines
                                        ? "Each tier uses a fixed message count. You can only edit the emote line."
                                        : "Emote-only lines. Any plain text gets stripped automatically."}
                                  </Typography>
                                </Box>
                                {usesRuleTable ? (
                                  <Button
                                    size="small"
                                    startIcon={<AddRoundedIcon />}
                                    onClick={() => addHypeSpamRule(hypeSpamKey)}
                                    disabled={basicAlertEnabled}
                                  >
                                    Add
                                  </Button>
                                ) : null}
                              </Stack>

                              {usesRuleTable
                                ? config.rules?.map((rule) => (
                                    <Stack
                                      key={rule.id}
                                      direction={{ xs: "column", md: "row" }}
                                      spacing={1}
                                      alignItems={{
                                        xs: "stretch",
                                        md: "center",
                                      }}
                                    >
                                      <TextField
                                        sx={{ width: { xs: "100%", md: 150 } }}
                                        type="number"
                                        label={`Min ${amountUnit}`}
                                        value={rule.minimumAmount}
                                        onChange={(event) =>
                                          updateHypeSpamRule(
                                            hypeSpamKey,
                                            rule.id,
                                            {
                                              minimumAmount: Math.max(
                                                1,
                                                Number(
                                                  event.target.value || "1",
                                                ),
                                              ),
                                            },
                                          )
                                        }
                                        disabled={basicAlertEnabled}
                                      />
                                      <TextField
                                        sx={{ width: { xs: "100%", md: 140 } }}
                                        type="number"
                                        label="Lines"
                                        value={rule.lineCount}
                                        onChange={(event) =>
                                          updateHypeSpamRule(
                                            hypeSpamKey,
                                            rule.id,
                                            {
                                              lineCount: Math.max(
                                                1,
                                                Math.min(
                                                  maxHypeSpamLines,
                                                  Number(
                                                    event.target.value || "1",
                                                  ),
                                                ),
                                              ),
                                            },
                                          )
                                        }
                                        helperText={`Max ${maxHypeSpamLines}`}
                                        disabled={basicAlertEnabled}
                                      />
                                      <TextField
                                        fullWidth
                                        label="Message"
                                        value={rule.emoteLine}
                                        onChange={(event) =>
                                          updateHypeSpamRule(
                                            hypeSpamKey,
                                            rule.id,
                                            {
                                              emoteLine: sanitizeEmoteLine(
                                                event.target.value,
                                              ),
                                            },
                                          )
                                        }
                                        helperText="Use emote tokens only, like POGGIES PogU SubHype."
                                        disabled={basicAlertEnabled}
                                      />
                                      <IconButton
                                        color="error"
                                        onClick={() =>
                                          deleteHypeSpamRule(
                                            hypeSpamKey,
                                            rule.id,
                                          )
                                        }
                                        disabled={
                                          basicAlertEnabled ||
                                          (config.rules?.length ?? 0) <= 1
                                        }
                                      >
                                        <DeleteOutlineRoundedIcon />
                                      </IconButton>
                                    </Stack>
                                  ))
                                : null}

                              {usesSingleLine ? (
                                <TextField
                                  fullWidth
                                  label="Emote Line"
                                  value={config.singleLine ?? ""}
                                  onChange={(event) =>
                                    updateHypeSpamConfig(hypeSpamKey, {
                                      singleLine: sanitizeEmoteLine(
                                        event.target.value,
                                      ),
                                    })
                                  }
                                  helperText="This one emote line is repeated once per gifted sub."
                                  disabled={basicAlertEnabled}
                                />
                              ) : null}

                              {usesTierLines
                                ? config.tierLines?.map((entry) => (
                                    <Stack
                                      key={entry.id}
                                      direction={{ xs: "column", md: "row" }}
                                      spacing={1}
                                      alignItems={{
                                        xs: "stretch",
                                        md: "center",
                                      }}
                                    >
                                      <TextField
                                        sx={{ width: { xs: "100%", md: 170 } }}
                                        label="Tier"
                                        value={entry.label}
                                        disabled
                                      />
                                      <TextField
                                        sx={{ width: { xs: "100%", md: 160 } }}
                                        label="Messages"
                                        value={entry.messageCount}
                                        disabled
                                        helperText="Fixed for this hype mode"
                                      />
                                      <TextField
                                        fullWidth
                                        label="Emote Line"
                                        value={entry.emoteLine}
                                        onChange={(event) =>
                                          setHypeSpamConfigs((current) => ({
                                            ...current,
                                            [hypeSpamKey]: {
                                              ...current[hypeSpamKey],
                                              tierLines: current[
                                                hypeSpamKey
                                              ].tierLines?.map((line) =>
                                                line.id === entry.id
                                                  ? {
                                                      ...line,
                                                      emoteLine:
                                                        sanitizeEmoteLine(
                                                          event.target.value,
                                                        ),
                                                    }
                                                  : line,
                                              ),
                                            },
                                          }))
                                        }
                                        helperText="Emote-only line for this tier."
                                        disabled={basicAlertEnabled}
                                      />
                                    </Stack>
                                  ))
                                : null}
                            </Stack>
                          </Paper>
                        </Box>
                      );
                    }

                    const massGiftSuppressed =
                      section.title === "Mass Gift Subscription Alerts" &&
                      hypeSpamConfigs.giftedSubs.enabled;

                    return section.entries.map((entry) => (
                      <Stack
                        key={entry.id}
                        spacing={1.25}
                        sx={{
                          pt: 0.5,
                          borderTop: "1px solid",
                          borderColor: "divider",
                          opacity: massGiftSuppressed ? 0.55 : 1,
                          "&:first-of-type": {
                            pt: 0,
                            borderTop: "none",
                          },
                        }}
                      >
                        <Stack
                          direction="row"
                          spacing={2}
                          alignItems="center"
                          justifyContent="space-between"
                        >
                          <Box>
                            <Typography
                              sx={{ fontSize: "1.05rem", fontWeight: 600 }}
                            >
                              {entry.label}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              {massGiftSuppressed
                                ? "Ignored while gifted-sub hype spam is enabled."
                                : entry.behavior}
                            </Typography>
                          </Box>
                          <Switch
                            checked={entry.enabled}
                            onChange={() => {
                              if (!massGiftSuppressed) {
                                toggleAlert(entry.id);
                              }
                            }}
                            color="primary"
                            disabled={massGiftSuppressed}
                          />
                        </Stack>

                        <TextField
                          fullWidth
                          multiline
                          minRows={2}
                          label="Alert Message"
                          value={entry.template}
                          onChange={(event) =>
                            updateAlertTemplate(entry.id, event.target.value)
                          }
                          disabled={massGiftSuppressed}
                        />

                        {entry.minimumLabel ? (
                          <TextField
                            fullWidth
                            type="number"
                            label={entry.minimumLabel}
                            value={entry.minimumValue ?? 0}
                            onChange={(event) =>
                              updateAlert(entry.id, {
                                minimumValue: Number(event.target.value || "0"),
                              })
                            }
                            disabled={massGiftSuppressed}
                            InputProps={{
                              startAdornment:
                                entry.minimumPrefix != null ? (
                                  <InputAdornment position="start">
                                    {entry.minimumPrefix}
                                  </InputAdornment>
                                ) : undefined,
                              endAdornment:
                                entry.minimumUnit != null ? (
                                  <InputAdornment position="end">
                                    {entry.minimumUnit}
                                  </InputAdornment>
                                ) : undefined,
                            }}
                          />
                        ) : null}
                      </Stack>
                    ));
                  })()}

                  {provider === "twitch" && section.title === "Poll Alerts" ? (
                    <Box
                      sx={{
                        pt: 2,
                        borderTop: "1px solid",
                        borderColor: "divider",
                      }}
                    >
                      <Stack spacing={2.25}>
                        <Stack direction="row" spacing={1.2} alignItems="center">
                          <PollRoundedIcon sx={{ color: "primary.main" }} />
                          <Box>
                            <Typography variant="h6" sx={{ fontSize: "1.22rem" }}>
                              Poll point behavior
                            </Typography>
                            <Typography
                              variant="body2"
                              color="text.secondary"
                              sx={{ mt: 0.35 }}
                            >
                              Tune how extra-vote channel point totals get
                              surfaced in poll alerts.
                            </Typography>
                          </Box>
                        </Stack>

                        <Stack spacing={1.25}>
                          <CheckboxRow
                            label="Enable poll point add-ons"
                            checked={pollSettings.enabled}
                            onChange={(checked) =>
                              setPollSettings((current) => ({
                                ...current,
                                enabled: checked,
                              }))
                            }
                          />
                          <CheckboxRow
                            label="Show per-option point breakdown when a poll ends"
                            checked={pollSettings.showPointBreakdown}
                            onChange={(checked) =>
                              setPollSettings((current) => ({
                                ...current,
                                showPointBreakdown: checked,
                              }))
                            }
                          />
                          <CheckboxRow
                            label="Mention when extra voting with channel points was enabled"
                            checked={pollSettings.mentionExtraVoting}
                            onChange={(checked) =>
                              setPollSettings((current) => ({
                                ...current,
                                mentionExtraVoting: checked,
                              }))
                            }
                          />
                        </Stack>

                        <Box
                          sx={{
                            display: "grid",
                            gridTemplateColumns: {
                              xs: "1fr",
                              md: "280px minmax(0, 1fr)",
                            },
                            gap: 2,
                          }}
                        >
                          <TextField
                            fullWidth
                            type="number"
                            label="Minimum callout points"
                            value={pollSettings.minimumCalloutPoints}
                            onChange={(event) =>
                              setPollSettings((current) => ({
                                ...current,
                                minimumCalloutPoints:
                                  Number(event.target.value) || 0,
                              }))
                            }
                            inputProps={{ min: 0, step: 100 }}
                          />
                          <TextField
                            fullWidth
                            label="Completion template"
                            value={pollSettings.completionTemplate}
                            onChange={(event) =>
                              setPollSettings((current) => ({
                                ...current,
                                completionTemplate: event.target.value,
                              }))
                            }
                            multiline
                            minRows={3}
                          />
                        </Box>
                      </Stack>
                    </Box>
                  ) : null}

                  {provider === "twitch" &&
                  section.title === "Prediction Alerts" ? (
                    <Box
                      sx={{
                        pt: 2,
                        borderTop: "1px solid",
                        borderColor: "divider",
                      }}
                    >
                      <Stack spacing={2.25}>
                        <Stack direction="row" spacing={1.2} alignItems="center">
                          <StarsRoundedIcon sx={{ color: "primary.main" }} />
                          <Box>
                            <Typography variant="h6" sx={{ fontSize: "1.22rem" }}>
                              Prediction point behavior
                            </Typography>
                            <Typography
                              variant="body2"
                              color="text.secondary"
                              sx={{ mt: 0.35 }}
                            >
                              Decide when big prediction spends get called out
                              and how point winners are summarized back into
                              chat.
                            </Typography>
                          </Box>
                        </Stack>

                        <Stack spacing={1.25}>
                          <CheckboxRow
                            label="Enable prediction point add-ons"
                            checked={predictionSettings.enabled}
                            onChange={(checked) =>
                              setPredictionSettings((current) => ({
                                ...current,
                                enabled: checked,
                              }))
                            }
                          />
                          <CheckboxRow
                            label="Show locked-outcome summary"
                            checked={predictionSettings.showLockSummary}
                            onChange={(checked) =>
                              setPredictionSettings((current) => ({
                                ...current,
                                showLockSummary: checked,
                              }))
                            }
                          />
                          <CheckboxRow
                            label="Show outcome winner summary"
                            checked={predictionSettings.showOutcomeSummary}
                            onChange={(checked) =>
                              setPredictionSettings((current) => ({
                                ...current,
                                showOutcomeSummary: checked,
                              }))
                            }
                          />
                          <CheckboxRow
                            label="Mention top predictors on locks and results"
                            checked={predictionSettings.mentionTopPredictors}
                            onChange={(checked) =>
                              setPredictionSettings((current) => ({
                                ...current,
                                mentionTopPredictors: checked,
                              }))
                            }
                          />
                        </Stack>

                        <Box
                          sx={{
                            display: "grid",
                            gridTemplateColumns: {
                              xs: "1fr",
                              md: "220px minmax(0, 1fr)",
                            },
                            gap: 2,
                          }}
                        >
                          <TextField
                            fullWidth
                            type="number"
                            label="Large spend threshold"
                            value={predictionSettings.largeSpendThreshold}
                            onChange={(event) =>
                              setPredictionSettings((current) => ({
                                ...current,
                                largeSpendThreshold:
                                  Number(event.target.value) || 0,
                              }))
                            }
                            inputProps={{ min: 0, step: 1000 }}
                            helperText="Only surface prediction progress callouts above this spend."
                          />
                          <Box
                            sx={{
                              px: 1,
                              alignSelf: "center",
                            }}
                          >
                            <Typography variant="body2" color="text.secondary">
                              Lock and result wording is edited directly in the
                              <strong> Prediction Locked </strong>
                              and
                              <strong> Prediction Ended </strong>
                              alert rows above.
                            </Typography>
                          </Box>
                        </Box>
                      </Stack>
                    </Box>
                  ) : null}
                </Stack>
              </Paper>
            ))}
          </Box>
        )}
      </Box>
    </Paper>
  );
}

function CheckboxRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (next: boolean) => void;
}) {
  return (
    <Stack direction="row" spacing={1.1} alignItems="center">
      <Checkbox
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
      />
      <Typography>{label}</Typography>
    </Stack>
  );
}
