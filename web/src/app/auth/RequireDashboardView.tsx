import { Navigate, useLocation } from "react-router-dom";

import type { ViewKey } from "../moderator/types";
import { canAccessDashboardView } from "./dashboardPermissions";
import { useAuth } from "./AuthContext";

export function RequireDashboardView({
  view,
  children,
}: {
  view: ViewKey;
  children: JSX.Element;
}) {
  const location = useLocation();
  const { loading, session } = useAuth();

  if (loading) {
    return <div className="route-loading">checking page access...</div>;
  }

  if (!canAccessDashboardView(session, view)) {
    return <Navigate to="/403" replace state={{ from: location.pathname }} />;
  }

  return children;
}
