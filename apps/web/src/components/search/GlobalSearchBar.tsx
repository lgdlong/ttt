import React, { useState, useCallback, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Box,
  TextField,
  InputAdornment,
  Paper,
  List,
  ListItemButton,
  ListItemText,
  Typography,
  Chip,
  Stack,
  CircularProgress,
  ClickAwayListener,
  Popper,
  Fade,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import { useVideoSearch } from '~/hooks'

/**
 * GlobalSearchBar - Search bar component for navbar
 * Searches videos by title and tags with debounce
 */
const GlobalSearchBar: React.FC = () => {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const anchorRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Debounce the search query (300ms)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(query)
    }, 300)

    return () => clearTimeout(timer)
  }, [query])

  // Fetch search results
  const { data, isLoading, isFetching } = useVideoSearch(
    debouncedQuery,
    isOpen && query.length >= 2
  )

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setQuery(value)
    setIsOpen(value.length >= 2)
  }, [])

  const handleFocus = useCallback(() => {
    if (query.length >= 2) {
      setIsOpen(true)
    }
  }, [query])

  const handleClickAway = useCallback(() => {
    setIsOpen(false)
  }, [])

  const handleVideoClick = useCallback(
    (videoId: string) => {
      navigate(`/video/${videoId}`)
      setQuery('')
      setIsOpen(false)
    },
    [navigate]
  )

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setIsOpen(false)
      inputRef.current?.blur()
    }
  }, [])

  const showLoading = isLoading || isFetching
  const results = data?.data || []
  const hasResults = results.length > 0

  return (
    <ClickAwayListener onClickAway={handleClickAway}>
      <Box ref={anchorRef} sx={{ position: 'relative', width: { xs: 200, sm: 300, md: 400 } }}>
        {/* Search Input */}
        <TextField
          inputRef={inputRef}
          size="small"
          fullWidth
          placeholder="Tìm kiếm video..."
          value={query}
          onChange={handleInputChange}
          onFocus={handleFocus}
          onKeyDown={handleKeyDown}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="action" />
                </InputAdornment>
              ),
              endAdornment: showLoading ? (
                <InputAdornment position="end">
                  <CircularProgress size={18} />
                </InputAdornment>
              ) : null,
            },
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 3,
              bgcolor: 'background.paper',
              '&:hover': {
                bgcolor: 'background.paper',
              },
              '&.Mui-focused': {
                bgcolor: 'background.paper',
              },
            },
          }}
        />

        {/* Search Results Dropdown */}
        <Popper
          open={isOpen && query.length >= 2}
          anchorEl={anchorRef.current}
          placement="bottom-start"
          transition
          style={{ width: anchorRef.current?.clientWidth, zIndex: 1300 }}
        >
          {({ TransitionProps }) => (
            <Fade {...TransitionProps} timeout={200}>
              <Paper
                elevation={8}
                sx={{
                  mt: 0.5,
                  maxHeight: 400,
                  overflow: 'auto',
                  borderRadius: 2,
                }}
              >
                {!hasResults && !showLoading && (
                  <Box sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                      Không tìm thấy kết quả cho "{query}"
                    </Typography>
                  </Box>
                )}

                {hasResults && (
                  <List disablePadding>
                    {results.map((video) => (
                      <ListItemButton
                        key={video.id}
                        onClick={() => handleVideoClick(video.id)}
                        sx={{
                          py: 1.5,
                          px: 2,
                          borderBottom: 1,
                          borderColor: 'divider',
                          '&:last-child': {
                            borderBottom: 0,
                          },
                          '&:hover': {
                            bgcolor: 'action.hover',
                          },
                        }}
                      >
                        <ListItemText
                          primary={
                            <Typography
                              variant="body2"
                              fontWeight={500}
                              sx={{
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                display: '-webkit-box',
                                WebkitLineClamp: 2,
                                WebkitBoxOrient: 'vertical',
                              }}
                            >
                              {video.title}
                            </Typography>
                          }
                          secondary={
                            video.has_transcript && (
                              <Stack direction="row" spacing={0.5} sx={{ mt: 0.5 }}>
                                <Chip
                                  label="CC"
                                  size="small"
                                  color="primary"
                                  variant="outlined"
                                  sx={{ height: 20, fontSize: '0.65rem' }}
                                />
                                {video.review_count > 0 && (
                                  <Chip
                                    label="Đã duyệt"
                                    size="small"
                                    color="success"
                                    variant="outlined"
                                    sx={{ height: 20, fontSize: '0.65rem' }}
                                  />
                                )}
                              </Stack>
                            )
                          }
                        />
                      </ListItemButton>
                    ))}
                  </List>
                )}

                {hasResults && data && data.pagination.total_items > 10 && (
                  <Box
                    sx={{
                      p: 1.5,
                      textAlign: 'center',
                      borderTop: 1,
                      borderColor: 'divider',
                      bgcolor: 'background.default',
                    }}
                  >
                    <Typography variant="caption" color="text.secondary">
                      Hiển thị 10/{data.pagination.total_items} kết quả
                    </Typography>
                  </Box>
                )}
              </Paper>
            </Fade>
          )}
        </Popper>
      </Box>
    </ClickAwayListener>
  )
}

export default GlobalSearchBar
