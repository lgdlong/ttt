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
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Chip,
  InputAdornment,
  Tooltip,
  Autocomplete,
  Avatar,
  Stack,
} from '@mui/material'
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Search as SearchIcon,
  YouTube as YouTubeIcon,
  Edit as EditIcon,
} from '@mui/icons-material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axiosInstance from '~/lib/axios'
import type { TagResponse } from '~types/tag'
import type { Video, VideoListRequest, ModVideoListResponse } from '~types/video'

// API functions
const fetchVideos = async (
  params: VideoListRequest & { has_transcript?: string }
): Promise<ModVideoListResponse> => {
  const response = await axiosInstance.get('/mod/videos', {
    params: {
      page: params.page,
      page_size: params.page_size,
      q: params.q || undefined,
      tag_ids: params.tag_ids?.join(',') || undefined,
      has_transcript: params.has_transcript || undefined,
    },
  })
  return response.data
}

const fetchAllTags = async (): Promise<{ tags: TagResponse[] }> => {
  const response = await axiosInstance.get('/mod/tags', {
    params: { page_size: 1000 },
  })
  return response.data
}

const createVideo = async (youtubeId: string, tagIds?: number[]): Promise<Video> => {
  const response = await axiosInstance.post('/mod/videos', {
    youtube_id: youtubeId,
    tag_ids: tagIds,
  })
  return response.data
}

