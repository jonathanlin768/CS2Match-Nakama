import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom'
import './index.css'
import { AuthProvider } from './context/AuthContext'
import AppLayout from './layout/AppLayout'
import ProtectedRoute from './components/ProtectedRoute'
import Home from './pages/Home'
import LoginPage from './pages/LoginPage'
import MatchPage from './pages/MatchPage'
import GachaPage from './pages/GachaPage'
import RankingPage from './pages/RankingPage'
import ProfilePage from './pages/ProfilePage'
import ProfileTab from './pages/profile/ProfileTab'
import FriendsTab from './pages/profile/FriendsTab'
import MessagesTab from './pages/profile/MessagesTab'
import CoinsTab from './pages/profile/CoinsTab'
import CouponsTab from './pages/profile/CouponsTab'
import GemsTab from './pages/profile/GemsTab'
import TasksTab from './pages/profile/TasksTab'
import VipTab from './pages/profile/VipTab'

const router = createBrowserRouter([
  {
    path: '/',
    element: <LoginPage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: 'home', element: <Home /> },
          { path: 'match', element: <MatchPage /> },
          { path: 'gacha', element: <GachaPage /> },
          { path: 'ranking', element: <RankingPage /> },
          {
            path: 'profile',
            element: <ProfilePage />,
            children: [
              { index: true, element: <Navigate to="/profile/me" replace /> },
              { path: 'me', element: <ProfileTab /> },
              { path: 'friends', element: <FriendsTab /> },
              { path: 'messages', element: <MessagesTab /> },
              { path: 'coins', element: <CoinsTab /> },
              { path: 'coupons', element: <CouponsTab /> },
              { path: 'gems', element: <GemsTab /> },
              { path: 'tasks', element: <TasksTab /> },
              { path: 'vip', element: <VipTab /> },
            ],
          },
        ],
      },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  </StrictMode>,
)
