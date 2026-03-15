import PersonRoundedIcon from "@mui/icons-material/PersonRounded";
import RedeemRoundedIcon from "@mui/icons-material/RedeemRounded";
import ShieldRoundedIcon from "@mui/icons-material/ShieldRounded";
import StackedBarChartRoundedIcon from "@mui/icons-material/StackedBarChartRounded";
import {
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
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

export function PublicProfilePage() {
  const { twitchUsernameRaw } = useParams<{ twitchUsernameRaw?: string }>();
  const { session } = useAuth();
  const [profile, setProfile] = useState<PublicUserProfile>(defaultPublicUserProfile);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

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
  const roleLabel =
    profile.broadcasterType === "partner"
      ? "Partner"
      : profile.broadcasterType === "affiliate"
        ? "Affiliate"
        : "Viewer";

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
                <Typography variant="h4">{displayName}</Typography>
                <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                  @{profile.login}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.4 }}>
                  <Chip color="primary" icon={<ShieldRoundedIcon />} label={roleLabel} />
                  {profile.redemptionStatsReady ? (
                    <Chip
                      color="success"
                      icon={<RedeemRoundedIcon />}
                      label={`${profile.redemptionCount} redemption${profile.redemptionCount === 1 ? "" : "s"}`}
                    />
                  ) : (
                    <Chip
                      color="default"
                      icon={<StackedBarChartRoundedIcon />}
                      label="Stats still growing"
                    />
                  )}
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
                <Button component={RouterLink} to="/dashboard" variant="outlined">
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
          gridTemplateColumns: { xs: "1fr", md: "repeat(4, minmax(0, 1fr))" },
        }}
      >
        <ProfileStat label="Twitch Login" value={`@${profile.login}`} />
        <ProfileStat label="Member Since" value={profile.createdAt ? new Date(profile.createdAt).toLocaleDateString() : "Unknown"} />
        <ProfileStat label="Points Spent" value={profile.totalPointsSpent.toLocaleString()} />
        <ProfileStat label="Last Redemption" value={formatRelativeTime(profile.lastRedeemedAt)} />
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", xl: "minmax(0, 1.05fr) minmax(0, 0.95fr)" },
        }}
      >
        <Card>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={1.25}>
              <Stack direction="row" spacing={1} alignItems="center">
                <RedeemRoundedIcon color="primary" />
                <Typography variant="h6">Channel Point Activity</Typography>
              </Stack>
              <Typography color="text.secondary">
                First pass of public user stats, built from the redemption events DankBot is already
                storing.
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
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={1.25}>
              <Stack direction="row" spacing={1} alignItems="center">
                <PersonRoundedIcon color="primary" />
                <Typography variant="h6">Recent Activity</Typography>
              </Stack>
              <Typography color="text.secondary">
                This is where we can grow into fuller chat stats and user activity over time.
              </Typography>

              {profile.recentRedemptions.length > 0 ? (
                <Stack spacing={1.1} sx={{ pt: 1 }}>
                  {profile.recentRedemptions.map((item) => (
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
                      {item.userInput !== "" ? (
                        <Typography sx={{ mt: 0.75, fontSize: "0.95rem" }}>{item.userInput}</Typography>
                      ) : null}
                    </Box>
                  ))}
                </Stack>
              ) : (
                <Typography color="text.secondary" sx={{ pt: 1 }}>
                  Recent activity will show up here once this user starts using channel point rewards.
                </Typography>
              )}
            </Stack>
          </CardContent>
        </Card>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
        }}
      >
        <Card>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={1.1}>
              <Typography variant="h6">Chat Stats</Typography>
              <Typography color="text.secondary">
                We do not store per-user chat message history yet, so this section is the next good
                place to grow.
              </Typography>
            </Stack>
          </CardContent>
        </Card>
        <Card>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={1.1}>
              <Typography variant="h6">Poll Stats</Typography>
              <Typography color="text.secondary">
                Poll summaries are stored, but not per-user votes yet. Once we track that, this page
                can show poll participation and prediction history too.
              </Typography>
            </Stack>
          </CardContent>
        </Card>
      </Box>
    </Stack>
  );
}
