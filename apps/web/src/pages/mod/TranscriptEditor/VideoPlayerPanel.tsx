import { Box, Paper, IconButton, Tooltip, Typography } from '@mui/material'
import { Replay } from '@mui/icons-material'
import YouTube, { YouTubePlayer } from 'react-youtube'

interface VideoPlayerPanelProps {
  videoId: string
  videoTitle: string
  onPlayerReady: (player: YouTubePlayer) => void
  onReplay: () => void
}

export const VideoPlayerPanel: React.FC<VideoPlayerPanelProps> = ({
  videoId,
  videoTitle,
  onPlayerReady,
  onReplay,
}) => {
  const opts = {
    height: '100%',
    width: '100%',
    playerVars: {
      autoplay: 0,
      controls: 1, // Bật thanh điều khiển YouTube
      disablekb: 0, // Cho phép keyboard shortcuts của YouTube
      fs: 1, // Cho phép fullscreen
      modestbranding: 1,
      rel: 0,
      showinfo: 0,
      iv_load_policy: 3,
      origin: window.location.origin,
    },
  }

  return (
    <>
      <Typography variant="h6" sx={{ fontWeight: 600, lineHeight: 1.3 }}>
        {videoTitle}
      </Typography>

      <Paper elevation={3} sx={{ aspectRatio: '16/9', overflow: 'hidden' }}>
        <YouTube
          videoId={videoId}
          opts={opts}
          onReady={(event) => onPlayerReady(event.target)}
          style={{ height: '100%', width: '100%' }}
        />
      </Paper>

      <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', justifyContent: 'flex-end' }}>
        <Tooltip title="Replay hiện tại (Ctrl+R)">
          <IconButton onClick={onReplay}>
            <Replay />
          </IconButton>
        </Tooltip>
      </Box>
    </>
  )
}
