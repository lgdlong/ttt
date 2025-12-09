import React from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  CircularProgress,
  Box,
  Typography,
} from '@mui/material'
import type { Video } from '~/types/video'

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

interface TranscriptViewDialogProps {
  open: boolean
  video: Video | null
  transcript: TranscriptResponse | null
  loading: boolean
  onClose: () => void
}

const formatTime = (seconds: number): string => {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }
  return `${m}:${s.toString().padStart(2, '0')}`
}

export const TranscriptViewDialog: React.FC<TranscriptViewDialogProps> = ({
  open,
  video,
  transcript,
  loading,
  onClose,
}) => {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { maxHeight: '80vh' },
      }}
    >
      <DialogTitle>
        Transcript
        {video && (
          <Typography variant="body2" color="text.secondary">
            {video.title}
          </Typography>
        )}
      </DialogTitle>
      <DialogContent dividers>
        {loading ? (
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
        <Button onClick={onClose}>Đóng</Button>
      </DialogActions>
    </Dialog>
  )
}

export type { TranscriptSegment, TranscriptResponse }
