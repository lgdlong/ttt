import React, { Suspense, useRef, useState, useEffect, useCallback } from 'react'
import { useParams, Link as RouterLink } from 'react-router-dom'
import {
  Box,
  Container,
  Typography,
  Link,
  Stack,
  Button,
  Skeleton,
  IconButton,
  Grid,
} from '@mui/material'
import BookmarkBorderIcon from '@mui/icons-material/BookmarkBorder'
import ThumbUpOffAltIcon from '@mui/icons-material/ThumbUpOffAlt'
import ShareIcon from '@mui/icons-material/Share'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import { useVideoDetail, useVideoTranscript } from '~/hooks'
import type { SegmentResponse, TranscriptParagraph } from '~/types/video'
import { VideoSummary } from '~/components/video/VideoSummary'
import { VideoChapters } from '~/components/video/VideoChapters'

/**
 * Group transcript segments into paragraphs (9 segments each)
 */
const groupIntoParagraphs = (segments: SegmentResponse[]): TranscriptParagraph[] => {
  const paragraphs: TranscriptParagraph[] = []
  const SEGMENTS_PER_PARAGRAPH = 9

  for (let i = 0; i < segments.length; i += SEGMENTS_PER_PARAGRAPH) {
    const paragraphSegments = segments.slice(i, i + SEGMENTS_PER_PARAGRAPH)
    paragraphs.push({
      id: `paragraph-${Math.floor(i / SEGMENTS_PER_PARAGRAPH)}`,
      startTime: paragraphSegments[0]?.start_time || 0,
      segments: paragraphSegments,
    })
  }

  return paragraphs
}

/**
 * Format milliseconds to MM:SS or HH:MM:SS
 */
