import React, { useState, useCallback, useEffect } from 'react'
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  IconButton,
  TextField,
  CircularProgress,
  InputAdornment,
  Tooltip,
  Avatar,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
} from '@mui/material'
import {
  Search as SearchIcon,
  YouTube as YouTubeIcon,
  Visibility as ViewIcon,
  CheckCircle as HasTranscriptIcon,
  Cancel as NoTranscriptIcon,
} from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import axiosInstance from '~/lib/axios'
import type { Video, ModVideoListResponse } from '~types/video'

// API functions
const fetchVideos = async (params: {
  page: number
  pageSize: number
  q?: string
}): Promise<ModVideoListResponse> => {
  const response = await axiosInstance.get('/videos', {
    params: {
      page: params.page,
      page_size: params.pageSize,
      q: params.q || undefined,
    },
  })
  return response.data
}

interface TranscriptSegment {
  start_time: number
  end_time: number
  text: string
}

interface TranscriptResponse {
  video_id: number
  youtube_id: string
  segments: TranscriptSegment[]
  total_segments: number
}

const fetchTranscript = async (videoId: number): Promise<TranscriptResponse | null> => {
  try {
    const response = await axiosInstance.get(`/videos/${videoId}/transcript`)
    return response.data
  } catch {
    return null
  }
}

// Format time from seconds to MM:SS or HH:MM:SS
const formatTime = (seconds: number): string => {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }
  return `${m}:${s.toString().padStart(2, '0')}`
}

/**
 * TranscriptManagement Component
 * View and manage video transcripts
 */
const TranscriptManagement: React.FC = () => {
  // State
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Dialog state
  const [openViewDialog, setOpenViewDialog] = useState(false)
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null)
  const [transcript, setTranscript] = useState<TranscriptResponse | null>(null)
  const [loadingTranscript, setLoadingTranscript] = useState(false)

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(0)
    }, 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // Query
  const { data, isLoading, error } = useQuery({
    queryKey: ['mod-videos-transcript', page, pageSize, debouncedSearch],
    queryFn: () => fetchVideos({ page: page + 1, pageSize, q: debouncedSearch }),
  })

  // Handlers
  const handleChangePage = useCallback((_: unknown, newPage: number) => {
    setPage(newPage)
  }, [])

  const handleChangeRowsPerPage = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setPageSize(parseInt(event.target.value, 10))
    setPage(0)
  }, [])

  const handleViewTranscript = useCallback(async (video: Video) => {
    setSelectedVideo(video)
    setOpenViewDialog(true)
    setLoadingTranscript(true)
    try {
      const data = await fetchTranscript(video.id)
      setTranscript(data)
    } finally {
      setLoadingTranscript(false)
    }
  }, [])

  const handleCloseViewDialog = useCallback(() => {
    setOpenViewDialog(false)
    setSelectedVideo(null)
    setTranscript(null)
  }, [])

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography color="error">Lỗi khi tải danh sách video</Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" fontWeight={600} gutterBottom>
          Quản lý Transcript
        </Typography>
        <Typography color="text.secondary">Xem và quản lý transcript của video</Typography>
      </Box>

      {/* Search */}
      <Paper elevation={0} sx={{ p: 2, mb: 3, border: '1px solid', borderColor: 'divider' }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Tìm kiếm video..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon color="action" />
                </InputAdornment>
              ),
            },
          }}
        />
      </Paper>

      {/* Videos Table */}
      <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Video</TableCell>
                <TableCell>Trạng thái Transcript</TableCell>
                <TableCell align="right">Thao tác</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={3} align="center" sx={{ py: 4 }}>
                    <CircularProgress size={32} />
                  </TableCell>
                </TableRow>
              ) : data?.videos && data.videos.length > 0 ? (
                data.videos.map((video) => (
                  <TableRow key={video.id} hover>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Avatar
                          variant="rounded"
                          src={video.thumbnail_url}
                          sx={{ width: 80, height: 45 }}
                        >
                          <YouTubeIcon />
                        </Avatar>
                        <Box>
                          <Typography
                            variant="body2"
                            fontWeight={500}
                            noWrap
                            sx={{ maxWidth: 400 }}
                          >
                            {video.title}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {video.youtube_id}
                          </Typography>
                        </Box>
                      </Box>
                    </TableCell>
                    <TableCell>
                      {video.has_transcript ? (
                        <Chip
                          icon={<HasTranscriptIcon sx={{ fontSize: 16 }} />}
                          label="Có transcript"
                          size="small"
                          color="success"
                          sx={{ borderRadius: 0 }}
                        />
                      ) : (
                        <Chip
                          icon={<NoTranscriptIcon sx={{ fontSize: 16 }} />}
                          label="Chưa có"
                          size="small"
                          color="default"
                          sx={{ borderRadius: 0 }}
                        />
                      )}
                    </TableCell>
                    <TableCell align="right">
                      <Tooltip title="Xem transcript">
                        <span>
                          <IconButton
                            size="small"
                            onClick={() => handleViewTranscript(video)}
                            disabled={!video.has_transcript}
                          >
                            <ViewIcon fontSize="small" />
                          </IconButton>
                        </span>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={3} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">
                      {debouncedSearch ? 'Không tìm thấy video phù hợp' : 'Chưa có video nào'}
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          component="div"
          count={data?.total || 0}
          page={page}
          onPageChange={handleChangePage}
          rowsPerPage={pageSize}
          onRowsPerPageChange={handleChangeRowsPerPage}
          rowsPerPageOptions={[5, 10, 25, 50]}
          labelRowsPerPage="Số hàng:"
          labelDisplayedRows={({ from, to, count }) => `${from}–${to} / ${count}`}
        />
      </Paper>

      {/* View Transcript Dialog */}
      <Dialog
        open={openViewDialog}
        onClose={handleCloseViewDialog}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: { maxHeight: '80vh' },
        }}
      >
        <DialogTitle>
          Transcript
          {selectedVideo && (
            <Typography variant="body2" color="text.secondary">
              {selectedVideo.title}
            </Typography>
          )}
        </DialogTitle>
        <DialogContent dividers>
          {loadingTranscript ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : transcript && transcript.segments.length > 0 ? (
            <Box>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Tổng số segment: {transcript.total_segments}
              </Typography>
              <Box sx={{ maxHeight: 400, overflow: 'auto' }}>
                {transcript.segments.map((segment, index) => (
                  <Box
                    key={index}
                    sx={{
                      display: 'flex',
                      gap: 2,
                      py: 1,
                      borderBottom: '1px solid',
                      borderColor: 'divider',
                      '&:last-child': { borderBottom: 'none' },
                    }}
                  >
                    <Typography
                      variant="caption"
                      sx={{
                        minWidth: 80,
                        color: 'primary.main',
                        fontFamily: 'monospace',
                      }}
                    >
                      {formatTime(segment.start_time)}
                    </Typography>
                    <Typography variant="body2">{segment.text}</Typography>
                  </Box>
                ))}
              </Box>
            </Box>
          ) : (
            <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
              Không có transcript cho video này
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseViewDialog}>Đóng</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default TranscriptManagement
