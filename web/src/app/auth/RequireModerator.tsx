import { useLocation } from "react-router-dom";

import { useAuth } from "./AuthContext";
import { ForbiddenPage } from "../errors/ForbiddenPage";

export function RequireModerator({ children }: { children: JSX.Element }) {
  const location = useLocation();
  const { loading, session } = useAuth();

  if (loading) {
    return <div className="route-loading">checking dashboard access...</div>;
  }

  if (!session.canAccessDashboard) {
    return <ForbiddenPage fromPath={location.pathname} />;
  }

  return children;
}
