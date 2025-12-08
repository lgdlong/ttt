import React from 'react'
import { useNavigate } from 'react-router-dom'
import { Box, Typography, IconButton, Avatar, InputBase } from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'

/**
 * Header Component
 * - Sticky position, 64px height
 * - Logo on left, Search bar in center, Avatar on right
 */
const Header: React.FC = () => {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = React.useState('')

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      // TODO: Implement search navigation when API is ready
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
    }
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

      {/* User Avatar - Navigate to login if not authenticated */}
      {/* TODO: Replace with actual user data from auth API */}
      <Avatar
        sx={{
          width: 32,
          height: 32,
          cursor: 'pointer',
          flexShrink: 0,
        }}
        onClick={() => navigate('/login')}
      />
    </Box>
  )
}

export default Header
