import MusicNoteRoundedIcon from "@mui/icons-material/MusicNoteRounded";
import OpenInNewRoundedIcon from "@mui/icons-material/OpenInNewRounded";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
import ShieldRoundedIcon from "@mui/icons-material/ShieldRounded";
import {
  Alert,
  Autocomplete,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControlLabel,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "../../auth/AuthContext";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import {
  assignDashboardEditor,
  deleteDashboardEditor,
  fetchDashboardRoles,
  fetchPublicHomeSettings,
  savePublicHomeSettings,
  searchDashboardTwitchUsers,
} from "../api";
import { useModerator } from "../ModeratorContext";
import type { DashboardRoleEntry, PublicHomeSettings, TwitchUserSearchEntry } from "../types";

const defaultSettings: PublicHomeSettings = {
  showNowPlaying: true,
  showNowPlayingAlbumArt: true,
  showNowPlayingProgress: true,
  showNowPlayingLinks: true,
  commandPrefix: "!",
  promoLinks: [],
  robloxLinkCommandTarget: "dankbot",
  robloxLinkCommandTemplate: "",
  robloxLinkCommandDeleteTemplate: "",
};

const linkCommandTemplateDefaults: Record<
  PublicHomeSettings["robloxLinkCommandTarget"],
  string
> = {
  dankbot: "",
  nightbot: "!commands edit !link {link}",
  fossabot: "!setcommand !link {link}",
  pajbot: "!setcommand !link {link}",
  custom: "",
};

const linkCommandDeleteTemplateDefaults: Record<
  PublicHomeSettings["robloxLinkCommandTarget"],
  string
> = {
  dankbot: "",
  nightbot: "!commands del !link",
  fossabot: "!delcommand !link",
  pajbot: "!delcommand !link",
  custom: "",
};

function normalizeTwitchLoginInput(raw: string): string {
  const value = raw.trim();
  if (value === "") {
    return "";
  }

  const parenMatch = value.match(/\(@?([a-z0-9_]{2,25})\)/i);
  if (parenMatch?.[1]) {
    return parenMatch[1].toLowerCase();
  }

  const directMatch = value.match(/^@?([a-z0-9_]{2,25})$/i);
  if (directMatch?.[1]) {
    return directMatch[1].toLowerCase();
  }

  const fallbackMatch = value.match(/@?([a-z0-9_]{2,25})/i);
  if (fallbackMatch?.[1]) {
    return fallbackMatch[1].toLowerCase();
  }

  return "";
}

export function SettingsPage() {
  const { summary } = useModerator();
  const { session, refresh } = useAuth();
  const [settings, setSettings] = useState<PublicHomeSettings>(defaultSettings);
  const [dashboardRoles, setDashboardRoles] = useState<DashboardRoleEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [rolesLoading, setRolesLoading] = useState(true);
  const [rolesSaving, setRolesSaving] = useState(false);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [userSearchError, setUserSearchError] = useState("");
  const [message, setMessage] = useState("");
  const [rolesMessage, setRolesMessage] = useState("");
  const [editorLogin, setEditorLogin] = useState("");
  const [editorSearchResults, setEditorSearchResults] = useState<TwitchUserSearchEntry[]>([]);
  const [selectedEditorCandidate, setSelectedEditorCandidate] = useState<TwitchUserSearchEntry | null>(null);
  const [editorToRemove, setEditorToRemove] = useState<DashboardRoleEntry | null>(null);

  const spotifyLinked = useMemo(
    () => summary.integrations.some((entry) => entry.id === "spotify" && entry.status === "linked"),
    [summary.integrations],
  );
  const commandPrefixDisplay = settings.commandPrefix.trim() === "" ? "!" : settings.commandPrefix.trim();
  const canManageEditors = session.user?.isBroadcaster === true || session.user?.isAdmin === true;
  const editorRoles = useMemo(
    () => dashboardRoles.filter((entry) => entry.roleName === "editor"),
    [dashboardRoles],
  );

  useEffect(() => {
    const controller = new AbortController();

    fetchPublicHomeSettings(controller.signal)
      .then((nextSettings) => {
        setSettings(nextSettings);
      })
      .catch(() => {
        setSettings(defaultSettings);
      })
      .finally(() => {
        setLoading(false);
      });

    fetchDashboardRoles(controller.signal)
      .then((nextRoles) => {
        setDashboardRoles(nextRoles);
      })
      .catch(() => {
        setDashboardRoles([]);
      })
      .finally(() => {
        setRolesLoading(false);
      });

    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (!canManageEditors) {
      return;
    }

    const query = editorLogin.trim();
    if (query.length < 2) {
      setEditorSearchResults([]);
      setUserSearchError("");
      return;
    }

    const controller = new AbortController();
    const timeoutID = window.setTimeout(() => {
      setUserSearchLoading(true);
      setUserSearchError("");
      searchDashboardTwitchUsers(query, controller.signal)
        .then((results) => {
          setEditorSearchResults(results);
        })
        .catch(() => {
          setEditorSearchResults([]);
          setUserSearchError("Could not load Twitch suggestions right now. Exact login still works.");
        })
        .finally(() => {
          setUserSearchLoading(false);
        });
    }, 250);

    return () => {
      controller.abort();
      window.clearTimeout(timeoutID);
    };
  }, [canManageEditors, editorLogin]);

  const editorSearchOptions = useMemo(() => {
    const normalizedLogin = editorLogin.trim().replace(/^@+/, "").toLowerCase();
    if (normalizedLogin.length < 2) {
      return editorSearchResults;
    }

    const hasExactMatch = editorSearchResults.some(
      (entry) => entry.login.trim().toLowerCase() === normalizedLogin,
    );
    if (hasExactMatch) {
      return editorSearchResults;
    }

    return [
      {
        userId: `exact-login:${normalizedLogin}`,
        login: normalizedLogin,
        displayName: "",
        avatarURL: "",
      },
      ...editorSearchResults,
    ];
  }, [editorLogin, editorSearchResults]);

  const updateField = <K extends keyof PublicHomeSettings>(key: K, value: PublicHomeSettings[K]) => {
    setSettings((current) => ({
      ...current,
      [key]: value,
    }));
    setMessage("");
  };

  const handleLinkTargetChange = (target: PublicHomeSettings["robloxLinkCommandTarget"]) => {
    setSettings((current) => {
      const nextTemplate =
        current.robloxLinkCommandTemplate.trim() === "" ||
        current.robloxLinkCommandTemplate === linkCommandTemplateDefaults[current.robloxLinkCommandTarget]
          ? linkCommandTemplateDefaults[target]
          : current.robloxLinkCommandTemplate;
      const nextDeleteTemplate =
        current.robloxLinkCommandDeleteTemplate.trim() === "" ||
        current.robloxLinkCommandDeleteTemplate ===
          linkCommandDeleteTemplateDefaults[current.robloxLinkCommandTarget]
          ? linkCommandDeleteTemplateDefaults[target]
          : current.robloxLinkCommandDeleteTemplate;

      return {
        ...current,
        robloxLinkCommandTarget: target,
        robloxLinkCommandTemplate: nextTemplate,
        robloxLinkCommandDeleteTemplate: nextDeleteTemplate,
      };
    });
    setMessage("");
  };

  const updatePromoLink = (index: number, key: "label" | "href", value: string) => {
    setSettings((current) => {
      const nextLinks = [...current.promoLinks];
      while (nextLinks.length <= index) {
        nextLinks.push({ label: "", href: "" });
      }
      nextLinks[index] = {
        ...nextLinks[index],
        [key]: value,
      };
      return {
        ...current,
        promoLinks: nextLinks,
      };
    });
    setMessage("");
  };

  const addPromoLink = () => {
    setSettings((current) => ({
      ...current,
      promoLinks: [...current.promoLinks, { label: "", href: "" }],
    }));
    setMessage("");
  };

  const removePromoLink = (index: number) => {
    setSettings((current) => ({
      ...current,
      promoLinks: current.promoLinks.filter((_, itemIndex) => itemIndex !== index),
    }));
    setMessage("");
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage("");

    try {
      const saved = await savePublicHomeSettings(settings);
      setSettings(saved);
      setMessage("Channel settings saved.");
    } catch {
      setMessage("Could not save channel settings right now.");
    } finally {
      setSaving(false);
    }
  };

  const handleAssignEditor = async () => {
    const selectedCandidate =
      selectedEditorCandidate != null && !selectedEditorCandidate.userId.startsWith("exact-login:")
        ? selectedEditorCandidate
        : null;
    const login = normalizeTwitchLoginInput(selectedCandidate?.login || editorLogin);
    if (login === "") {
      setRolesMessage("Enter a Twitch login before adding an editor.");
      return;
    }

    setRolesSaving(true);
    setRolesMessage("");

    try {
      const nextRoles = await assignDashboardEditor({
        login,
        userId: selectedCandidate?.userId,
        displayName: selectedCandidate?.displayName,
      });
      setDashboardRoles(nextRoles);
      setEditorLogin("");
      setSelectedEditorCandidate(null);
      setEditorSearchResults([]);
      setRolesMessage(`Editor access added for ${login}.`);
      await refresh();
    } catch (error) {
      const detail = error instanceof Error ? error.message.trim() : "";
      setRolesMessage(detail !== "" ? `Could not add editor: ${detail}` : "Could not add that editor right now.");
    } finally {
      setRolesSaving(false);
    }
  };

  const handleRemoveEditor = async () => {
    if (editorToRemove == null) {
      return;
    }

    setRolesSaving(true);
    setRolesMessage("");

    try {
      const nextRoles = await deleteDashboardEditor(editorToRemove.userId);
      setDashboardRoles(nextRoles);
      setRolesMessage(`Removed editor access for ${editorToRemove.displayName || editorToRemove.login}.`);
      setEditorToRemove(null);
      await refresh();
    } catch {
      setRolesMessage("Could not remove that editor right now.");
    } finally {
      setRolesSaving(false);
    }
  };

  return (
    <>
      <Stack spacing={2.5}>
      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack
            spacing={2}
          >
            <Box>
              <Typography variant="h5">Channel Settings</Typography>
              <Typography color="text.secondary" sx={{ mt: 0.55 }}>
                Control what public-facing widgets show up on the home page.
              </Typography>
            </Box>
          </Stack>
        </CardContent>
      </Card>

      {canManageEditors ? (
        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Stack spacing={2.5}>
              <Box>
                <Stack direction="row" spacing={1} alignItems="center">
                  <ShieldRoundedIcon color="primary" />
                  <Typography variant="h5">Dashboard Access</Typography>
                </Stack>
                <Typography color="text.secondary" sx={{ mt: 0.8, maxWidth: 860 }}>
                  Twitch moderators automatically get dashboard access. Editors are manual website
                  roles that can be handed out by the broadcaster or the configured admin.
                </Typography>
              </Box>

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", xl: "minmax(0, 340px) minmax(0, 1fr)" },
                  gap: 2.5,
                }}
              >
                <Card variant="outlined">
                  <CardContent sx={{ p: 2 }}>
                    <Stack spacing={1.5}>
                      <Typography variant="h6">Give Editor Access</Typography>
                      <Typography color="text.secondary" sx={{ fontSize: "0.95rem" }}>
                        Add a Twitch login here to grant dashboard access without making that person
                        a channel moderator.
                      </Typography>
                      <Autocomplete
                        options={editorSearchOptions}
                        loading={userSearchLoading}
                        filterOptions={(options) => options}
                        value={selectedEditorCandidate}
                        onChange={(_, value) => {
                          setSelectedEditorCandidate(value);
                          setEditorLogin(value?.login ?? "");
                          setRolesMessage("");
                        }}
                        inputValue={editorLogin}
                        onInputChange={(_, value, reason) => {
                          if (reason === "reset") {
                            if (selectedEditorCandidate != null) {
                              setEditorLogin(selectedEditorCandidate.login);
                            }
                            return;
                          }
                          if (reason === "clear") {
                            setSelectedEditorCandidate(null);
                            setEditorLogin("");
                            setRolesMessage("");
                            return;
                          }
                          setSelectedEditorCandidate(null);
                          setEditorLogin(value);
                          setRolesMessage("");
                        }}
                        getOptionLabel={(option) =>
                          typeof option === "string"
                            ? option
                            : option.displayName
                              ? `${option.displayName} (@${option.login})`
                              : option.login
                        }
                        isOptionEqualToValue={(option, value) => option.userId === value.userId}
                        noOptionsText={
                          editorLogin.trim().length < 2
                            ? "Type at least 2 characters"
                            : "No Twitch users found. Exact login works best here."
                        }
                        disabled={rolesSaving || rolesLoading}
                        renderOption={(props, option) => (
                          <Box component="li" {...props} sx={{ display: "flex", gap: 1.25, alignItems: "center" }}>
                            <Avatar
                              src={option.avatarURL || undefined}
                              sx={{ width: 32, height: 32, bgcolor: "primary.dark" }}
                            >
                              {(option.userId.startsWith("exact-login:")
                                ? option.login
                                : option.displayName || option.login)
                                .slice(0, 2)
                                .toUpperCase()}
                            </Avatar>
                            <Box sx={{ minWidth: 0 }}>
                              <Typography sx={{ fontWeight: 700 }} noWrap>
                                {option.userId.startsWith("exact-login:")
                                  ? `Use exact login @${option.login}`
                                  : option.displayName || option.login}
                              </Typography>
                              {option.userId.startsWith("exact-login:") ? (
                                <Typography color="text.secondary" sx={{ fontSize: "0.88rem" }} noWrap>
                                  Add this exact login even if Twitch shows no suggestions.
                                </Typography>
                              ) : (
                                <Typography color="text.secondary" sx={{ fontSize: "0.88rem" }} noWrap>
                                  @{option.login}
                                </Typography>
                              )}
                            </Box>
                          </Box>
                        )}
                        renderInput={(params) => (
                          <TextField
                            {...params}
                            label="Search Twitch users"
                            placeholder="Start typing a Twitch login or display name"
                            helperText={
                              userSearchError !== ""
                                ? userSearchError
                                : "Exact Twitch login lookup works for any user. Broader suggestions depend on Twitch channel search."
                            }
                            error={userSearchError !== ""}
                          />
                        )}
                      />
                      <Button
                        variant="contained"
                        onClick={() => {
                          void handleAssignEditor();
                        }}
                        disabled={rolesSaving || rolesLoading}
                      >
                        {rolesSaving ? "Saving..." : "Add Editor"}
                      </Button>
                      {rolesMessage !== "" ? (
                        <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                          {rolesMessage}
                        </Typography>
                      ) : null}
                    </Stack>
                  </CardContent>
                </Card>

                <Card variant="outlined">
                  <CardContent sx={{ p: 2 }}>
                    <Stack spacing={1.5}>
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        spacing={1}
                        alignItems="center"
                      >
                        <Typography variant="h6">Editors</Typography>
                        <Chip
                          label={`${editorRoles.length} assigned`}
                          color={editorRoles.length > 0 ? "primary" : "default"}
                          variant="outlined"
                        />
                      </Stack>

                      {rolesLoading ? (
                        <Typography color="text.secondary">Loading dashboard roles...</Typography>
                      ) : editorRoles.length === 0 ? (
                        <Alert severity="info">
                          No editors have been assigned yet. Moderator access still comes from Twitch
                          automatically.
                        </Alert>
                      ) : (
                        <Stack spacing={1.25}>
                          {editorRoles.map((entry) => {
                            const name = entry.displayName || entry.login;
                            const initials =
                              name
                                .trim()
                                .split(/\s+/)
                                .filter(Boolean)
                                .slice(0, 2)
                                .map((part) => part[0]?.toUpperCase() ?? "")
                                .join("") || "ED";

                            return (
                              <Box
                                key={`${entry.userId}-${entry.roleName}`}
                                sx={{
                                  display: "flex",
                                  alignItems: "center",
                                  justifyContent: "space-between",
                                  gap: 1.5,
                                  p: 1.5,
                                  border: "1px solid",
                                  borderColor: "divider",
                                  borderRadius: 2,
                                  bgcolor: "background.default",
                                }}
                              >
                                <Stack direction="row" spacing={1.25} alignItems="center">
                                  <Avatar sx={{ width: 40, height: 40, bgcolor: "primary.dark" }}>
                                    {initials}
                                  </Avatar>
                                  <Box>
                                    <Typography sx={{ fontWeight: 700 }}>{name}</Typography>
                                    <Typography color="text.secondary" sx={{ fontSize: "0.9rem" }}>
                                      @{entry.login}
                                    </Typography>
                                    {entry.assignedByLogin !== "" ? (
                                      <Typography
                                        color="text.secondary"
                                        sx={{ fontSize: "0.82rem", mt: 0.25 }}
                                      >
                                        Added by @{entry.assignedByLogin}
                                      </Typography>
                                    ) : null}
                                  </Box>
                                </Stack>

                                <Button
                                  variant="outlined"
                                  color="error"
                                  onClick={() => setEditorToRemove(entry)}
                                  disabled={rolesSaving}
                                >
                                  Remove
                                </Button>
                              </Box>
                            );
                          })}
                        </Stack>
                      )}
                    </Stack>
                  </CardContent>
                </Card>
              </Box>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack spacing={2.5}>
            <Box>
              <Stack direction="row" spacing={1} alignItems="center">
                <MusicNoteRoundedIcon color="primary" />
                <Typography variant="h5">Public Home Now Playing</Typography>
              </Stack>
              <Typography color="text.secondary" sx={{ mt: 0.8, maxWidth: 760 }}>
                This controls the Spotify card on the public home page. When Spotify is linked and
                the streamer is actively listening, the card can show the current track, album art,
                progress, and Spotify links.
              </Typography>
            </Box>

            {!spotifyLinked ? (
              <Alert severity="info">
                Link Spotify first if you want the public home page to show the streamer&apos;s
                current track.
              </Alert>
            ) : null}

            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) minmax(340px, 0.9fr)" },
                gap: 2.5,
              }}
            >
              <Stack spacing={1.2}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={settings.showNowPlaying}
                      onChange={(event) => updateField("showNowPlaying", event.target.checked)}
                    />
                  }
                  label="Show now playing card on the public home page"
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={settings.showNowPlayingAlbumArt}
                      onChange={(event) =>
                        updateField("showNowPlayingAlbumArt", event.target.checked)
                      }
                      disabled={!settings.showNowPlaying}
                    />
                  }
                  label="Show album art"
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={settings.showNowPlayingProgress}
                      onChange={(event) =>
                        updateField("showNowPlayingProgress", event.target.checked)
                      }
                      disabled={!settings.showNowPlaying}
                    />
                  }
                  label="Show playback progress"
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={settings.showNowPlayingLinks}
                      onChange={(event) =>
                        updateField("showNowPlayingLinks", event.target.checked)
                      }
                      disabled={!settings.showNowPlaying}
                    />
                  }
                  label="Show Spotify quick links"
                />

                <Stack direction="row" spacing={1.25} sx={{ pt: 1 }}>
                  <Button
                    variant="contained"
                    onClick={() => {
                      void handleSave();
                    }}
                    disabled={saving || loading}
                  >
                    {saving ? "Saving..." : "Save"}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setSettings(defaultSettings);
                      setMessage("");
                    }}
                    disabled={saving || loading}
                  >
                    Reset
                  </Button>
                </Stack>

                {message !== "" ? (
                  <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                    {message}
                  </Typography>
                ) : null}
              </Stack>

              <Card variant="outlined">
                <CardContent sx={{ p: 2 }}>
                  <Typography
                    sx={{
                      fontSize: "0.78rem",
                      fontWeight: 800,
                      textTransform: "uppercase",
                      letterSpacing: "0.08em",
                      color: "text.secondary",
                      mb: 1.4,
                    }}
                  >
                    Preview
                  </Typography>
                  <Stack spacing={1.5}>
                    <Stack direction="row" spacing={1.5} alignItems="center">
                      {settings.showNowPlayingAlbumArt ? (
                        <Box
                          sx={{
                            width: 92,
                            height: 92,
                            borderRadius: 1.25,
                            background:
                              "linear-gradient(135deg, rgba(74,137,255,0.2), rgba(112,214,163,0.16))",
                            border: "1px solid",
                            borderColor: "divider",
                          }}
                        />
                      ) : null}
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="h6">Contemplative Reptile</Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.45 }}>
                          Sounds of Nature
                        </Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.25, fontSize: "0.9rem" }}>
                          Listening now
                        </Typography>
                      </Box>
                    </Stack>

                    {settings.showNowPlayingProgress ? (
                      <Box
                        sx={{
                          height: 6,
                          borderRadius: 999,
                          backgroundColor: "rgba(255,255,255,0.08)",
                          overflow: "hidden",
                        }}
                      >
                        <Box
                          sx={{
                            width: "42%",
                            height: "100%",
                            bgcolor: "primary.main",
                          }}
                        />
                      </Box>
                    ) : null}

                    {settings.showNowPlayingLinks ? (
                      <Stack direction="row" spacing={1}>
                        <Button size="small" variant="outlined" startIcon={<OpenInNewRoundedIcon />}>
                          Track
                        </Button>
                        <Button size="small" variant="outlined">
                          Album
                        </Button>
                        <Button size="small" variant="outlined">
                          Artist
                        </Button>
                      </Stack>
                    ) : null}
                  </Stack>
                </CardContent>
              </Card>
            </Box>
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack spacing={2.5}>
            <Box>
              <Stack direction="row" spacing={1} alignItems="center">
                <LinkRoundedIcon color="primary" />
                <Typography variant="h5">Chat Command Prefix</Typography>
              </Stack>
              <Typography color="text.secondary" sx={{ mt: 0.8, maxWidth: 760 }}>
                Set the default command prefix for chat commands. This is global and not tied to
                Roblox link syncing.
              </Typography>
            </Box>

            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) minmax(340px, 0.95fr)" },
                gap: 2.5,
              }}
            >
              <Stack spacing={2}>
                <TextField
                  label="Chat command prefix"
                  value={settings.commandPrefix}
                  onChange={(event) => updateField("commandPrefix", event.target.value)}
                  disabled={saving || loading}
                  helperText="Default is !. Example prefixes: !, ?, ."
                  placeholder="!"
                />

                <Stack direction="row" spacing={1.25}>
                  <Button
                    variant="contained"
                    onClick={() => {
                      void handleSave();
                    }}
                    disabled={saving || loading}
                  >
                    {saving ? "Saving..." : "Save Prefix"}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      updateField("commandPrefix", "!");
                    }}
                    disabled={saving || loading}
                  >
                    Default (!)
                  </Button>
                </Stack>

                {message !== "" ? (
                  <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                    {message}
                  </Typography>
                ) : null}
              </Stack>

              <Card variant="outlined">
                <CardContent sx={{ p: 2 }}>
                  <Typography
                    sx={{
                      fontSize: "0.78rem",
                      fontWeight: 800,
                      textTransform: "uppercase",
                      letterSpacing: "0.08em",
                      color: "text.secondary",
                      mb: 1.4,
                    }}
                  >
                    Preview
                  </Typography>
                  <Typography sx={{ fontWeight: 700 }}>
                    Example: {commandPrefixDisplay}help
                  </Typography>
                  <Typography color="text.secondary" sx={{ mt: 0.8, fontSize: "0.9rem" }}>
                    The same prefix is used across command usage and UI previews.
                  </Typography>
                </CardContent>
              </Card>
            </Box>
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack spacing={2.5}>
            <Box>
              <Stack direction="row" spacing={1} alignItems="center">
                <OpenInNewRoundedIcon color="primary" />
                <Typography variant="h5">Public Home Promo Links</Typography>
              </Stack>
              <Typography color="text.secondary" sx={{ mt: 0.8, maxWidth: 820 }}>
                Add quick links you want to feature on the public home page, like merch, Discord,
                socials, or a schedule page.
              </Typography>
            </Box>

            <Stack spacing={1.5}>
              {settings.promoLinks.map((link, index) => (
                <Card key={`promo-link-${index}`} variant="outlined">
                  <CardContent sx={{ p: 2 }}>
                    <Stack spacing={1.5}>
                      <Stack
                        direction={{ xs: "column", md: "row" }}
                        spacing={1.5}
                        alignItems={{ xs: "stretch", md: "flex-start" }}
                      >
                        <TextField
                          label="Button label"
                          value={link.label}
                          onChange={(event) => updatePromoLink(index, "label", event.target.value)}
                          fullWidth
                        />
                        <TextField
                          label="URL"
                          value={link.href}
                          onChange={(event) => updatePromoLink(index, "href", event.target.value)}
                          placeholder="https://example.com"
                          fullWidth
                        />
                        <Button
                          variant="outlined"
                          color="error"
                          onClick={() => removePromoLink(index)}
                          sx={{ minWidth: { xs: "100%", md: 110 } }}
                        >
                          Remove
                        </Button>
                      </Stack>
                    </Stack>
                  </CardContent>
                </Card>
              ))}

              <Stack direction="row" spacing={1.25}>
                <Button
                  variant="outlined"
                  onClick={addPromoLink}
                  disabled={settings.promoLinks.length >= 6}
                >
                  Add Promo Link
                </Button>
                <Typography color="text.secondary" sx={{ alignSelf: "center", fontSize: "0.92rem" }}>
                  Up to 6 links. Empty rows will be ignored.
                </Typography>
              </Stack>

              <Stack direction="row" spacing={1.25} sx={{ pt: 0.5 }}>
                <Button
                  variant="contained"
                  onClick={() => {
                    void handleSave();
                  }}
                  disabled={saving || loading}
                >
                  {saving ? "Saving..." : "Save Promo Links"}
                </Button>
                <Button
                  variant="outlined"
                  onClick={() => {
                    setSettings(defaultSettings);
                    setMessage("");
                  }}
                  disabled={saving || loading}
                >
                  Reset
                </Button>
              </Stack>

              {message !== "" ? (
                <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                  {message}
                </Typography>
              ) : null}
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack spacing={2.5}>
            <Box>
              <Stack direction="row" spacing={1} alignItems="center">
                <LinkRoundedIcon color="primary" />
                <Typography variant="h5">Roblox Link Command</Typography>
              </Stack>
              <Typography color="text.secondary" sx={{ mt: 0.8, maxWidth: 820 }}>
                Choose who should own <strong>{`${commandPrefixDisplay}link`}</strong> after a moderator posts a Roblox
                private server link. DankBot can answer it directly, or it can push a chat
                management command into another bot.
              </Typography>
            </Box>

            <Alert severity="info">
              External bot sync is best-effort. Nightbot and Fossabot have good default templates.
              Pajbot and other bots can use the custom template field below. Use{" "}
              <code>{"{link}"}</code> where the private server URL should go.
            </Alert>

            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) minmax(340px, 0.95fr)" },
                gap: 2.5,
              }}
            >
              <Stack spacing={2}>
                <TextField
                  select
                  label="Link command owner"
                  value={settings.robloxLinkCommandTarget}
                  onChange={(event) =>
                    handleLinkTargetChange(
                      event.target.value as PublicHomeSettings["robloxLinkCommandTarget"],
                    )
                  }
                  disabled={saving || loading}
                >
                  <MenuItem value="dankbot">{`DankBot built-in ${commandPrefixDisplay}link`}</MenuItem>
                  <MenuItem value="nightbot">Nightbot</MenuItem>
                  <MenuItem value="fossabot">Fossabot</MenuItem>
                  <MenuItem value="pajbot">Pajbot</MenuItem>
                  <MenuItem value="custom">Custom bot command</MenuItem>
                </TextField>

                <TextField
                  label="External bot command template"
                  value={settings.robloxLinkCommandTemplate}
                  onChange={(event) =>
                    updateField("robloxLinkCommandTemplate", event.target.value)
                  }
                  disabled={
                    saving || loading || settings.robloxLinkCommandTarget === "dankbot"
                  }
                  helperText={
                    settings.robloxLinkCommandTarget === "nightbot"
                      ? "Nightbot example: !commands edit !link {link}"
                      : settings.robloxLinkCommandTarget === "fossabot"
                        ? "Fossabot example: !setcommand !link {link}"
                        : settings.robloxLinkCommandTarget === "pajbot"
                          ? "Pajbot example: !add command !link {link}"
                          : "Use {link} where the Roblox private server URL should be inserted."
                  }
                  multiline
                  minRows={2}
                  placeholder={
                    linkCommandTemplateDefaults[settings.robloxLinkCommandTarget] ||
                    "!commands edit !link {link}"
                  }
                />

                <TextField
                  label="External bot delete command template"
                  value={settings.robloxLinkCommandDeleteTemplate}
                  onChange={(event) =>
                    updateField("robloxLinkCommandDeleteTemplate", event.target.value)
                  }
                  disabled={
                    saving || loading || settings.robloxLinkCommandTarget === "dankbot"
                  }
                  helperText={
                    settings.robloxLinkCommandTarget === "nightbot"
                      ? "Nightbot example: !commands del !link"
                      : settings.robloxLinkCommandTarget === "fossabot"
                        ? "Fossabot example: !delcommand !link"
                        : settings.robloxLinkCommandTarget === "pajbot"
                          ? "Pajbot example: !delcommand !link"
                          : "Runs when link mode is turned off or switched to another mode."
                  }
                  multiline
                  minRows={2}
                  placeholder={
                    linkCommandDeleteTemplateDefaults[settings.robloxLinkCommandTarget] ||
                    "!commands del !link"
                  }
                />

                <Stack direction="row" spacing={1.25}>
                  <Button
                    variant="contained"
                    onClick={() => {
                      void handleSave();
                    }}
                    disabled={saving || loading}
                  >
                    {saving ? "Saving..." : "Save"}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setSettings(defaultSettings);
                      setMessage("");
                    }}
                    disabled={saving || loading}
                  >
                    Reset
                  </Button>
                </Stack>

                {message !== "" ? (
                  <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                    {message}
                  </Typography>
                ) : null}
              </Stack>

              <Card variant="outlined">
                <CardContent sx={{ p: 2 }}>
                  <Typography
                    sx={{
                      fontSize: "0.78rem",
                      fontWeight: 800,
                      textTransform: "uppercase",
                      letterSpacing: "0.08em",
                      color: "text.secondary",
                      mb: 1.4,
                    }}
                  >
                    Sync Preview
                  </Typography>

                  <Stack spacing={1.1}>
                    <Typography sx={{ fontWeight: 700 }}>
                      {settings.robloxLinkCommandTarget === "dankbot"
                        ? `DankBot will answer ${commandPrefixDisplay}link itself`
                        : "DankBot will send this command into chat"}
                    </Typography>
                    <Box
                      sx={{
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 1.25,
                        p: 1.5,
                        bgcolor: "background.default",
                        fontFamily: "monospace",
                        fontSize: "0.95rem",
                        wordBreak: "break-word",
                      }}
                    >
                      {settings.robloxLinkCommandTarget === "dankbot"
                        ? `${commandPrefixDisplay}link -> https://www.roblox.com/share?code=example&type=Server`
                        : (settings.robloxLinkCommandTemplate || "{link}").replace(
                            "{link}",
                            "https://www.roblox.com/share?code=example&type=Server",
                          )}
                    </Box>
                    <Typography color="text.secondary" sx={{ fontSize: "0.9rem" }}>
                      This runs when a moderator posts a Roblox private server link or when link
                      mode is explicitly set with a private server URL.
                    </Typography>

                    <Typography sx={{ mt: 0.8, fontWeight: 700 }}>
                      {settings.robloxLinkCommandTarget === "dankbot"
                        ? "No external delete command needed"
                        : "When link mode is turned off, DankBot will send"}
                    </Typography>
                    <Box
                      sx={{
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 1.25,
                        p: 1.5,
                        bgcolor: "background.default",
                        fontFamily: "monospace",
                        fontSize: "0.95rem",
                        wordBreak: "break-word",
                      }}
                    >
                      {settings.robloxLinkCommandTarget === "dankbot"
                        ? "-"
                        : settings.robloxLinkCommandDeleteTemplate || "!commands del !link"}
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            </Box>
          </Stack>
        </CardContent>
      </Card>
      </Stack>

      <ConfirmActionDialog
        open={editorToRemove != null}
        title="Remove editor access"
        description={
          editorToRemove == null
            ? ""
            : `${editorToRemove.displayName || editorToRemove.login} will lose their manual editor access. Twitch moderators will still keep dashboard access automatically.`
        }
        confirmLabel="Remove"
        onCancel={() => setEditorToRemove(null)}
        onConfirm={() => {
          void handleRemoveEditor();
        }}
      />
    </>
  );
}
