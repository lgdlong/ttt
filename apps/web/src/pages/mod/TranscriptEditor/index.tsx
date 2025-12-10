import React, { useEffect, useRef, useState } from 'react'
import { Box, CircularProgress, Alert, Button, Snackbar } from '@mui/material'
import { CheckCircle, Save, Add } from '@mui/icons-material'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import type { YouTubePlayer } from 'react-youtube'
import axiosInstance from '~/lib/axios'
import type { Video, SegmentResponse } from '~/types/video'
import { VideoPlayerPanel } from './VideoPlayerPanel'
import { useVideoSync } from './useVideoSync'
import { useTranscriptEditor } from './useTranscriptEditor'
import { VirtualTranscriptList } from './components/VirtualTranscriptList'
import { useSubmitReview } from './hooks/useReviewMutation'
import { AddSegmentDialog } from './components/AddSegmentDialog'

export const TranscriptEditor: React.FC = () => {
  // State: kiểm soát scroll khi điều hướng
  const [shouldScrollToActive, setShouldScrollToActive] = useState(true)
  const { videoId } = useParams<{ videoId: string }>()
  const playerRef = useRef<YouTubePlayer | null>(null)
  const [snackbarOpen, setSnackbarOpen] = useState(false)
  const [snackbarMessage, setSnackbarMessage] = useState('')
  const [addSegmentDialogOpen, setAddSegmentDialogOpen] = useState(false)

  // Fetch video metadata only (transcript fetched by VirtualTranscriptList)
  const {
    data: videoData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['video', videoId],
    queryFn: async () => {
      if (!videoId) throw new Error('Video ID không hợp lệ')
      const response = await axiosInstance.get<Video>(`/videos/${videoId}`)
      return response.data
    },
    enabled: !!videoId,
  })

  // Review submission mutation
  const { mutate: submitReview, isPending: isSubmittingReview } = useSubmitReview({
    videoId: videoId || '',
  })

  // Fetch segments for sync hook (lightweight query)
  const { data: transcriptData } = useQuery({
    queryKey: ['transcript', videoId],
    queryFn: async () => {
      if (!videoId) throw new Error('Invalid video ID')
      const response = await axiosInstance.get<{ segments: SegmentResponse[] }>(
        `/videos/${videoId}/transcript`
      )
      return response.data
    },
    enabled: !!videoId,
  })

  // Video sync hook - simplified
  const { activeIndex, setActiveIndex } = useVideoSync({
    playerRef,
    segments: transcriptData?.segments || [],
    playing: false, // Will be controlled by editor hook
  })

  // Editor logic hook - simplified
  const { playing, setPlaying, handleKeyDown } = useTranscriptEditor({
    playerRef,
    segments: transcriptData?.segments || [],
    setActiveIndex,
  })

  // Control YouTube player based on playing state
  useEffect(() => {
    if (!playerRef.current) return

    try {
      if (playing) {
        playerRef.current.playVideo()
      } else {
        playerRef.current.pauseVideo()
      }
    } catch (error) {
      console.error('Error controlling player:', error)
    }
  }, [playing])

  if (isLoading) {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '80vh',
        }}
      >
        <CircularProgress />
      </Box>
    )
  }

  if (error || !videoData || !videoId) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">{error?.toString() || 'Không có dữ liệu'}</Alert>
      </Box>
    )
  }

  return (
    <Box sx={{ display: 'flex', height: 'calc(100vh - 64px)', overflow: 'hidden' }}>
      {/* Left Column: Sticky Video */}
      <Box
        sx={{
          width: '35%',
          minWidth: 400,
          position: 'sticky',
          top: 0,
          height: '100vh',
          p: 3,
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
          borderRight: '1px solid',
          borderColor: 'divider',
          backgroundColor: 'background.default',
        }}
      >
        <VideoPlayerPanel
          videoId={videoData.youtube_id}
          videoTitle={videoData.title}
          onPlayerReady={(player) => {
            playerRef.current = player
          }}
          onReplay={() => {
            if (playerRef.current && transcriptData?.segments[activeIndex]) {
              playerRef.current.seekTo(transcriptData.segments[activeIndex].start_time / 1000, true)
              setPlaying(true)
            }
          }}
        />

        {/* Add Segment Button */}
        <Button
          variant="outlined"
          color="primary"
          size="large"
          startIcon={<Add />}
          onClick={() => setAddSegmentDialogOpen(true)}
          sx={{
            py: 1.5,
            fontSize: '16px',
            fontWeight: 600,
            textTransform: 'none',
          }}
        >
          Thêm Segment
        </Button>

        {/* Review Action Button */}
        <Button
          variant="contained"
          color="success"
          size="large"
          startIcon={
            isSubmittingReview ? <CircularProgress size={20} color="inherit" /> : <CheckCircle />
          }
          disabled={isSubmittingReview}
          onClick={() => {
            submitReview(
              {},
              {
                onSuccess: (data) => {
                  setSnackbarMessage(data.message || 'Review submitted successfully!')
                  setSnackbarOpen(true)
                },
                onError: (error: unknown) => {
                  const err = error as { response?: { data?: { message?: string } } }
                  const errorMsg = err.response?.data?.message || 'Failed to submit review'
                  setSnackbarMessage(errorMsg)
                  setSnackbarOpen(true)
                },
              }
            )
          }}
          sx={{
            py: 1.5,
            fontSize: '16px',
            fontWeight: 600,
            textTransform: 'none',
          }}
        >
          {isSubmittingReview ? 'Đang xác nhận...' : 'Xác nhận đã duyệt'}
        </Button>

        {/* Auto-save status indicator */}
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            gap: 1,
            p: 1.5,
            borderRadius: 1,
            backgroundColor: 'success.light',
            color: 'success.contrastText',
            fontSize: '13px',
          }}
        >
          <Save sx={{ fontSize: 16 }} />
          <span>Tự động lưu: Mọi thay đổi được lưu ngay lập tức</span>
        </Box>

        {/* Keyboard shortcuts info */}
        <Box
          sx={{
            p: 2,
            borderRadius: 1,
            backgroundColor: 'info.light',
            color: 'info.contrastText',
          }}
        >
          <Box sx={{ fontSize: '12px', fontWeight: 600, mb: 1 }}>Phím tắt thao tác:</Box>
          <Box sx={{ fontSize: '11px', lineHeight: 1.6 }}>
            <div>• Enter: Chuyển sang đoạn tiếp theo + Phát</div>
            <div>• Shift+Enter: Quay lại đoạn trước</div>
            <div>• Ctrl+Space: Phát/Tạm dừng video</div>
            <div>• Ctrl+R: Phát lại đoạn hiện tại</div>
          </Box>
        </Box>
      </Box>

      {/* Right Column: Virtualized Transcript List */}
      <Box
        sx={{
          flex: 1,
          height: '100vh',
          backgroundColor: 'background.default',
        }}
      >
        <VirtualTranscriptList
          videoId={videoId}
          playerRef={playerRef}
          activeIndex={activeIndex}
          shouldScrollToActive={shouldScrollToActive}
          onActiveIndexChange={(idx) => {
            setActiveIndex(idx)
            setShouldScrollToActive(false) // Khi click chuột chọn row, không scroll
          }}
          onEditStart={() => setPlaying(false)}
          onKeyDown={(e, idx) => {
            handleKeyDown(e, idx)
            // Nếu là điều hướng bằng phím tắt thì scroll
            if (['Enter', 'ArrowDown', 'ArrowUp'].includes(e.key)) {
              setShouldScrollToActive(true)
            }
          }}
        />
      </Box>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={() => setSnackbarOpen(false)}
        message={snackbarMessage}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      />

      {/* Add Segment Dialog */}
      {videoId && (
        <AddSegmentDialog
          open={addSegmentDialogOpen}
          onClose={() => setAddSegmentDialogOpen(false)}
          videoId={videoId}
        />
      )}
    </Box>
  )
}

export default TranscriptEditor
