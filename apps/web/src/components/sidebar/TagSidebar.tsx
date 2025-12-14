import React, { useState, useMemo, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Box,
  TextField,
  Typography,
  InputAdornment,
  Skeleton,
  List,
  ListItemButton,
  ListItemText,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import LocalOfferIcon from '@mui/icons-material/LocalOffer'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useTags } from '~/hooks'

/**
 * TagSidebar - Sidebar component with virtualized tag list and search
 * Uses TanStack Virtual for performance with large tag lists
 */
const TagSidebar: React.FC = () => {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState('')
  const { data: tags = [], isLoading } = useTags()
  const parentRef = useRef<HTMLDivElement>(null)

  // Filter tags based on search query
  const filteredTags = useMemo(() => {
    if (!searchQuery.trim()) return tags
    const query = searchQuery.toLowerCase()
    return tags.filter((tag) => tag.name.toLowerCase().includes(query))
  }, [tags, searchQuery])

  // Set up virtualizer for performance with large lists
  const virtualizer = useVirtualizer({
    count: filteredTags.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 40, // Estimated row height
    overscan: 5, // Render extra items for smooth scrolling
  })

  const handleTagClick = useCallback(
    (tagId: string) => {
      navigate(`/tag/${tagId}`)
    },
    [navigate]
  )

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value)
  }, [])

  if (isLoading) {
    return (
      <Box sx={{ p: 2 }}>
        <Skeleton variant="rounded" height={40} sx={{ mb: 2 }} />
        {Array.from({ length: 10 }).map((_, i) => (
          <Skeleton key={i} variant="text" height={40} sx={{ mb: 0.5 }} />
        ))}
      </Box>
    )
  }

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
        borderRight: 1,
        borderColor: 'divider',
      }}
    >
      {/* Header */}
      <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
        <Typography
          variant="subtitle1"
          fontWeight={600}
          sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}
        >
          <LocalOfferIcon fontSize="small" />
          Tags
        </Typography>

        {/* Search Input */}
        <TextField
          size="small"
          fullWidth
          placeholder="Tìm tag..."
          value={searchQuery}
          onChange={handleSearchChange}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="action" />
                </InputAdornment>
              ),
            },
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
              bgcolor: 'background.default',
            },
          }}
        />
      </Box>

      {/* Tag List with Virtual Scroll */}
      <Box
        ref={parentRef}
        sx={{
          flex: 1,
          overflow: 'auto',
          '&::-webkit-scrollbar': { width: '6px' },
          '&::-webkit-scrollbar-thumb': {
            backgroundColor: 'action.hover',
            borderRadius: '4px',
          },
        }}
      >
        {filteredTags.length === 0 ? (
          <Box sx={{ p: 2, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              {searchQuery ? 'Không tìm thấy tag nào' : 'Chưa có tag nào'}
            </Typography>
          </Box>
        ) : (
          <List
            disablePadding
            sx={{
              height: `${virtualizer.getTotalSize()}px`,
              width: '100%',
              position: 'relative',
            }}
          >
            {virtualizer.getVirtualItems().map((virtualItem) => {
              const tag = filteredTags[virtualItem.index]
              return (
                <ListItemButton
                  key={tag.id}
                  onClick={() => handleTagClick(tag.id)}
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: `${virtualItem.size}px`,
                    transform: `translateY(${virtualItem.start}px)`,
                    borderRadius: 1,
                    mx: 0.5,
                    '&:hover': {
                      bgcolor: 'action.hover',
                    },
                  }}
                >
                  <ListItemText
                    primary={tag.name}
                    primaryTypographyProps={{
                      variant: 'body2',
                      noWrap: true,
                      sx: { fontWeight: 500 },
                    }}
                  />
                </ListItemButton>
              )
            })}
          </List>
        )}
      </Box>

      {/* Footer with count */}
      <Box
        sx={{
          p: 1.5,
          borderTop: 1,
          borderColor: 'divider',
          bgcolor: 'background.default',
        }}
      >
        <Typography variant="caption" color="text.secondary">
          {filteredTags.length} tags
          {searchQuery && ` (từ ${tags.length})`}
        </Typography>
      </Box>
    </Box>
  )
}

export default TagSidebar
