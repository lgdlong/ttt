import React, { useState, useEffect } from 'react'
import { Box, Typography, Paper, TablePagination, TextField, InputAdornment } from '@mui/material'
import { Search as SearchIcon } from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { VideoListTable } from './VideoListTable'
import { TranscriptViewDialog } from './TranscriptViewDialog'
import { useTranscriptDialog } from './useTranscriptDialog'

import { fetchModVideos } from '~/api/modApi'

const TranscriptManagement: React.FC = () => {
  const navigate = useNavigate()
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  const {
    openViewDialog,
    selectedVideo,
    transcript,
    loadingTranscript,
    handleViewTranscript,
    handleCloseViewDialog,
  } = useTranscriptDialog()

  // Hàm chuyển sang trang Editor
  const handleEditTranscript = (videoId: string) => {
    navigate(`/mod/videos/${videoId}/transcript`)
  }

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(0)
    }, 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  const { data, isLoading, error } = useQuery({
    queryKey: ['mod-videos-transcript', page, pageSize, debouncedSearch],
    queryFn: () => fetchModVideos({ page: page + 1, page_size: pageSize, q: debouncedSearch }),
  })

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

      <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
        <VideoListTable
          videos={data?.videos || []}
          isLoading={isLoading}
          searchQuery={debouncedSearch}
          onViewTranscript={handleViewTranscript}
          onEditTranscript={handleEditTranscript}
        />
        <TablePagination
          component="div"
          count={data?.total || 0}
          page={page}
          onPageChange={(_, newPage) => setPage(newPage)}
          rowsPerPage={pageSize}
          onRowsPerPageChange={(e) => {
            setPageSize(parseInt(e.target.value, 10))
            setPage(0)
          }}
          rowsPerPageOptions={[5, 10, 25, 50]}
          labelRowsPerPage="Số hàng:"
          labelDisplayedRows={({ from, to, count }) => `${from}–${to} / ${count}`}
        />
      </Paper>

      <TranscriptViewDialog
        open={openViewDialog}
        video={selectedVideo}
        transcript={transcript}
        loading={loadingTranscript}
        onClose={handleCloseViewDialog}
      />
    </Box>
  )
}

export default TranscriptManagement
