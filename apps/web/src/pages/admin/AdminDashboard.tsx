import React, { useState } from 'react'
import {
  Box,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  Divider,
  useTheme,
  useMediaQuery,
  IconButton,
  AppBar,
  Toolbar,
} from '@mui/material'
import {
  People as PeopleIcon,
  Dashboard as DashboardIcon,
  Menu as MenuIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material'
import { useAuth } from '~/providers/AuthProvider'
import AdminOverview from './AdminOverview'
import UserManagement from './UserManagement'

const DRAWER_WIDTH = 260

type AdminView = 'overview' | 'users' | 'settings'

/**
 * AdminDashboard Component
 * Admin control panel with sidebar navigation
 * - User management with ban/search functionality
 * - Role-based display (user/mod/admin)
 */
const AdminDashboard: React.FC = () => {
  const { user } = useAuth()
  const theme = useTheme()
  const isMobile = useMediaQuery(theme.breakpoints.down('md'))
  const [mobileOpen, setMobileOpen] = useState(false)
  const [currentView, setCurrentView] = useState<AdminView>('users')

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen)
  }

  const menuItems = [
    { id: 'overview' as AdminView, label: 'Tổng quan', icon: <DashboardIcon /> },
    { id: 'users' as AdminView, label: 'Quản lý người dùng', icon: <PeopleIcon /> },
    { id: 'settings' as AdminView, label: 'Cài đặt', icon: <SettingsIcon /> },
  ]

  const drawer = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <Box sx={{ p: 2, borderBottom: '1px solid', borderColor: 'divider' }}>
        <Typography variant="h6" fontWeight={700} color="primary.main">
          Admin Panel
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {user?.full_name || user?.username}
        </Typography>
      </Box>

      {/* Navigation */}
      <List sx={{ flex: 1, py: 2 }}>
        {menuItems.map((item) => (
          <ListItem key={item.id} disablePadding>
            <ListItemButton
              selected={currentView === item.id}
              onClick={() => {
                setCurrentView(item.id)
                if (isMobile) setMobileOpen(false)
              }}
              sx={{
                mx: 1,
                borderRadius: 0,
                '&.Mui-selected': {
                  bgcolor: 'primary.main',
                  color: 'white',
                  '&:hover': {
                    bgcolor: 'primary.dark',
                  },
                  '& .MuiListItemIcon-root': {
                    color: 'white',
                  },
                },
              }}
            >
              <ListItemIcon sx={{ minWidth: 40 }}>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>

      {/* Footer */}
      <Divider />
      <Box sx={{ p: 2 }}>
        <Typography variant="caption" color="text.secondary">
          TTT Admin v1.0
        </Typography>
      </Box>
    </Box>
  )

  const renderContent = () => {
    switch (currentView) {
      case 'overview':
        return <AdminOverview />
      case 'users':
        return <UserManagement />
      case 'settings':
        return (
          <Box sx={{ p: 3 }}>
            <Typography variant="h5" fontWeight={600} gutterBottom>
              Cài đặt
            </Typography>
            <Typography color="text.secondary">Chức năng đang phát triển...</Typography>
          </Box>
        )
      default:
        return <AdminOverview />
    }
  }

  return (
    <Box sx={{ display: 'flex', minHeight: 'calc(100vh - 64px)' }}>
      {/* Mobile AppBar */}
      {isMobile && (
        <AppBar
          position="fixed"
          sx={{
            top: 64,
            bgcolor: 'background.paper',
            borderBottom: '1px solid',
            borderColor: 'divider',
          }}
          elevation={0}
        >
          <Toolbar>
            <IconButton
              color="inherit"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2, color: 'text.primary' }}
            >
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" color="text.primary" fontWeight={600}>
              {menuItems.find((m) => m.id === currentView)?.label}
            </Typography>
          </Toolbar>
        </AppBar>
      )}

      {/* Sidebar Drawer */}
      <Box component="nav" sx={{ width: { md: DRAWER_WIDTH }, flexShrink: { md: 0 } }}>
        {/* Mobile drawer */}
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{ keepMounted: true }}
          sx={{
            display: { xs: 'block', md: 'none' },
            '& .MuiDrawer-paper': {
              boxSizing: 'border-box',
              width: DRAWER_WIDTH,
              top: 64,
              height: 'calc(100% - 64px)',
            },
          }}
        >
          {drawer}
        </Drawer>

        {/* Desktop drawer */}
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', md: 'block' },
            '& .MuiDrawer-paper': {
              boxSizing: 'border-box',
              width: DRAWER_WIDTH,
              position: 'relative',
              height: '100%',
              border: 'none',
              borderRight: '1px solid',
              borderColor: 'divider',
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          width: { md: `calc(100% - ${DRAWER_WIDTH}px)` },
          mt: { xs: 7, md: 0 },
          bgcolor: 'background.default',
          minHeight: '100%',
        }}
      >
        {renderContent()}
      </Box>
    </Box>
  )
}

export default AdminDashboard
