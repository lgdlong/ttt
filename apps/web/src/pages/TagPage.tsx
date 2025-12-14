import React, { Suspense, useState } from 'react'
import { useParams, Link as RouterLink } from 'react-router-dom'
import { Box, Container, Skeleton, Grid, Stack, Typography, Breadcrumbs, Link } from '@mui/material'
import NavigateNextIcon from '@mui/icons-material/NavigateNext'
import { VideoCard } from '~/components/video'
import VideoGridPagination from '~/components/video/VideoGridPagination'
import { useVideos, useTagDetail } from '~/hooks'
import type { VideoSort } from '~/types/video'

/**
 * VideoGridSkeleton - Loading state for video grid
 */
const VideoGridSkeleton: React.FC = () => (
  <Grid container spacing={3}>
    {Array.from({ length: 8 }).map((_, index) => (
      <Grid size={{ xs: 12, sm: 6, md: 4, lg: 3 }} key={index}>
        <Box>
          <Skeleton variant="rectangular" height={180} />
          <Skeleton variant="text" sx={{ mt: 1 }} />
          <Skeleton variant="text" width="60%" />
        </Box>
      </Grid>
    ))}
  </Grid>
)

/**
 * VideoGrid - Component that fetches and displays videos filtered by tag
 */
interface VideoGridProps {
  tagId: string
  sort?: VideoSort
  page: number
}

const VideoGrid: React.FC<VideoGridProps> = ({ tagId, sort = 'newest', page }) => {
  const { data } = useVideos({
    page,
    limit: 12,
    sort,
    tag_id: tagId,
    has_transcript: true,
  })

  if (data.data.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', py: 8 }}>
        <Typography variant="h6" color="text.secondary">
          Không tìm thấy video nào với tag này
        </Typography>
      </Box>
    )
  }

  return (
    <Grid container spacing={3}>
      {data.data.map((video) => (
        <Grid size={{ xs: 12, sm: 6, md: 4, lg: 3 }} key={video.id}>
          <VideoCard video={video} />
        </Grid>
      ))}
    </Grid>
  )
}

/**
 * TagPageContent - Inner component with data fetching
 */
interface TagPageContentProps {
  tagId: string
  page: number
  sort: VideoSort
  onPageChange: (page: number) => void
}

const TagPageContent: React.FC<TagPageContentProps> = ({ tagId, page, sort, onPageChange }) => {
  const { data: tag } = useTagDetail(tagId)

  const tagName = tag?.name || 'Tag'

  return (
    <>
      {/* Breadcrumbs */}
      <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />} sx={{ mb: 3 }}>
        <Link component={RouterLink} to="/" color="inherit" underline="hover">
          Trang chủ
        </Link>
        <Typography color="text.primary">{tagName}</Typography>
      </Breadcrumbs>

      {/* Page Title */}
      <Typography variant="h4" fontWeight={700} sx={{ mb: 4 }}>
        Videos về "{tagName}"
      </Typography>

      {/* Video Grid */}
      <VideoGrid tagId={tagId} sort={sort} page={page} />

      {/* Pagination */}
      <Stack alignItems="center" sx={{ mt: 4 }}>
        <VideoGridPagination
          page={page}
          onPageChange={onPageChange}
          selectedTagId={tagId}
          sort={sort}
        />
      </Stack>
    </>
  )
}

/**
 * TagPage Component
 * Displays videos filtered by a specific tag
 * Route: /tag/:tagId
 */
const TagPage: React.FC = () => {
  const { tagId } = useParams<{ tagId: string }>()
  const [page, setPage] = useState(1)
  const [sort] = useState<VideoSort>('newest')

  const handlePageChange = (value: number) => {
    setPage(value)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  if (!tagId) {
    return (
      <Container maxWidth="xl" sx={{ py: 4 }}>
        <Typography variant="h6" color="error">
          Tag ID không hợp lệ
        </Typography>
      </Container>
    )
  }

  return (
    <Box sx={{ bgcolor: 'background.default', minHeight: 'calc(100vh - 64px)' }}>
      <Container maxWidth="xl" sx={{ py: 3 }}>
        <Suspense
          fallback={
            <>
              <Skeleton variant="text" width={200} height={24} sx={{ mb: 3 }} />
              <Skeleton variant="text" width={300} height={40} sx={{ mb: 4 }} />
              <VideoGridSkeleton />
            </>
          }
        >
          <TagPageContent tagId={tagId} page={page} sort={sort} onPageChange={handlePageChange} />
        </Suspense>
      </Container>
    </Box>
  )
}

export default TagPage
