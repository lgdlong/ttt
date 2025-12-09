import React from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  IconButton,
  Tooltip,
  Avatar,
  Chip,
  Box,
  Typography,
} from '@mui/material'
import {
  YouTube as YouTubeIcon,
  Visibility as ViewIcon,
  CheckCircle as HasTranscriptIcon,
  Cancel as NoTranscriptIcon,
  Edit as EditIcon,
} from '@mui/icons-material'
import type { Video } from '~/types/video'

interface VideoListTableProps {
  videos: Video[]
  isLoading: boolean
  searchQuery: string
  onViewTranscript: (video: Video) => void
  onEditTranscript: (videoId: number) => void
}

export const VideoListTable: React.FC<VideoListTableProps> = ({
  videos,
  isLoading,
  searchQuery,
  onViewTranscript,
  onEditTranscript,
}) => {
  if (isLoading) {
    return (
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
            <TableRow>
              <TableCell colSpan={3} align="center" sx={{ py: 4 }}>
                <CircularProgress size={32} />
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  if (!videos || videos.length === 0) {
    return (
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
            <TableRow>
              <TableCell colSpan={3} align="center" sx={{ py: 4 }}>
                <Typography color="text.secondary">
                  {searchQuery ? 'Không tìm thấy video phù hợp' : 'Chưa có video nào'}
                </Typography>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    )
  }

  return (
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
          {videos.map((video) => (
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
                    <Typography variant="body2" fontWeight={500} noWrap sx={{ maxWidth: 400 }}>
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
                <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                  <Tooltip title="Xem nhanh (Dialog)">
                    <span>
                      <IconButton
                        size="small"
                        onClick={() => onViewTranscript(video)}
                        disabled={!video.has_transcript}
                      >
                        <ViewIcon fontSize="small" />
                      </IconButton>
                    </span>
                  </Tooltip>
                  <Tooltip title="Mở Trình Sửa Transcript">
                    <IconButton
                      size="small"
                      color="primary"
                      onClick={() => onEditTranscript(video.id)}
                      disabled={!video.has_transcript}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Box>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  )
}
