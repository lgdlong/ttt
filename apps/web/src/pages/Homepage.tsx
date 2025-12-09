import React, { Suspense, useState } from 'react'
import { Box, Container, Skeleton, Grid, Stack } from '@mui/material'
import { FilterBar, VideoCard } from '~/components/video'
import VideoGridPagination from '~/components/video/VideoGridPagination'
import { useVideos, useTags } from '~/hooks'
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
 * VideoGrid - Component that fetches and displays videos
 */
interface VideoGridProps {
  selectedTagId: string | null
  sort?: VideoSort
  page: number
}

const VideoGrid: React.FC<VideoGridProps> = ({ selectedTagId, sort = 'newest', page }) => {
  const { data } = useVideos({
    page,
    limit: 12,
    sort,
    tag_id: selectedTagId ?? undefined,
  })

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
 * Homepage Component
 * Main page showing video grid with tag filters and pagination
 * Following UI_Spec.md specifications
 */
const Homepage: React.FC = () => {
  const [selectedTagId, setSelectedTagId] = useState<string | null>(null)
  const [sort] = useState<VideoSort>('newest')
  const [page, setPage] = useState(1)
  const { data: tags = [] } = useTags()

  // Create categories from tags - add "All" option
  const categories = [
    { id: null, name: 'Tất cả' },
    ...tags.map((tag) => ({ id: tag.id, name: tag.name })),
  ]

  const handleCategoryChange = (categoryName: string) => {
    const category = categories.find((c) => c.name === categoryName)
    setSelectedTagId(category?.id ?? null)
    setPage(1) // Reset to page 1 when category changes
  }

  const handlePageChange = (value: number) => {
    setPage(value)
    // Scroll to top when page changes
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const selectedCategoryName = categories.find((c) => c.id === selectedTagId)?.name || 'Tất cả'

  return (
    <Box sx={{ bgcolor: 'background.default', minHeight: 'calc(100vh - 64px)' }}>
      <Container maxWidth="xl" sx={{ py: 3 }}>
        {/* Filter Bar */}
        <FilterBar
          categories={categories.map((c) => c.name)}
          selectedCategory={selectedCategoryName}
          onCategoryChange={handleCategoryChange}
        />

        {/* Video Grid with Suspense */}
        <Suspense fallback={<VideoGridSkeleton />}>
          <VideoGrid selectedTagId={selectedTagId} sort={sort} page={page} />
        </Suspense>

        {/* Pagination */}
        <Stack alignItems="center" sx={{ mt: 4 }}>
          <VideoGridPagination
            page={page}
            onPageChange={handlePageChange}
            selectedTagId={selectedTagId}
            sort={sort}
          />
        </Stack>
      </Container>
    </Box>
  )
}

export default Homepage
