import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { ApolloProvider } from "@apollo/client/react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { ThemeController } from "./types/site_settings/ThemeController";
import { IsMobileProvider } from "./common/IsMobileProvider";
import { useEffect, useState } from "react";
import { ToastContainer, Slide } from "react-toastify";
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { MakeThemeProvider } from "./common/MakeThemeProvider";
import { appRouter } from "./AppRouter";
import { RouterProvider } from "react-router-dom";


const link = new HttpLink({
  uri: import.meta.env.VITE_GRAPHQL_URL ?? "http://localhost:8080/graphql", 
  credentials: "include",
});

const apolloClient = new ApolloClient({
  link: link,
  cache: new InMemoryCache(),
});

export default function App() {
  const [theme, setTheme] = useState(ThemeController.activeTheme.getTheme());

  ThemeController.addThemeWatcher(setTheme);

  useEffect(() => {
    // Executes once when client is loaded
    ThemeController.setActiveTheme(localStorage.getItem("themeMode") ?? "")
  }, [])

  return (
    <ApolloProvider client={apolloClient}>
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <MakeThemeProvider>
          <ThemeProvider theme={theme}>
            <CssBaseline />
            <IsMobileProvider>
              <>
                <RouterProvider router={appRouter} />
                <ToastContainer position="bottom-left" transition={Slide} />
              </>
            </IsMobileProvider>
          </ThemeProvider>
        </MakeThemeProvider>
      </LocalizationProvider>
    </ApolloProvider>
  );
}
