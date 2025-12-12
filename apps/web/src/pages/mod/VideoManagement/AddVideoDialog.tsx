import React, { useState, useCallback } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  InputAdornment,
  CircularProgress,
  Paper,
  Box,
  Typography,
  Avatar,
  Autocomplete,
  Chip,
} from '@mui/material'
import { YouTube as YouTubeIcon } from '@mui/icons-material'
import axiosInstance from '~/lib/axios'
import type { Video } from '~types/video'
import type { TagResponse } from '~types/tag'
import { extractYoutubeId, formatDuration } from './utils'

interface AddVideoDialogProps {
  open: boolean
  onClose: () => void
  onSave: (youtubeId: string, tagIds?: number[]) => void
  isSaving: boolean
  availableTags: TagResponse[]
}

export const AddVideoDialog: React.FC<AddVideoDialogProps> = ({
  open,
  onClose,
  onSave,
  isSaving,
  availableTags,
}) => {
  const [youtubeInput, setYoutubeInput] = useState('')
  const [selectedTags, setSelectedTags] = useState<TagResponse[]>([])
  const [youtubeError, setYoutubeError] = useState('')
  const [previewVideo, setPreviewVideo] = useState<Video | null>(null)
  const [isFetching, setIsFetching] = useState(false)

  const handleFetchPreview = useCallback(async () => {
    const youtubeId = extractYoutubeId(youtubeInput.trim())
    if (!youtubeId) {
      setYoutubeError('URL hoặc ID YouTube không hợp lệ')
      return
    }

    setYoutubeError('')
    setIsFetching(true)

    try {
      const response = await axiosInstance.get(`/mod/videos/preview/${youtubeId}`)
      setPreviewVideo(response.data)
    } catch (err: any) {
      setYoutubeError(err.response?.data?.error || 'Không thể lấy thông tin video')
    } finally {
      setIsFetching(false)
    }
  }, [youtubeInput])

  const handleSave = useCallback(() => {
    if (!previewVideo) {
      setYoutubeError('Vui lòng fetch preview trước')
      return
    }
    onSave(
      previewVideo.youtube_id,
      undefined // Tags managed separately via tag management
    )
  }, [previewVideo, selectedTags, onSave])

  const handleClose = useCallback(() => {
    setYoutubeInput('')
    setSelectedTags([])
    setYoutubeError('')
    setPreviewVideo(null)
    onClose()
  }, [onClose])

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
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
            options={availableTags}
            getOptionLabel={(option) => option.name}
            value={selectedTags}
            onChange={(_, newValue) => setSelectedTags(newValue)}
            renderInput={(params) => (
              <TextField {...params} label="Tags (tùy chọn)" placeholder="Chọn tags..." />
            )}
            renderTags={(value, getTagProps) =>
              value.map((option, index) => {
                const { key, ...tagProps } = getTagProps({ index })
                return (
                  <Chip
                    key={key}
                    {...tagProps}
                    label={option.name}
                    size="small"
                    sx={{ borderRadius: 0 }}
                  />
                )
              })
            }
          />
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Hủy</Button>
        <Button variant="contained" onClick={handleSave} disabled={!previewVideo || isSaving}>
          {isSaving ? 'Đang lưu...' : 'Save Video'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default AddVideoDialog
