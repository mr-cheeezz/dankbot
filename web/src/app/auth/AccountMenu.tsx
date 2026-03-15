import {
  Avatar,
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
} from "@mui/material";
import PersonRoundedIcon from "@mui/icons-material/PersonRounded";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import { useAuth } from "./AuthContext";

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

type AccountMenuProps = {
  className?: string;
};

export function AccountMenu({ className = "" }: AccountMenuProps) {
  const { session, logout } = useAuth();
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const navigate = useNavigate();

  const user = session.user;

  const fallbackInitials = useMemo(() => {
    if (!user) {
      return "DB";
    }

    return initials(user.displayName || user.login);
  }, [user]);

  if (!user) {
    return null;
  }

  const handleLogout = async () => {
    setAnchorEl(null);
    await logout();
    window.location.assign("/");
  };

  const handleProfile = () => {
    setAnchorEl(null);
    navigate(`/user/${encodeURIComponent(user.login)}`);
  };

  return (
    <>
      <IconButton
        className={className}
        onClick={(event) => setAnchorEl(event.currentTarget)}
        aria-haspopup="menu"
        aria-expanded={anchorEl ? "true" : undefined}
        sx={{
          width: 40,
          height: 40,
          border: "1px solid",
          borderColor: "divider",
          backgroundColor: "background.paper",
        }}
      >
        {user.avatarURL !== "" ? (
          <Avatar
            src={user.avatarURL}
            alt={user.displayName}
            sx={{ width: 32, height: 32 }}
          />
        ) : (
          <Avatar sx={{ width: 32, height: 32, bgcolor: "primary.dark", fontSize: "0.8rem" }}>
            {fallbackInitials}
          </Avatar>
        )}
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
        PaperProps={{
          sx: {
            mt: 1,
            minWidth: 220,
            backgroundColor: "background.paper",
          },
        }}
      >
        <MenuItem disabled sx={{ opacity: 1 }}>
          <ListItemText
            primary={user.displayName || user.login}
            secondary={`@${user.login}`}
            primaryTypographyProps={{ fontWeight: 700 }}
          />
        </MenuItem>
        <MenuItem onClick={handleProfile}>
          <ListItemIcon>
            <PersonRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Profile" />
        </MenuItem>
        <MenuItem onClick={handleLogout}>
          <ListItemIcon>
            <LogoutRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Logout" />
        </MenuItem>
      </Menu>
    </>
  );
}
