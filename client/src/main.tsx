import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './index.css'
import AppLayout from './layout/AppLayout'
import Home from './pages/Home'
import LoginPage from './pages/LoginPage'
import MatchPage from './pages/MatchPage'
import GachaPage from './pages/GachaPage'
import RankingPage from './pages/RankingPage'

const router = createBrowserRouter([
  {
    path: '/',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { path: 'home', element: <Home /> },
      { path: 'match', element: <MatchPage /> },
      { path: 'gacha', element: <GachaPage /> },
      { path: 'ranking', element: <RankingPage /> },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
