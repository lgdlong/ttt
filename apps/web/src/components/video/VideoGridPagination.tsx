import React from 'react'
import { Box, Pagination } from '@mui/material'
import { useVideos } from '~/hooks/useVideos'
import type { VideoSort } from '~/types/video'

interface VideoGridPaginationProps {
  page: number
  limit?: number // Added limit prop for consistency
  onPageChange: (page: number) => void
  selectedTagId: string | null
  sort: VideoSort
}

const VideoGridPagination: React.FC<VideoGridPaginationProps> = ({
  page,
  limit = 12, // Default to 12 to match VideoGrid
  onPageChange,
  selectedTagId,
  sort,
}) => {
  const { data } = useVideos({
    page,
    limit,
    sort,
    tag_id: selectedTagId ?? undefined,
    has_transcript: true,
  })

  if (!data || data.pagination.total_pages <= 1) {
    return null
  }

  return (
    <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
      <Pagination
        count={data.pagination.total_pages}
        page={page}
        onChange={(_, value) => onPageChange(value)}
        color="primary"
        size="large"
        showFirstButton
        showLastButton
        sx={{ '& .MuiPaginationItem-root': { borderRadius: 0 } }}
      />
    </Box>
  )
}

export default VideoGridPagination
