import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '~/providers/AuthProvider'
import { Box, CircularProgress, Typography } from '@mui/material'

/**
 * Protected Route - requires authentication
 */
export function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
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
        <CircularProgress />
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (!isAuthenticated) {
    // Redirect to login with return url
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <Outlet />
}

/**
 * Admin Route - requires admin role
 */
export function AdminRoute() {
  const { isAuthenticated, isAdmin, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
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
        <CircularProgress />
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (!isAdmin) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}

/**
 * Moderator Route - requires mod or admin role
 */
export function ModeratorRoute() {
  const { isAuthenticated, isModerator, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
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
        <CircularProgress />
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (!isModerator) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}

/**
 * Guest Route - only accessible when not logged in
 */
export function GuestRoute() {
  const { isAuthenticated, isLoading, user } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return (
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
        <CircularProgress />
        <Typography color="text.secondary">Loading...</Typography>
      </Box>
    )
  }

  if (isAuthenticated && user) {
    // Redirect based on role
    const from = (location.state as { from?: Location })?.from?.pathname
    if (from) {
      return <Navigate to={from} replace />
    }

    // Redirect based on role
    if (user.role === 'admin') {
      return <Navigate to="/admin" replace />
    } else if (user.role === 'mod') {
      return <Navigate to="/mod" replace />
    } else {
      return <Navigate to="/" replace />
    }
  }

  return <Outlet />
}
