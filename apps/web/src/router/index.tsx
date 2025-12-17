import React, { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import { Box, CircularProgress, Typography } from '@mui/material'
import { AppShell } from '~/components/layout'
import { ProtectedRoute, AdminRoute, ModeratorRoute, GuestRoute } from '~/components/RouteGuards'

// Lazy load pages for better initial bundle size
const Homepage = lazy(() => import('~/pages/Homepage'))
const VideoDetailPage = lazy(() => import('~/pages/VideoDetailPage'))
const TagPage = lazy(() => import('~/pages/TagPage'))
const LoginPage = lazy(() => import('~/pages/LoginPage'))
const RegisterPage = lazy(() => import('~/pages/RegisterPage'))
const ProfilePage = lazy(() => import('~/pages/ProfilePage'))

// Admin pages (lazy loaded)
const AdminDashboard = lazy(() =>
  import('~/pages/admin/AdminDashboard').catch(() => ({
    default: () => (
      <Box sx={{ p: 4 }}>
        <Typography variant="h4">Admin Dashboard</Typography>
        <Typography color="text.secondary">Coming soon...</Typography>
      </Box>
    ),
  }))
)

// Mod pages (lazy loaded)
const ModDashboard = lazy(() =>
  import('~/pages/mod/ModDashboard').catch(() => ({
    default: () => (
      <Box sx={{ p: 4 }}>
        <Typography variant="h4">Moderator Dashboard</Typography>
        <Typography color="text.secondary">Coming soon...</Typography>
      </Box>
    ),
  }))
)

const TranscriptEditor = lazy(() =>
  import('~/pages/mod/TranscriptEditor').catch(() => ({
    default: () => (
      <Box sx={{ p: 4 }}>
        <Typography variant="h4">Transcript Editor</Typography>
        <Typography color="text.secondary">Error loading editor...</Typography>
      </Box>
    ),
  }))
)

/**
 * Loading fallback component
 */
const PageLoader: React.FC = () => (
  <Box
    sx={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: 'calc(100vh - 64px)',
    }}
  >
    <CircularProgress color="primary" />
  </Box>
)

/**
 * AppRouter Component
 * Provides routing for the application with role-based access control
 */
const AppRouter: React.FC = () => {
  return (
    <Routes>
      {/* Guest-only routes (login, register) */}
      <Route element={<GuestRoute />}>
        <Route
          path="/login"
          element={
            <Suspense fallback={<PageLoader />}>
              <LoginPage />
            </Suspense>
          }
        />
        <Route
          path="/register"
          element={
            <Suspense fallback={<PageLoader />}>
              <RegisterPage />
            </Suspense>
          }
        />
      </Route>

      {/* Public routes with AppShell */}
      <Route path="/" element={<AppShell />}>
        <Route
          index
          element={
            <Suspense fallback={<PageLoader />}>
              <Homepage />
            </Suspense>
          }
        />
        <Route
          path="video/:id"
          element={
            <Suspense fallback={<PageLoader />}>
              <VideoDetailPage />
            </Suspense>
          }
        />
        <Route
          path="tag/:tagId"
          element={
            <Suspense fallback={<PageLoader />}>
              <TagPage />
            </Suspense>
          }
        />
        <Route
          path="search"
          element={
            <Suspense fallback={<PageLoader />}>
              <Homepage />
            </Suspense>
          }
        />
      </Route>

      {/* Admin routes */}
      <Route element={<AdminRoute />}>
        <Route path="/admin" element={<AppShell />}>
          <Route
            index
            element={
              <Suspense fallback={<PageLoader />}>
                <AdminDashboard />
              </Suspense>
            }
          />
          {/* Add more admin routes here */}
        </Route>
      </Route>

      {/* Moderator routes */}
      <Route element={<ModeratorRoute />}>
        <Route path="/mod" element={<AppShell />}>
          <Route
            index
            element={
              <Suspense fallback={<PageLoader />}>
                <ModDashboard />
              </Suspense>
            }
          />
          <Route
            path="videos/:videoId/transcript"
            element={
              <Suspense fallback={<PageLoader />}>
                <TranscriptEditor />
              </Suspense>
            }
          />
          {/* Add more mod routes here */}
        </Route>
      </Route>

      {/* Protected routes (requires authentication) */}
      <Route element={<ProtectedRoute />}>
        <Route path="/profile" element={<AppShell />}>
          <Route
            index
            element={
              <Suspense fallback={<PageLoader />}>
                <ProfilePage />
              </Suspense>
            }
          />
        </Route>
      </Route>

      {/* 404 Not Found */}
      <Route
        path="*"
        element={
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
              height: '100vh',
              gap: 2,
            }}
          >
            <Typography variant="h1">404</Typography>
            <Typography color="text.secondary">Page not found</Typography>
          </Box>
        }
      />
    </Routes>
  )
}

export default AppRouter
