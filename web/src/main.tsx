import React from "react";
import ReactDOM from "react-dom/client";

import { App } from "./app/App";
import { ThemeModeProvider } from "./app/mui/ModeratorThemeProvider";
import "./app/styles.scss";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ThemeModeProvider>
      <App />
    </ThemeModeProvider>
  </React.StrictMode>,
);