const formatTime = (milliseconds: number): string => {
  const totalSeconds = Math.floor(milliseconds / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const secs = totalSeconds % 60

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`
}

/**
 * TranscriptView - Interactive transcript display
 */
interface TranscriptViewProps {
  transcript: SegmentResponse[]
  currentTime: number // in milliseconds
  onSeek: (time: number) => void
}

const TranscriptView: React.FC<TranscriptViewProps> = ({ transcript, currentTime, onSeek }) => {
  const paragraphs = groupIntoParagraphs(transcript)
  const activeSegmentRef = useRef<HTMLSpanElement>(null)

  // Auto-scroll to active segment
  useEffect(() => {
    if (activeSegmentRef.current) {
      activeSegmentRef.current.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      })
    }
  }, [currentTime])

  const isSegmentActive = (segment: SegmentResponse) => {
    return currentTime >= segment.start_time && currentTime < segment.end_time
  }

  return (
    <Box>
      {paragraphs.map((paragraph) => (
        <Box key={paragraph.id} sx={{ mb: 3 }}>
          {/* Timestamp */}
          <Typography
            variant="caption"
            color="primary.main"
            fontWeight={600}
            sx={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 0.5,
              cursor: 'pointer',
              mb: 1,
              '&:hover': {
                textDecoration: 'underline',
              },
            }}
            onClick={() => onSeek(paragraph.startTime)}
          >
            <PlayArrowIcon sx={{ fontSize: 14 }} />
            {formatTime(paragraph.startTime)}
          </Typography>

          {/* Paragraph text */}
          <Typography variant="body2" color="text.secondary" lineHeight={1.8}>
            {paragraph.segments.map((segment) => (
              <Box
                key={segment.id}
                component="span"
                ref={isSegmentActive(segment) ? activeSegmentRef : undefined}
                onClick={() => onSeek(segment.start_time)}
                sx={{
                  cursor: 'pointer',
                  px: 0.25,
                  py: 0.125,
                  borderRadius: 0.5,
                  transition: 'background-color 0.2s',
                  bgcolor: isSegmentActive(segment) ? 'primary.light' : 'transparent',
                  '&:hover': {
                    bgcolor: isSegmentActive(segment) ? 'primary.light' : 'action.hover',
                  },
                }}
              >
                {segment.text}{' '}
              </Box>
            ))}
          </Typography>
        </Box>
      ))}
    </Box>
  )
}

/**
 * VideoContent - Main video content with player and info
 */
const VideoContent: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const { data: video } = useVideoDetail(id!)
  const { data: transcript } = useVideoTranscript(id!)
  const [currentTime, setCurrentTime] = useState(0) // in milliseconds
  const playerRef = useRef<HTMLIFrameElement>(null)

  // TODO: Implement YouTube Player API integration for time sync
  // This requires loading the YouTube IFrame API and setting up event listeners
  const handleSeek = useCallback((time: number) => {
    // TODO: Implement seek functionality with YouTube Player API
    // For now, just update local state
    setCurrentTime(time)
    console.log('Seek to:', formatTime(time))
  }, [])

  // TODO: Handle like action
  const handleLike = () => {
    console.log('Like video:', video.id)
    // TODO: Call API to like video
  }

  // TODO: Handle save/bookmark action
  const handleSave = () => {
    console.log('Save video:', video.id)
    // TODO: Call API to save video
  }

  // TODO: Handle share action
  const handleShare = () => {
    console.log('Share video:', video.id)
    // TODO: Implement share functionality
  }

  return (
    <Grid container spacing={4}>
      {/* Left Column: Video Player & Info */}
      <Grid size={{ xs: 12, md: 7 }}>
        {/* Video Player */}
        <Box
          sx={{
            width: '100%',
            aspectRatio: '16/9',
            bgcolor: 'black',
            borderRadius: 3,
            mb: 2,
            overflow: 'hidden',
          }}
        >
          <iframe
            ref={playerRef}
            width="100%"
            height="100%"
            src={`https://www.youtube.com/embed/${video.youtube_id}?enablejsapi=1`}
            title={video.title}
            frameBorder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          />
        </Box>

        {/* Title */}
        <Typography variant="h5" fontWeight={700} fontFamily="Inter" gutterBottom>
          {video.title}
        </Typography>

        {/* Actions & Meta */}
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          justifyContent="space-between"
          alignItems={{ xs: 'flex-start', sm: 'center' }}
          spacing={2}
          sx={{ mb: 2 }}
        >
          <Typography variant="body2" color="text.secondary">
            {video.view_count.toLocaleString()} lượt xem • {video.published_at}
          </Typography>

          <Stack direction="row" spacing={1}>
            <Button startIcon={<ThumbUpOffAltIcon />} color="inherit" onClick={handleLike}>
              Thích
            </Button>
            <Button startIcon={<BookmarkBorderIcon />} color="inherit" onClick={handleSave}>
              Lưu
            </Button>
            <IconButton onClick={handleShare}>
              <ShareIcon />
            </IconButton>
          </Stack>
        </Stack>

        {/* Tags */}
        {video.tags && video.tags.length > 0 && (
          <Stack
            direction="row"
            spacing={1.5}
            mb={3}
            flexWrap="wrap"
            useFlexGap
            alignItems="center"
          >
            <Typography variant="body2" color="text.secondary">
              Tags:
            </Typography>
            {video.tags.map((tag) => (
              <React.Fragment key={tag.id}>
                <Link
                  component={RouterLink}
                  to={`/tag/${tag.id}`}
                  underline="always"
                  color="primary"
                  sx={{
                    fontWeight: 500,
                    '&:hover': {
                      color: 'primary.dark',
                    },
                  }}
                >
                  {tag.name}
                </Link>
                {/* {index < video.tags.length - 1 && (
                  <Typography component="span" color="text.secondary" sx={{ mx: -0.5 }}>
                    •
                  </Typography>
                )} */}
              </React.Fragment>
            ))}
          </Stack>
        )}

        {/* Summary & Chapters */}
        <VideoSummary summary={video.summary} />
        <VideoChapters chapters={video.chapters} onSeek={handleSeek} />
      </Grid>

      {/* Right Column: Transcript */}
      <Grid size={{ xs: 12, md: 5 }}>
        <Box
          sx={{
            height: { md: 'calc(100vh - 100px)' },
            position: { md: 'sticky' },
            top: { md: 80 },
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Typography variant="h6" fontWeight={600} mb={2}>
            Transcript
          </Typography>

          {/* Scrollable transcript area */}
          <Box
            sx={{
              flex: 1,
              overflowY: 'auto',
              pr: 1,
              // Custom scrollbar
              '&::-webkit-scrollbar': { width: '6px' },
              '&::-webkit-scrollbar-thumb': {
                backgroundColor: '#CBD5E1',
                borderRadius: '4px',
              },
            }}
          >
            {transcript.segments.length > 0 ? (
              <TranscriptView
                transcript={transcript.segments}
                currentTime={currentTime}
                onSeek={handleSeek}
              />
            ) : (
              <Typography variant="body2" color="text.secondary">
                Transcript chưa có sẵn cho video này.
              </Typography>
            )}
          </Box>
        </Box>
      </Grid>
    </Grid>
  )
}

/**
 * Loading skeleton for video detail page
 */
const VideoDetailSkeleton: React.FC = () => (
  <Grid container spacing={4}>
    <Grid size={{ xs: 12, md: 8 }}>
      <Skeleton variant="rectangular" height={400} sx={{ borderRadius: 3, mb: 2 }} />
      <Skeleton variant="text" height={40} />
      <Skeleton variant="text" width="40%" />
      <Stack direction="row" spacing={1} mt={2}>
        <Skeleton variant="rounded" width={80} height={32} />
        <Skeleton variant="rounded" width={80} height={32} />
      </Stack>
    </Grid>
    <Grid size={{ xs: 12, md: 4 }}>
      <Skeleton variant="text" height={32} width={120} />
      {Array.from({ length: 5 }).map((_, i) => (
        <Box key={i} sx={{ mb: 2 }}>
          <Skeleton variant="text" width={60} />
          <Skeleton variant="text" />
          <Skeleton variant="text" />
          <Skeleton variant="text" width="80%" />
        </Box>
      ))}
    </Grid>
  </Grid>
)

/**
 * VideoDetailPage Component
 * Split view with video player and interactive transcript
 */
const VideoDetailPage: React.FC = () => {
  return (
    <Container maxWidth="xl" sx={{ mt: 3, mb: 5 }}>
      <Suspense fallback={<VideoDetailSkeleton />}>
        <VideoContent />
      </Suspense>
    </Container>
  )
}

export default VideoDetailPage
