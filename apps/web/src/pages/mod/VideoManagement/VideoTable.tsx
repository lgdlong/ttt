import React from 'react'
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  IconButton,
  CircularProgress,
  Typography,
  Avatar,
  Chip,
  Tooltip,
} from '@mui/material'
import { Delete as DeleteIcon, YouTube as YouTubeIcon, Edit as EditIcon } from '@mui/icons-material'
import type { Video } from '~types/video'
import { formatDuration } from './utils'

interface VideoTableProps {
  videos: Video[]
  isLoading: boolean
  page: number
  pageSize: number
  total: number
  debouncedSearch: string
  onPageChange: (_: unknown, newPage: number) => void
  onRowsPerPageChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  onOpenTagDialog: (video: Video) => void
  onOpenDeleteDialog: (video: Video) => void
}

export const VideoTable: React.FC<VideoTableProps> = ({
  videos,
  isLoading,
  page,
  pageSize,
  total,
  debouncedSearch,
  onPageChange,
  onRowsPerPageChange,
  onOpenTagDialog,
  onOpenDeleteDialog,
}) => {
  return (
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
                <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={32} />
                </TableCell>
              </TableRow>
            ) : videos.length > 0 ? (
              videos.map((video) => (
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
                      <IconButton size="small" onClick={() => onOpenTagDialog(video)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Xóa">
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => onOpenDeleteDialog(video)}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
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
        count={total}
        page={page}
        onPageChange={onPageChange}
        rowsPerPage={pageSize}
        onRowsPerPageChange={onRowsPerPageChange}
        rowsPerPageOptions={[5, 10, 25, 50]}
        labelRowsPerPage="Số hàng:"
        labelDisplayedRows={({ from, to, count }) => `${from}–${to} / ${count}`}
      />
    </Paper>
  )
}

export default VideoTable
