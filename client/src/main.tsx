import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { createBrowserRouter, Navigate, RouterProvider } from "react-router-dom"
import { Toaster } from "sonner"
import "./index.css"
import { AuthProvider } from "./context/AuthContext"
import AppShell from "./components/AppShell"
import HomePage from "./pages/HomePage"
import MatchPage from "./pages/MatchPage"
import TutorialPage from "./pages/TutorialPage"
import BattlePage from "./pages/BattlePage"
import FriendsPage from "./pages/FriendsPage"

const router = createBrowserRouter([{
  element: <AppShell />,
  children: [
    { path: "/", element: <HomePage /> },
    { path: "/match", element: <MatchPage /> },
    { path: "/tutorial", element: <TutorialPage /> },
    { path: "/battle", element: <BattlePage /> },
    { path: "/friends", element: <FriendsPage /> },
    { path: "*", element: <Navigate to="/" replace /> },
  ],
}])

createRoot(document.getElementById("root")!).render(
  <StrictMode><AuthProvider><RouterProvider router={router} /><Toaster theme="dark" position="top-center" /></AuthProvider></StrictMode>,
)
