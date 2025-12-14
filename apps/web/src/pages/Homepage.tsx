import React, { useState } from 'react'
import {
  Box,
  Container,
  Skeleton,
  Grid,
  Stack,
  useMediaQuery,
  useTheme,
  CircularProgress,
} from '@mui/material'
import { FilterBar, VideoCard } from '~/components/video'
import VideoGridPagination from '~/components/video/VideoGridPagination'
import { TagSidebar } from '~/components/sidebar'
import { useVideosQuery, useTags } from '~/hooks'
import type { VideoSort } from '~/types/video'

/** Sidebar width constant */
const SIDEBAR_WIDTH = 260

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
  const { data, isLoading } = useVideosQuery({
    page,
    limit: 12,
    sort,
    tag_id: selectedTagId ?? undefined,
    has_transcript: true,
  })

  // Show loading spinner in the center if data is loading
  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    )
  }

  // Show skeleton if data hasn't loaded yet
  if (!data) {
    return <VideoGridSkeleton />
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

  // Check if we should show sidebar (desktop only)
  const theme = useTheme()
  const isDesktop = useMediaQuery(theme.breakpoints.up('md'))

  return (
    <Box
      sx={{
        bgcolor: 'background.default',
        minHeight: 'calc(100vh - 64px)',
        display: 'flex',
      }}
    >
      {/* Sidebar - Desktop only */}
      {isDesktop && (
        <Box
          component="aside"
          sx={{
            width: SIDEBAR_WIDTH,
            flexShrink: 0,
            position: 'sticky',
            top: 64, // Height of navbar
            height: 'calc(100vh - 64px)',
            overflow: 'hidden',
          }}
        >
          <TagSidebar />
        </Box>
      )}

      {/* Main Content */}
      <Box component="main" sx={{ flex: 1, minWidth: 0 }}>
        <Container maxWidth="xl" sx={{ py: 3 }}>
          {/* Filter Bar */}
          <FilterBar
            categories={categories.map((c) => c.name)}
            selectedCategory={selectedCategoryName}
            onCategoryChange={handleCategoryChange}
          />

          {/* Video Grid - No Suspense boundary needed, uses regular useQuery with loading state */}
          <VideoGrid selectedTagId={selectedTagId} sort={sort} page={page} />

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
    </Box>
  )
}

export default Homepage
