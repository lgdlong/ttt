import React, { lazy, Suspense } from 'react'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { Box, CircularProgress } from '@mui/material'
import { AppShell } from '~/components/layout'

// Lazy load pages for better initial bundle size
const Homepage = lazy(() => import('~/pages/Homepage'))
const VideoDetailPage = lazy(() => import('~/pages/VideoDetailPage'))
const LoginPage = lazy(() => import('~/pages/LoginPage'))
const RegisterPage = lazy(() => import('~/pages/RegisterPage'))

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
 * App Router Configuration
 */
const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      {
        index: true,
        element: (
          <Suspense fallback={<PageLoader />}>
            <Homepage />
          </Suspense>
        ),
      },
      {
        path: 'video/:id',
        element: (
          <Suspense fallback={<PageLoader />}>
            <VideoDetailPage />
          </Suspense>
        ),
      },
      {
        path: 'search',
        element: (
          <Suspense fallback={<PageLoader />}>
            {/* TODO: Create SearchPage component */}
            <Homepage />
          </Suspense>
        ),
      },
    ],
  },
  {
    path: '/login',
    element: (
      <Suspense fallback={<PageLoader />}>
        <LoginPage />
      </Suspense>
    ),
  },
  {
    path: '/register',
    element: (
      <Suspense fallback={<PageLoader />}>
        <RegisterPage />
      </Suspense>
    ),
  },
])

/**
 * AppRouter Component
 * Provides routing for the application
 */
const AppRouter: React.FC = () => {
  return <RouterProvider router={router} />
}

export default AppRouter
