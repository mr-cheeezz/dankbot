import { Navigate, useLocation } from "react-router-dom";

import { useAuth } from "./AuthContext";

export function RequireModerator({ children }: { children: JSX.Element }) {
  const location = useLocation();
  const { loading, session } = useAuth();

  if (loading) {
    return <div className="route-loading">checking dashboard access...</div>;
  }

  if (!session.canAccessDashboard) {
    return <Navigate to="/403" replace state={{ from: location.pathname }} />;
  }

  return children;
}
