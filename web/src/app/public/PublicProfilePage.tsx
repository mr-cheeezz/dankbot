import DiamondRoundedIcon from "@mui/icons-material/DiamondRounded";
import GavelRoundedIcon from "@mui/icons-material/GavelRounded";
import MilitaryTechRoundedIcon from "@mui/icons-material/MilitaryTechRounded";
import RedeemRoundedIcon from "@mui/icons-material/RedeemRounded";
import RecordVoiceOverRoundedIcon from "@mui/icons-material/RecordVoiceOverRounded";
import WorkspacePremiumRoundedIcon from "@mui/icons-material/WorkspacePremiumRounded";
import VerifiedRoundedIcon from "@mui/icons-material/VerifiedRounded";
import {
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { Link as RouterLink, Navigate, useParams } from "react-router-dom";

import { useAuth } from "../auth/AuthContext";
import { formatStreamerTitle } from "../title";
import {
  defaultPublicUserProfile,
  fetchPublicUserProfile,
  fetchPublicUserTabHistory,
  type PublicUserProfile,
} from "./api";

function initials(value: string) {
  const parts = value.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "DB";
  }

  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

function formatRelativeTime(value: string) {
  if (value.trim() === "") {
    return "No activity yet";
  }

  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return "No activity yet";
  }

  const diff = Math.max(0, Date.now() - timestamp);
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ago`;
  }
  if (hours > 0) {
    return `${hours}h ago`;
  }
  return `${Math.max(1, minutes)}m ago`;
}

function formatMoneyFromCents(cents: number) {
  const dollars = Math.abs(Math.trunc(cents)) / 100;
  return `${cents < 0 ? "-" : ""}$${dollars.toFixed(2)}`;
}

function formatTabAction(action: string) {
  const normalized = action.trim().toLowerCase();
  switch (normalized) {
    case "add":
      return "Added";
    case "set":
      return "Set";
    case "paid":
      return "Paid";
    case "give":
      return "Opened";
    default:
      return normalized === "" ? "Updated" : normalized;
  }
}

function ProfileStat({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <Card sx={{ height: "100%" }}>
      <CardContent sx={{ p: 2.25 }}>
        <Typography
          sx={{
            fontSize: "0.74rem",
            fontWeight: 700,
            letterSpacing: "0.08em",
            textTransform: "uppercase",
            color: "text.secondary",
          }}
        >
          {label}
        </Typography>
        <Typography sx={{ mt: 1, fontSize: "1.1rem", fontWeight: 700 }}>{value}</Typography>
      </CardContent>
    </Card>
  );
}

function streamRoleChipProps(role: PublicUserProfile["streamRole"]) {
  switch (role) {
    case "vip":
      return {
        label: "VIP",
        icon: <DiamondRoundedIcon fontSize="small" />,
        color: "secondary" as const,
      };
    case "moderator":
      return {
        label: "Moderator",
        icon: <GavelRoundedIcon fontSize="small" />,
        color: "info" as const,
      };
    case "lead_mod":
      return {
        label: "Lead Mod",
        icon: <MilitaryTechRoundedIcon fontSize="small" />,
        color: "warning" as const,
      };
    case "broadcaster":
      return {
        label: "Broadcaster",
        icon: <RecordVoiceOverRoundedIcon fontSize="small" />,
        color: "primary" as const,
      };
    case "viewer":
    default:
      return {
        label: "Viewer",
        icon: undefined,
        color: "default" as const,
      };
  }
}

export function PublicProfilePage() {
  const { twitchUsernameRaw } = useParams<{ twitchUsernameRaw?: string }>();
  const { session } = useAuth();
  const [profile, setProfile] = useState<PublicUserProfile>(defaultPublicUserProfile);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [fullHistory, setFullHistory] = useState<PublicUserProfile["recentTabEvents"]>([]);

  const requestedLogin = useMemo(() => {
    if (twitchUsernameRaw && twitchUsernameRaw.trim() !== "") {
      return twitchUsernameRaw.trim().toLowerCase();
    }
    if (session.user?.login) {
      return session.user.login.trim().toLowerCase();
    }
    return "";
  }, [session.user?.login, twitchUsernameRaw]);

  useEffect(() => {
    if (requestedLogin === "") {
      setProfile(defaultPublicUserProfile);
      setLoading(false);
      setNotFound(false);
      return;
    }

    const controller = new AbortController();
    setLoading(true);
    setNotFound(false);

    fetchPublicUserProfile(requestedLogin, controller.signal)
      .then((nextProfile) => {
        setProfile(nextProfile);
      })
      .catch((error) => {
        if (error instanceof Error && error.message === "not found") {
          setNotFound(true);
        }
        setProfile(defaultPublicUserProfile);
      })
      .finally(() => {
        setLoading(false);
      });

    return () => controller.abort();
  }, [requestedLogin]);

  if (!twitchUsernameRaw && session.loggedIn && session.user?.login) {
    return <Navigate to={`/user/${encodeURIComponent(session.user.login)}`} replace />;
  }

  if (requestedLogin === "") {
    return (
      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={1.25}>
            <Typography variant="h4">User Profile</Typography>
            <Typography color="text.secondary">
              Sign in with Twitch or open a profile directly at /user/twitchusername.
            </Typography>
            <Stack direction="row" spacing={1.25} sx={{ pt: 1 }}>
              <Button href="/auth/login" variant="contained">
                Login With Twitch
              </Button>
              <Button component={RouterLink} to="/" variant="outlined">
                Back Home
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  if (loading) {
    return (
      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack direction="row" spacing={1.25} alignItems="center">
            <CircularProgress size={22} />
            <Typography>Loading profile for @{requestedLogin}...</Typography>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  if (notFound) {
    return (
      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={1.25}>
            <Typography variant="h4">User not found</Typography>
            <Typography color="text.secondary">
              We could not find a Twitch user for @{requestedLogin}.
            </Typography>
            <Stack direction="row" spacing={1.25} sx={{ pt: 1 }}>
              <Button component={RouterLink} to="/" variant="contained">
                Back Home
              </Button>
              <Button
                href={`https://twitch.tv/${encodeURIComponent(requestedLogin)}`}
                target="_blank"
                rel="noreferrer"
                variant="outlined"
              >
                Try Twitch
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  const displayName = profile.displayName || formatStreamerTitle(profile.login);
  const isPartner = profile.broadcasterType === "partner";
  const isAffiliate = profile.broadcasterType === "affiliate";
  const roleChip = streamRoleChipProps(profile.streamRole);

  if (!profile.profileEnabled) {
    return (
      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack spacing={1.25}>
            <Typography variant="h4">{displayName}</Typography>
            <Typography color="text.secondary">
              This user profile is currently hidden by channel settings.
            </Typography>
            <Stack direction="row" spacing={1.25} sx={{ pt: 1 }}>
              <Button
                href={profile.twitchURL}
                target="_blank"
                rel="noreferrer"
                variant="contained"
              >
                Open Twitch Profile
              </Button>
              <Button component={RouterLink} to="/" variant="outlined">
                Back Home
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    );
  }

  const openFullHistory = () => {
    if (requestedLogin.trim() === "") {
      return;
    }
    setHistoryOpen(true);
    setHistoryLoading(true);

    fetchPublicUserTabHistory(requestedLogin)
      .then((items) => {
        setFullHistory(items);
      })
      .catch(() => {
        setFullHistory([]);
      })
      .finally(() => {
        setHistoryLoading(false);
      });
  };

  return (
    <Stack spacing={2.5}>
      <Card>
        <CardContent sx={{ p: 3 }}>
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={2.5}
            alignItems={{ xs: "flex-start", md: "center" }}
            justifyContent="space-between"
          >
            <Stack direction="row" spacing={2} alignItems="center">
              {profile.avatarURL !== "" ? (
                <Avatar src={profile.avatarURL} alt={displayName} sx={{ width: 88, height: 88 }} />
              ) : (
                <Avatar sx={{ width: 88, height: 88, bgcolor: "primary.dark", fontSize: "1.5rem" }}>
                  {initials(displayName)}
                </Avatar>
              )}

              <Box>
                <Stack direction="row" spacing={0.9} alignItems="center" flexWrap="wrap" useFlexGap>
                  <Typography variant="h4">{displayName}</Typography>
                  {isPartner ? <VerifiedRoundedIcon color="primary" fontSize="small" /> : null}
                  {isAffiliate ? <WorkspacePremiumRoundedIcon color="success" fontSize="small" /> : null}
                </Stack>
                <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                  @{profile.login}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.4 }}>
                  <Chip label={roleChip.label} icon={roleChip.icon} color={roleChip.color} />
                </Stack>
              </Box>
            </Stack>

            <Stack direction="row" spacing={1.25} flexWrap="wrap" useFlexGap>
              <Button
                href={profile.twitchURL}
                target="_blank"
                rel="noreferrer"
                variant="contained"
              >
                Open Twitch Profile
              </Button>
              {session.user?.login?.toLowerCase() === profile.login.toLowerCase() && session.user.canAccessDashboard ? (
                <Button component={RouterLink} to="/d" variant="outlined">
                  Open Dashboard
                </Button>
              ) : null}
            </Stack>
          </Stack>

          {profile.description !== "" ? (
            <Typography color="text.secondary" sx={{ mt: 2, lineHeight: 1.7 }}>
              {profile.description}
            </Typography>
          ) : null}
        </CardContent>
      </Card>

      <Box
        sx={{
          display: "grid",
          gap: 1.5,
          gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0, 1fr))" },
        }}
      >
        <ProfileStat label="Member Since" value={profile.createdAt ? new Date(profile.createdAt).toLocaleDateString() : "Unknown"} />
        <ProfileStat label="Last Seen" value={profile.showLastSeen ? formatRelativeTime(profile.lastSeenAt) : "Hidden"} />
        <ProfileStat label="Last Active" value={profile.showLastChatActivity ? formatRelativeTime(profile.lastChatActivityAt) : "Hidden"} />
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", xl: "minmax(0, 1.05fr) minmax(0, 0.95fr)" },
        }}
      >
        {profile.showTabSection ? (
          <Card>
            <CardContent sx={{ p: 3 }}>
              <Stack spacing={1.25}>
                <Stack direction="row" alignItems="center" justifyContent="space-between">
                  <Typography variant="h6">Tab Balance</Typography>
                  <Chip
                    color={profile.hasOpenTab ? "warning" : "default"}
                    variant={profile.hasOpenTab ? "filled" : "outlined"}
                    label={profile.hasOpenTab ? formatMoneyFromCents(profile.tabBalanceCents) : "No tab"}
                  />
                </Stack>
                <Typography color="text.secondary">
                  Recent tab activity for this user.
                </Typography>

                {profile.showTabHistory && profile.recentTabEvents.length > 0 ? (
                  <Stack spacing={1}>
                    {profile.recentTabEvents.map((item) => (
                      <Box
                        key={item.id}
                        sx={{
                          p: 1.35,
                          border: "1px solid",
                          borderColor: "divider",
                          borderRadius: 1.25,
                          bgcolor: "rgba(255,255,255,0.02)",
                        }}
                      >
                        <Stack direction="row" justifyContent="space-between" spacing={1}>
                          <Typography sx={{ fontWeight: 700 }}>
                            {formatTabAction(item.action)} • {formatMoneyFromCents(item.amountCents)}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {formatRelativeTime(item.createdAt)}
                          </Typography>
                        </Stack>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
                          Balance: {formatMoneyFromCents(item.balanceCents)}
                        </Typography>
                      </Box>
                    ))}
                    <Button variant="outlined" onClick={openFullHistory}>
                      Full History
                    </Button>
                  </Stack>
                ) : (
                  <Typography color="text.secondary">
                    No recent tab entries yet.
                  </Typography>
                )}
              </Stack>
            </CardContent>
          </Card>
        ) : null}

        {profile.showRedemptionActivity ? (
          <Card>
            <CardContent sx={{ p: 3 }}>
              <Stack spacing={1.25}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <RedeemRoundedIcon color="primary" />
                  <Typography variant="h6">Channel Point Activity</Typography>
                </Stack>
                <Typography color="text.secondary">
                  First pass of public user stats from redemption events.
                </Typography>

                {profile.redemptionStatsReady && profile.topRewards.length > 0 ? (
                  <Stack spacing={1.1} sx={{ pt: 1 }}>
                    {profile.topRewards.map((reward) => (
                      <Box
                        key={reward.rewardTitle}
                        sx={{
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "space-between",
                          gap: 1.25,
                          p: 1.5,
                          border: "1px solid",
                          borderColor: "divider",
                          borderRadius: 1.25,
                          bgcolor: "rgba(255,255,255,0.02)",
                        }}
                      >
                        <Box sx={{ minWidth: 0 }}>
                          <Typography sx={{ fontWeight: 700 }}>{reward.rewardTitle}</Typography>
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
                            {reward.redemptionCount} redemption{reward.redemptionCount === 1 ? "" : "s"}
                          </Typography>
                        </Box>
                        <Chip
                          label={`${reward.totalPointsSpent.toLocaleString()} pts`}
                          color="primary"
                          variant="outlined"
                        />
                      </Box>
                    ))}
                  </Stack>
                ) : (
                  <Typography color="text.secondary" sx={{ pt: 1 }}>
                    No stored redemption activity for this user yet.
                  </Typography>
                )}

                {profile.recentRedemptions.length > 0 ? (
                  <Stack spacing={1.1} sx={{ pt: 1 }}>
                    {profile.recentRedemptions.slice(0, 4).map((item) => (
                      <Box
                        key={`${item.rewardTitle}-${item.redeemedAt}-${item.rewardCost}`}
                        sx={{
                          p: 1.5,
                          border: "1px solid",
                          borderColor: "divider",
                          borderRadius: 1.25,
                          bgcolor: "rgba(255,255,255,0.02)",
                        }}
                      >
                        <Stack
                          direction="row"
                          justifyContent="space-between"
                          spacing={1.25}
                          alignItems="center"
                        >
                          <Typography sx={{ fontWeight: 700 }}>{item.rewardTitle}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {formatRelativeTime(item.redeemedAt)}
                          </Typography>
                        </Stack>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.45 }}>
                          {item.rewardCost.toLocaleString()} points • {item.status || "fulfilled"}
                        </Typography>
                      </Box>
                    ))}
                  </Stack>
                ) : null}
              </Stack>
            </CardContent>
          </Card>
        ) : null}
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
        }}
      >
        {profile.showPollStats ? (
          <Card>
            <CardContent sx={{ p: 3 }}>
              <Stack spacing={1.1}>
                <Typography variant="h6">Poll Stats</Typography>
                <Typography color="text.secondary">
                  Based on saved poll event snapshots for this channel.
                </Typography>
                <ProfileStat label="Poll Events" value={profile.pollCount.toLocaleString()} />
                <ProfileStat label="Polls Ended" value={profile.pollEndedCount.toLocaleString()} />
                <ProfileStat label="Last Poll" value={formatRelativeTime(profile.lastPollAt)} />
              </Stack>
            </CardContent>
          </Card>
        ) : null}
        {profile.showPredictionStats ? (
          <Card>
            <CardContent sx={{ p: 3 }}>
              <Stack spacing={1.1}>
                <Typography variant="h6">Prediction Stats</Typography>
                <Typography color="text.secondary">
                  Based on saved prediction event snapshots for this channel.
                </Typography>
                <ProfileStat label="Prediction Events" value={profile.predictionCount.toLocaleString()} />
                <ProfileStat label="Predictions Ended" value={profile.predictionEndedCount.toLocaleString()} />
                <ProfileStat label="Last Prediction" value={formatRelativeTime(profile.lastPredictionAt)} />
              </Stack>
            </CardContent>
          </Card>
        ) : null}
      </Box>

      <Dialog open={historyOpen} onClose={() => setHistoryOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Tab History</DialogTitle>
        <DialogContent dividers>
          {historyLoading ? (
            <Stack direction="row" spacing={1.25} alignItems="center">
              <CircularProgress size={20} />
              <Typography>Loading full history…</Typography>
            </Stack>
          ) : fullHistory.length > 0 ? (
            <Stack spacing={1}>
              {fullHistory.map((item) => (
                <Box
                  key={item.id}
                  sx={{
                    p: 1.25,
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 1.1,
                  }}
                >
                  <Stack direction="row" justifyContent="space-between" spacing={1}>
                    <Typography sx={{ fontWeight: 700 }}>
                      {formatTabAction(item.action)} • {formatMoneyFromCents(item.amountCents)}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {new Date(item.createdAt).toLocaleString()}
                    </Typography>
                  </Stack>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
                    Balance: {formatMoneyFromCents(item.balanceCents)}
                  </Typography>
                </Box>
              ))}
            </Stack>
          ) : (
            <Typography color="text.secondary">No history found yet.</Typography>
          )}
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
