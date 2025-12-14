import React from 'react'
import { Outlet } from 'react-router-dom'
import { Box } from '@mui/material'
import Header from './Header'

/**
 * AppShell - Main layout wrapper
 * Provides consistent header and content area for all pages
 */
const AppShell: React.FC = () => {
  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <Header />
      <Box component="main">
        <Outlet />
      </Box>
    </Box>
  )
}

export default AppShell
