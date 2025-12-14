import React from 'react'
import { Card, CardMedia, CardContent, Typography, Box } from '@mui/material'
import ClosedCaptionIcon from '@mui/icons-material/ClosedCaption'
import { useNavigate } from 'react-router-dom'
import type { VideoCardResponse } from '~/types/video'

interface VideoCardProps {
  video: VideoCardResponse
}

/**
 * VideoCard Component
 * Displays video thumbnail, title, verified badge, and metadata
 * Following UI_Spec.md specifications
 */
const VideoCard: React.FC<VideoCardProps> = ({ video }) => {
  const navigate = useNavigate()

  const handleClick = () => {
    navigate(`/video/${video.id}`)
  }

  // Format view count (e.g., 1500000 -> 1.5M)
  const formatViews = (views: number): string => {
    if (views >= 1000000) {
      return `${(views / 1000000).toFixed(1)}M`
    }
    if (views >= 1000) {
      return `${(views / 1000).toFixed(1)}K`
    }
    return views.toString()
  }

  // Format duration from seconds to MM:SS or HH:MM:SS
  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <Card
      onClick={handleClick}
      sx={{
        height: '100%',
        borderRadius: 2,
        cursor: 'pointer',
        transition: 'transform 0.2s ease-in-out',
        '&:hover': {
          transform: 'translateY(-4px)',
        },
      }}
    >
      {/* Thumbnail with duration badge */}
      <Box sx={{ position: 'relative' }}>
        <CardMedia
          component="img"
          height="180"
          image={video.thumbnail_url || 'https://placehold.co/600x400?text=No+Thumbnail'}
          alt={video.title}
          sx={{
            aspectRatio: '16/9',
            objectFit: 'cover',
          }}
        />
        <Box
          sx={{
            position: 'absolute',
            bottom: 8,
            right: 8,
            bgcolor: 'rgba(0,0,0,0.8)',
            color: 'white',
            fontSize: 12,
            px: 0.75,
            py: 0.25,
            borderRadius: 1,
            fontWeight: 500,
          }}
        >
          {formatDuration(video.duration)}
        </Box>
        {/* CC Badge if has transcript */}
        {video.has_transcript && (
          <Box
            sx={{
              position: 'absolute',
              bottom: 8,
              left: 8,
              bgcolor: 'rgba(0,0,0,0.8)',
              color: 'white',
              display: 'flex',
              alignItems: 'center',
              px: 0.5,
              py: 0.25,
              borderRadius: 1,
            }}
          >
            <ClosedCaptionIcon sx={{ fontSize: 14 }} />
          </Box>
        )}
        {/* Verified Badge if reviewed */}
        {video.review_count > 0 && (
          <Box
            sx={{
              position: 'absolute',
              top: 8,
              left: 8,
              bgcolor: 'success.main',
              color: 'white',
              fontSize: 11,
              fontWeight: 600,
              px: 1,
              py: 0.5,
              borderRadius: 1,
              zIndex: 2,
            }}
          >
            ĐÃ DUYỆT
          </Box>
        )}
      </Box>

      {/* Content */}
      <CardContent sx={{ pb: '16px !important' }}>
        <Typography
          variant="subtitle1"
          fontWeight={600}
          lineHeight={1.3}
          mb={1}
          sx={{
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
          }}
        >
          {video.title}
        </Typography>

        {/* Metadata */}
        <Typography variant="caption" color="text.disabled">
          {video.published_at} • {formatViews(video.view_count)} views
        </Typography>
      </CardContent>
    </Card>
  )
}

export default VideoCard
