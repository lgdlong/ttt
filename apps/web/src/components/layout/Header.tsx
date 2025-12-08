import React from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Box,
  Typography,
  IconButton,
  Avatar,
  InputBase,
  Menu,
  MenuItem,
  Divider,
  ListItemIcon,
  ListItemText,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import LogoutIcon from '@mui/icons-material/Logout'
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings'
import PersonIcon from '@mui/icons-material/Person'
import { useAuth } from '~/providers/AuthProvider'

/**
 * Header Component
 * - Sticky position, 64px height
 * - Logo on left, Search bar in center, Avatar on right
 */
const Header: React.FC = () => {
  const navigate = useNavigate()
  const { user, isAuthenticated, logout } = useAuth()
  const [searchQuery, setSearchQuery] = React.useState('')
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null)
  const menuOpen = Boolean(anchorEl)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      // TODO: Implement search navigation when API is ready
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
    }
  }

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleMenuClose = () => {
    setAnchorEl(null)
  }

  const handleLogout = async () => {
    handleMenuClose()
    await logout()
    navigate('/login')
  }

  const handleAdminClick = () => {
    handleMenuClose()
    navigate('/admin')
  }

  const handleProfileClick = () => {
    handleMenuClose()
    navigate('/profile')
  }

  return (
    <Box
      component="header"
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 1100,
        bgcolor: 'background.paper',
        borderBottom: '1px solid',
        borderColor: 'divider',
        height: 64,
        display: 'flex',
        alignItems: 'center',
        px: 3,
        gap: 2,
      }}
    >
      {/* Logo */}
      <Typography
        variant="h6"
        fontWeight={800}
        color="primary.main"
        sx={{ cursor: 'pointer', flexShrink: 0 }}
        onClick={() => navigate('/')}
      >
        TTT ARCHIVE
      </Typography>

      {/* Search Bar */}
      <Box
        component="form"
        onSubmit={handleSearch}
        sx={{
          flex: 1,
          maxWidth: 600,
          mx: 'auto',
          bgcolor: '#F1F5F9',
          borderRadius: 99,
          display: 'flex',
          alignItems: 'center',
          px: 2,
          py: 0.5,
        }}
      >
        <InputBase
          placeholder="Tìm kiếm..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          sx={{
            flex: 1,
            '& input': {
              padding: '8px',
            },
          }}
        />
        <IconButton type="submit" size="small">
          <SearchIcon />
        </IconButton>
      </Box>

      {/* User Avatar */}
      {isAuthenticated ? (
        <>
          <Avatar
            sx={{
              width: 40,
              height: 40,
              cursor: 'pointer',
              flexShrink: 0,
              bgcolor: 'primary.main',
              fontSize: '0.875rem',
              fontWeight: 'bold',
            }}
            onClick={handleMenuOpen}
          >
            {user?.full_name ? user.full_name.charAt(0).toUpperCase() : 'U'}
          </Avatar>

          <Menu
            anchorEl={anchorEl}
            open={menuOpen}
            onClose={handleMenuClose}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
          >
            {/* User Info Header */}
            <MenuItem disabled sx={{ flexDirection: 'column', alignItems: 'flex-start', gap: 0.5 }}>
              <Typography variant="subtitle2" fontWeight={600}>
                {user?.full_name || user?.username || 'User'}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {user?.email}
              </Typography>
            </MenuItem>

            <Divider />

            {/* Profile */}
            <MenuItem onClick={handleProfileClick}>
              <ListItemIcon>
                <PersonIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Profile</ListItemText>
            </MenuItem>

            {/* Admin Panel - Only show if user is admin or mod */}
            {user?.role && (user.role === 'admin' || user.role === 'mod') && (
              <MenuItem onClick={handleAdminClick}>
                <ListItemIcon>
                  <AdminPanelSettingsIcon fontSize="small" color="primary" />
                </ListItemIcon>
                <ListItemText>{user.role === 'admin' ? 'Admin Panel' : 'Mod Panel'}</ListItemText>
              </MenuItem>
            )}

            <Divider />

            {/* Logout */}
            <MenuItem onClick={handleLogout} sx={{ color: 'error.main' }}>
              <ListItemIcon sx={{ color: 'inherit' }}>
                <LogoutIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Logout</ListItemText>
            </MenuItem>
          </Menu>
        </>
      ) : (
        <Avatar
          sx={{
            width: 40,
            height: 40,
            cursor: 'pointer',
            flexShrink: 0,
            bgcolor: 'action.hover',
          }}
          onClick={() => navigate('/login')}
        />
      )}
    </Box>
  )
}

export default Header