const deleteVideo = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/mod/videos/${id}`)
}

const addTagsToVideo = async (videoId: number, tagIds: number[]): Promise<Video> => {
  const response = await axiosInstance.post(`/mod/videos/${videoId}/tags`, {
    tag_ids: tagIds,
  })
  return response.data
}

const removeTagFromVideo = async (videoId: number, tagId: number): Promise<void> => {
  await axiosInstance.delete(`/mod/videos/${videoId}/tags/${tagId}`)
}

// Helper to extract YouTube ID from URL
const extractYoutubeId = (input: string): string | null => {
  // If it's already a valid ID (11 chars alphanumeric with - and _)
  if (/^[a-zA-Z0-9_-]{11}$/.test(input)) {
    return input
  }

  // Try to extract from various YouTube URL formats
  const patterns = [
    /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/|youtube\.com\/v\/)([a-zA-Z0-9_-]{11})/,
    /youtube\.com\/shorts\/([a-zA-Z0-9_-]{11})/,
  ]

  for (const pattern of patterns) {
    const match = input.match(pattern)
    if (match) return match[1]
  }

  return null
}

// Format duration from seconds to readable format
const formatDuration = (seconds?: number): string => {
  if (!seconds) return '—'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }
  return `${m}:${s.toString().padStart(2, '0')}`
}

/**
 * VideoManagement Component
 * Manage videos - add via YouTube URL, delete, manage tags
 */
const VideoManagement: React.FC = () => {
  const queryClient = useQueryClient()

  // State
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Dialog state
  const [openAddDialog, setOpenAddDialog] = useState(false)
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false)
  const [openTagDialog, setOpenTagDialog] = useState(false)
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null)

  // Add video form with preview
  const [youtubeInput, setYoutubeInput] = useState('')
  const [selectedTags, setSelectedTags] = useState<TagResponse[]>([])
  const [youtubeError, setYoutubeError] = useState('')
  const [previewVideo, setPreviewVideo] = useState<Video | null>(null)
  const [isFetching, setIsFetching] = useState(false)

  // Tag management
  const [videoTags, setVideoTags] = useState<TagResponse[]>([])
  const [tagToAdd, setTagToAdd] = useState<TagResponse | null>(null)

  // Filter by script status
  const [scriptFilter, setScriptFilter] = useState<'all' | 'with' | 'without'>('all')

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(0)
    }, 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // Queries
  const { data, isLoading, error } = useQuery({
    queryKey: ['mod-videos', page, pageSize, debouncedSearch, scriptFilter],
    queryFn: () => {
      // Convert scriptFilter to backend param
      let hasTranscriptParam: string | undefined
      if (scriptFilter === 'with') hasTranscriptParam = 'true'
      else if (scriptFilter === 'without') hasTranscriptParam = 'false'
      // if 'all', leave undefined (no filter)

      return fetchVideos({
        page: page + 1,
        page_size: pageSize,
        q: debouncedSearch,
        has_transcript: hasTranscriptParam,
      })
    },
  })

  const { data: tagsData } = useQuery({
    queryKey: ['mod-all-tags'],
    queryFn: fetchAllTags,
  })

  // Mutations
  const createMutation = useMutation({
    mutationFn: ({ youtubeId, tagIds }: { youtubeId: string; tagIds?: number[] }) =>
      createVideo(youtubeId, tagIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      handleCloseAddDialog()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteVideo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      handleCloseDeleteDialog()
    },
  })

  const addTagMutation = useMutation({
    mutationFn: ({ videoId, tagIds }: { videoId: number; tagIds: number[] }) =>
      addTagsToVideo(videoId, tagIds),
    onSuccess: (updatedVideo) => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      setVideoTags(updatedVideo.tags || [])
    },
  })

  const removeTagMutation = useMutation({
    mutationFn: ({ videoId, tagId }: { videoId: number; tagId: number }) =>
      removeTagFromVideo(videoId, tagId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-videos'] })
      if (selectedVideo) {
        setVideoTags((prev) => prev.filter((t) => t.id !== removeTagMutation.variables?.tagId))
      }
    },
  })

  // Handlers
  const handleChangePage = useCallback((_: unknown, newPage: number) => {
    setPage(newPage)
  }, [])

  const handleChangeRowsPerPage = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setPageSize(parseInt(event.target.value, 10))
    setPage(0)
  }, [])

  const handleOpenAddDialog = useCallback(() => {
    setYoutubeInput('')
    setSelectedTags([])
    setYoutubeError('')
    setPreviewVideo(null)
    setOpenAddDialog(true)
  }, [])

  const handleCloseAddDialog = useCallback(() => {
    setOpenAddDialog(false)
    setYoutubeInput('')
    setSelectedTags([])
    setYoutubeError('')
    setPreviewVideo(null)
  }, [])

  // Fetch YouTube metadata preview
  const handleFetchPreview = useCallback(async () => {
    const youtubeId = extractYoutubeId(youtubeInput.trim())
    if (!youtubeId) {
      setYoutubeError('URL hoặc ID YouTube không hợp lệ')
      return
    }

    setYoutubeError('')
    setIsFetching(true)

    try {
      // Call backend to fetch metadata without saving
      const response = await axiosInstance.get(`/mod/videos/preview/${youtubeId}`)
      setPreviewVideo(response.data)
    } catch (err: any) {
      setYoutubeError(err.response?.data?.error || 'Không thể lấy thông tin video')
    } finally {
      setIsFetching(false)
    }
  }, [youtubeInput])

  const handleOpenDeleteDialog = useCallback((video: Video) => {
    setSelectedVideo(video)
    setOpenDeleteDialog(true)
  }, [])

  const handleCloseDeleteDialog = useCallback(() => {
    setOpenDeleteDialog(false)
    setSelectedVideo(null)
  }, [])

  const handleOpenTagDialog = useCallback((video: Video) => {
    setSelectedVideo(video)
    setVideoTags(video.tags || [])
    setTagToAdd(null)
    setOpenTagDialog(true)
  }, [])

  const handleCloseTagDialog = useCallback(() => {
    setOpenTagDialog(false)
    setSelectedVideo(null)
    setVideoTags([])
    setTagToAdd(null)
  }, [])

  const handleSaveVideo = useCallback(() => {
    if (!previewVideo) {
      setYoutubeError('Vui lòng fetch preview trước')
      return
    }
    createMutation.mutate({
      youtubeId: previewVideo.youtube_id,
      tagIds: selectedTags.map((t) => t.id),
    })
  }, [previewVideo, selectedTags, createMutation])

  const handleDelete = useCallback(() => {
    if (!selectedVideo) return
    deleteMutation.mutate(selectedVideo.id)
  }, [selectedVideo, deleteMutation])

  const handleAddTagToVideo = useCallback(() => {
    if (!selectedVideo || !tagToAdd) return
    addTagMutation.mutate({
      videoId: selectedVideo.id,
      tagIds: [tagToAdd.id],
    })
    setTagToAdd(null)
  }, [selectedVideo, tagToAdd, addTagMutation])

  const handleRemoveTagFromVideo = useCallback(
    (tagId: number) => {
      if (!selectedVideo) return
      removeTagMutation.mutate({
        videoId: selectedVideo.id,
        tagId,
      })
    },
    [selectedVideo, removeTagMutation]
  )

  // Available tags (exclude already added)
  const availableTags = (tagsData?.tags || []).filter(
    (tag) => !videoTags.some((vt) => vt.id === tag.id)
  )

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography color="error">Lỗi khi tải danh sách video</Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            Quản lý Videos
          </Typography>
          <Typography color="text.secondary">Thêm video từ YouTube và quản lý tags</Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenAddDialog}
          sx={{ borderRadius: 0 }}
        >
          Thêm Video
        </Button>
      </Box>

      {/* Search & Filters */}
      <Paper elevation={0} sx={{ p: 2, mb: 3, border: '1px solid', borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <TextField
            size="small"
            placeholder="Tìm kiếm video theo tiêu đề..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            sx={{ flex: 1, minWidth: 300 }}
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
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Chip
              label="Tất cả"
              onClick={() => setScriptFilter('all')}
              color={scriptFilter === 'all' ? 'primary' : 'default'}
              sx={{ borderRadius: 0 }}
            />
            <Chip
              label="Có script"
              onClick={() => setScriptFilter('with')}
              color={scriptFilter === 'with' ? 'success' : 'default'}
              sx={{ borderRadius: 0 }}
            />
            <Chip
              label="Chưa có script"
              onClick={() => setScriptFilter('without')}
              color={scriptFilter === 'without' ? 'warning' : 'default'}
              sx={{ borderRadius: 0 }}
            />
          </Box>
        </Box>
      </Paper>

      {/* Videos Table */}
      <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell width="40px">Thumbnail</TableCell>
                <TableCell>Title</TableCell>
                <TableCell width="100px">Duration</TableCell>
                <TableCell width="150px">Script Status</TableCell>
                <TableCell width="120px" align="right">
                  Actions
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
                    <CircularProgress size={32} />
                  </TableCell>
                </TableRow>
              ) : data?.videos && data.videos.length > 0 ? (
                data.videos.map((video) => (
                  <TableRow key={video.id} hover>
                    <TableCell>
                      <Avatar
                        variant="rounded"
                        src={video.thumbnail_url}
                        sx={{ width: 60, height: 34 }}
                      >
                        <YouTubeIcon fontSize="small" />
                      </Avatar>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" fontWeight={500} noWrap sx={{ maxWidth: 400 }}>
                        {video.title}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {video.youtube_id}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">{formatDuration(video.duration)}</Typography>
                    </TableCell>
                    <TableCell>
                      {video.has_transcript ? (
                        <Chip
                          label="Có script"
                          size="small"
                          color="success"
                          sx={{ borderRadius: 0 }}
                        />
                      ) : (
                        <Chip
                          label="Chưa có script"
                          size="small"
                          color="warning"
                          sx={{ borderRadius: 0 }}
                        />
                      )}
                    </TableCell>
                    <TableCell align="right">
                      <Tooltip title="Quản lý tags">
                        <IconButton size="small" onClick={() => handleOpenTagDialog(video)}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Xóa">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleOpenDeleteDialog(video)}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
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

      {/* Add Video Dialog with Preview */}
      <Dialog open={openAddDialog} onClose={handleCloseAddDialog} maxWidth="md" fullWidth>
        <DialogTitle>Thêm Video từ YouTube - One-Click Import</DialogTitle>
        <DialogContent>
          {/* Step 1: Paste URL & Fetch */}
          <Box sx={{ mb: 3 }}>
            <TextField
              autoFocus
              fullWidth
              label="Paste YouTube URL"
              placeholder="https://youtube.com/watch?v=xxx hoặc video ID"
              value={youtubeInput}
              onChange={(e) => {
                setYoutubeInput(e.target.value)
                setYoutubeError('')
                setPreviewVideo(null)
              }}
              error={!!youtubeError}
              helperText={youtubeError}
              margin="normal"
              slotProps={{
                input: {
                  startAdornment: (
                    <InputAdornment position="start">
                      <YouTubeIcon color="error" />
                    </InputAdornment>
                  ),
                  endAdornment: (
                    <InputAdornment position="end">
                      <Button
                        variant="contained"
                        onClick={handleFetchPreview}
                        disabled={!youtubeInput.trim() || isFetching}
                        sx={{ borderRadius: 0 }}
                      >
                        {isFetching ? <CircularProgress size={20} /> : 'Fetch'}
                      </Button>
                    </InputAdornment>
                  ),
                },
              }}
            />
          </Box>

          {/* Step 2: Preview */}
          {previewVideo && (
            <Paper elevation={0} sx={{ p: 2, border: '1px solid', borderColor: 'divider', mb: 3 }}>
              <Typography variant="subtitle2" gutterBottom color="success.main">
                ✓ Preview Video
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                <Avatar
                  variant="rounded"
                  src={previewVideo.thumbnail_url}
                  sx={{ width: 120, height: 68 }}
                >
                  <YouTubeIcon />
                </Avatar>
                <Box sx={{ flex: 1 }}>
                  <Typography variant="body1" fontWeight={600} gutterBottom>
                    {previewVideo.title}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block">
                    Duration: {formatDuration(previewVideo.duration)}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block">
                    Published:{' '}
                    {previewVideo.published_at
                      ? new Date(previewVideo.published_at).toLocaleDateString('vi-VN')
                      : 'N/A'}
                  </Typography>
                </Box>
              </Box>
              {previewVideo.description && (
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{
                    maxHeight: 100,
                    overflow: 'auto',
                    whiteSpace: 'pre-wrap',
                  }}
                >
                  {previewVideo.description.slice(0, 200)}
                  {previewVideo.description.length > 200 && '...'}
                </Typography>
              )}
            </Paper>
          )}

          {/* Step 3: Select Tags */}
          {previewVideo && (
            <Autocomplete
              multiple
              options={tagsData?.tags || []}
              getOptionLabel={(option) => option.name}
              value={selectedTags}
              onChange={(_, newValue) => setSelectedTags(newValue)}
              renderInput={(params) => (
                <TextField {...params} label="Tags (tùy chọn)" placeholder="Chọn tags..." />
              )}
              renderTags={(value, getTagProps) =>
                value.map((option, index) => (
                  <Chip
                    {...getTagProps({ index })}
                    key={option.id}
                    label={option.name}
                    size="small"
                    sx={{ borderRadius: 0 }}
                  />
                ))
              }
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseAddDialog}>Hủy</Button>
          <Button
            variant="contained"
            onClick={handleSaveVideo}
            disabled={!previewVideo || createMutation.isPending}
          >
            {createMutation.isPending ? 'Đang lưu...' : 'Save Video'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Dialog */}
      <Dialog open={openDeleteDialog} onClose={handleCloseDeleteDialog}>
        <DialogTitle>Xác nhận xóa</DialogTitle>
        <DialogContent>
          <Typography>
            Bạn có chắc muốn xóa video <strong>{selectedVideo?.title}</strong>?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDeleteDialog}>Hủy</Button>
          <Button
            variant="contained"
            color="error"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Đang xóa...' : 'Xóa'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Tag Management Dialog */}
      <Dialog open={openTagDialog} onClose={handleCloseTagDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Quản lý Tags cho Video</DialogTitle>
        <DialogContent>
          {selectedVideo && (
            <Box sx={{ mb: 2 }}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {selectedVideo.title}
              </Typography>
            </Box>
          )}

          {/* Current tags */}
          <Typography variant="subtitle2" gutterBottom>
            Tags hiện tại:
          </Typography>
          <Box sx={{ mb: 2, minHeight: 40 }}>
            {videoTags.length > 0 ? (
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                {videoTags.map((tag) => (
                  <Chip
                    key={tag.id}
                    label={tag.name}
                    size="small"
                    onDelete={() => handleRemoveTagFromVideo(tag.id)}
                    sx={{ borderRadius: 0, mb: 1 }}
                  />
                ))}
              </Stack>
            ) : (
              <Typography variant="body2" color="text.secondary">
                Chưa có tag nào
              </Typography>
            )}
          </Box>

          {/* Add tag */}
          <Typography variant="subtitle2" gutterBottom>
            Thêm tag:
          </Typography>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Autocomplete
              fullWidth
              options={availableTags}
              getOptionLabel={(option) => option.name}
              value={tagToAdd}
              onChange={(_, newValue) => setTagToAdd(newValue)}
              renderInput={(params) => (
                <TextField {...params} size="small" placeholder="Chọn tag..." />
              )}
            />
            <Button
              variant="contained"
              onClick={handleAddTagToVideo}
              disabled={!tagToAdd || addTagMutation.isPending}
              sx={{ borderRadius: 0 }}
            >
              Thêm
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseTagDialog}>Đóng</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default VideoManagement
