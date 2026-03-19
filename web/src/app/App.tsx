import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";

import { AuthProvider } from "./auth/AuthContext";
import { RequireModerator } from "./auth/RequireModerator";
import { RequireDashboardView } from "./auth/RequireDashboardView";
import { ForbiddenPage } from "./errors/ForbiddenPage";
import { NotFoundPage } from "./errors/NotFoundPage";
import { ModeratorProvider } from "./moderator/ModeratorContext";
import { ModeratorLayout } from "./moderator/ModeratorLayout";
import { AlertsPage } from "./moderator/pages/AlertsPage";
import { BlockedTermsPage } from "./moderator/pages/BlockedTermsPage";
import { ChannelPointsPage } from "./moderator/pages/ChannelPointsPage";
import { CommandsPage } from "./moderator/pages/CommandsPage";
import { DashboardOverviewPage } from "./moderator/pages/DashboardOverviewPage";
import { DashboardNotFoundPage } from "./moderator/pages/DashboardNotFoundPage";
import {
  DiscordCommandsPage,
  DiscordGamePingsPage,
  DiscordModerationPage,
  DiscordPage,
  DiscordRolePingsPage,
} from "./moderator/pages/DiscordPage";
import { GiveawayDashboardPage } from "./moderator/pages/GiveawayDashboardPage";
import { GiveawaysPage } from "./moderator/pages/GiveawaysPage";
import { IntegrationsPage } from "./moderator/pages/IntegrationsPage";
import { KeywordsPage } from "./moderator/pages/KeywordsPage";
import { ModesPage } from "./moderator/pages/ModesPage";
import { ModuleEditorPage } from "./moderator/pages/ModuleEditorPage";
import { ModulesPage } from "./moderator/pages/ModulesPage";
import { MassModerationPage } from "./moderator/pages/MassModerationPage";
import { SettingsPage } from "./moderator/pages/SettingsPage";
import { SpamFiltersPage } from "./moderator/pages/SpamFiltersPage";
import { TimersPage } from "./moderator/pages/TimersPage";
import { PublicCommandsPage } from "./public/PublicCommandsPage";
import { PublicHomePage } from "./public/PublicHomePage";
import { PublicLayout } from "./public/PublicLayout";
import { PublicProfilePage } from "./public/PublicProfilePage";
import { PublicQuotesPage } from "./public/PublicQuotesPage";

function ModeratorApp() {
  return (
    <ModeratorProvider>
      <ModeratorLayout />
    </ModeratorProvider>
  );
}

function LegacyDashboardRedirect() {
  const location = useLocation();
  const targetPath = location.pathname.replace(/^\/dashboard(?=\/|$)/, "/d") || "/d";
  const next = `${targetPath}${location.search}${location.hash}`;
  return <Navigate to={next} replace />;
}

export function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/dashboard/*" element={<LegacyDashboardRedirect />} />
          <Route element={<PublicLayout />}>
            <Route path="/" element={<PublicHomePage />} />
            <Route path="/commands" element={<PublicCommandsPage />} />
            <Route path="/quotes" element={<PublicQuotesPage />} />
            <Route path="/profile" element={<PublicProfilePage />} />
            <Route
              path="/user/:twitchUsernameRaw"
              element={<PublicProfilePage />}
            />
            <Route path="/403" element={<ForbiddenPage />} />
            <Route path="/404" element={<NotFoundPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Route>

          <Route
            path="/d"
            element={
              <RequireModerator>
                <ModeratorApp />
              </RequireModerator>
            }
          >
            <Route index element={<DashboardOverviewPage />} />
            <Route path="commands" element={<CommandsPage />} />
            <Route path="keywords" element={<KeywordsPage />} />
            <Route
              path="modes"
              element={
                <RequireDashboardView view="modes">
                  <ModesPage />
                </RequireDashboardView>
              }
            />
            <Route path="timers" element={<TimersPage />} />
            <Route path="modules" element={<ModulesPage />} />
            <Route path="modules/:moduleId" element={<ModuleEditorPage />} />
            <Route
              path="discord"
              element={
                <RequireDashboardView view="discord">
                  <DiscordPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="discord/commands"
              element={
                <RequireDashboardView view="discord">
                  <DiscordCommandsPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="discord/moderation"
              element={
                <RequireDashboardView view="discord">
                  <DiscordModerationPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="discord/role-pings"
              element={
                <RequireDashboardView view="discord">
                  <DiscordRolePingsPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="discord/game-pings"
              element={
                <RequireDashboardView view="discord">
                  <DiscordGamePingsPage />
                </RequireDashboardView>
              }
            />
            <Route path="alerts" element={<AlertsPage />} />
            <Route path="spam-filters" element={<SpamFiltersPage />} />
            <Route
              path="blocked-terms"
              element={
                <RequireDashboardView view="blockedTerms">
                  <BlockedTermsPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="mass-moderation"
              element={
                <RequireDashboardView view="massModeration">
                  <MassModerationPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="channel-points"
              element={
                <RequireDashboardView view="channelPoints">
                  <ChannelPointsPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="giveaways"
              element={
                <RequireDashboardView view="giveaways">
                  <GiveawaysPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="giveaways/:giveawayId"
              element={
                <RequireDashboardView view="giveaways">
                  <GiveawayDashboardPage />
                </RequireDashboardView>
              }
            />
            <Route
              path="integrations"
              element={
                <RequireDashboardView view="integrations">
                  <IntegrationsPage />
                </RequireDashboardView>
              }
            />
            <Route path="settings" element={<SettingsPage />} />
            <Route path="*" element={<DashboardNotFoundPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}
