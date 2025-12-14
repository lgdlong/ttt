import React, { useState, useCallback, useEffect } from 'react'
import { Box, Typography, Button, Paper, TextField, InputAdornment, Chip } from '@mui/material'
import { Add as AddIcon, Search as SearchIcon } from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import type { Video } from '~types/video'
import type { TagResponse } from '~types/tag'
import { VideoTable } from './VideoTable'
import { AddVideoDialog } from './AddVideoDialog'
import { TagManagementDialog } from './TagManagementDialog'
import { DeleteConfirmDialog } from './DeleteConfirmDialog'
import { fetchModVideos, fetchAllTags } from '~/api/modApi'
import { useVideoMutations } from './useVideoMutations'

const VideoManagement: React.FC = () => {
  // State
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [scriptFilter, setScriptFilter] = useState<'all' | 'with' | 'without'>('all')

  // Dialog state
  const [openAddDialog, setOpenAddDialog] = useState(false)
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false)
  const [openTagDialog, setOpenTagDialog] = useState(false)
  const [selectedVideo, setSelectedVideo] = useState<Video | null>(null)

  // Tag management state
  const [videoTags, setVideoTags] = useState<TagResponse[]>([])
  const [tagsToAdd, setTagsToAdd] = useState<TagResponse[]>([])

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(0)
    }, 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // Queries
  const { data, isLoading, error } = useQuery({
    queryKey: ['mod-videos', page, pageSize, debouncedSearch, scriptFilter],
    queryFn: () => {
      let hasTranscriptParam: string | undefined
      if (scriptFilter === 'with') hasTranscriptParam = 'true'
      else if (scriptFilter === 'without') hasTranscriptParam = 'false'

      return fetchModVideos({
        page: page + 1,
        page_size: pageSize,
        q: debouncedSearch,
        has_transcript: hasTranscriptParam,
      })
    },
  })

  const { data: tagsData } = useQuery({
    queryKey: ['mod-all-tags'],
    queryFn: fetchAllTags,
  })

  // Mutations
  const { createMutation, deleteMutation, addTagMutation, removeTagMutation } = useVideoMutations({
    onCreateSuccess: () => handleCloseAddDialog(),
    onDeleteSuccess: () => handleCloseDeleteDialog(),
    onAddTagSuccess: (updatedTags) => {
      setVideoTags(updatedTags)
      setTagsToAdd([])
    },
    onRemoveTagSuccess: () => {
      if (selectedVideo && removeTagMutation.variables?.tagId) {
        const removedId = String(removeTagMutation.variables.tagId)
        setVideoTags((prev) => prev.filter((t) => t.id !== removedId))
      }
    },
  })

  // Handlers
  const handleChangePage = useCallback((_: unknown, newPage: number) => {
    setPage(newPage)
  }, [])

  const handleChangeRowsPerPage = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setPageSize(Number.parseInt(event.target.value, 10))
    setPage(0)
  }, [])

  const handleOpenAddDialog = useCallback(() => {
    setOpenAddDialog(true)
  }, [])

  const handleCloseAddDialog = useCallback(() => {
    setOpenAddDialog(false)
  }, [])

  const handleOpenDeleteDialog = useCallback((video: Video) => {
    setSelectedVideo(video)
    setOpenDeleteDialog(true)
  }, [])

  const handleCloseDeleteDialog = useCallback(() => {
    setOpenDeleteDialog(false)
    setSelectedVideo(null)
  }, [])

  const handleOpenTagDialog = useCallback((video: Video) => {
    setSelectedVideo(video)
    setVideoTags(video.tags || [])
    setTagsToAdd([])
    setOpenTagDialog(true)
  }, [])

  const handleCloseTagDialog = useCallback(() => {
    setOpenTagDialog(false)
    setSelectedVideo(null)
    setVideoTags([])
    setTagsToAdd([])
  }, [])

  const handleSaveVideo = useCallback(
    (youtubeId: string, tagIds?: number[]) => {
      createMutation.mutate({ youtubeId, tagIds })
    },
    [createMutation]
  )

  const handleDelete = useCallback(() => {
    if (!selectedVideo) return
    deleteMutation.mutate(selectedVideo.id)
  }, [selectedVideo, deleteMutation])

  const handleAddTagsToVideo = useCallback(() => {
    if (!selectedVideo || tagsToAdd.length === 0) return
    // Use v2 API - pass tag IDs as strings (they are UUIDs)
    const tagIds = tagsToAdd.map((tag) => tag.id)
    addTagMutation.mutate({ videoId: selectedVideo.id, tagIds })
  }, [selectedVideo, tagsToAdd, addTagMutation])

  const handleRemoveTagFromVideo = useCallback(
    (tagId: string) => {
      if (!selectedVideo) return
      removeTagMutation.mutate({
        videoId: selectedVideo.id,
        tagId,
      })
    },
    [selectedVideo, removeTagMutation]
  )

  // Note: availableTags filtering is now handled in TagManagementDialog

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography color="error">Lỗi khi tải danh sách video</Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            Quản lý Videos
          </Typography>
          <Typography color="text.secondary">Thêm video từ YouTube và quản lý tags</Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenAddDialog}
          sx={{ borderRadius: 0 }}
        >
          Thêm Video
        </Button>
      </Box>

      {/* Search & Filters */}
      <Paper elevation={0} sx={{ p: 2, mb: 3, border: '1px solid', borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <TextField
            size="small"
            placeholder="Tìm kiếm video theo tiêu đề..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            sx={{ flex: 1, minWidth: 300 }}
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
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Chip
              label="Tất cả"
              onClick={() => setScriptFilter('all')}
              color={scriptFilter === 'all' ? 'primary' : 'default'}
              sx={{ borderRadius: 0 }}
            />
            <Chip
              label="Có script"
              onClick={() => setScriptFilter('with')}
              color={scriptFilter === 'with' ? 'success' : 'default'}
              sx={{ borderRadius: 0 }}
            />
            <Chip
              label="Chưa có script"
              onClick={() => setScriptFilter('without')}
              color={scriptFilter === 'without' ? 'warning' : 'default'}
              sx={{ borderRadius: 0 }}
            />
          </Box>
        </Box>
      </Paper>

      {/* Videos Table */}
      <VideoTable
        videos={data?.videos || []}
        isLoading={isLoading}
        page={page}
        pageSize={pageSize}
        total={data?.total || 0}
        debouncedSearch={debouncedSearch}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        onOpenTagDialog={handleOpenTagDialog}
        onOpenDeleteDialog={handleOpenDeleteDialog}
      />

      {/* Add Video Dialog */}
      <AddVideoDialog
        open={openAddDialog}
        onClose={handleCloseAddDialog}
        onSave={handleSaveVideo}
        isSaving={createMutation.isPending}
        availableTags={tagsData?.tags || []}
      />

      {/* Delete Dialog */}
      <DeleteConfirmDialog
        open={openDeleteDialog}
        onClose={handleCloseDeleteDialog}
        onConfirm={handleDelete}
        video={selectedVideo}
        isDeleting={deleteMutation.isPending}
      />

      {/* Tag Management Dialog */}
      <TagManagementDialog
        open={openTagDialog}
        onClose={handleCloseTagDialog}
        video={selectedVideo}
        videoTags={videoTags}
        tagsToAdd={tagsToAdd}
        onTagsToAddChange={setTagsToAdd}
        onAddTags={handleAddTagsToVideo}
        onRemoveTag={handleRemoveTagFromVideo}
        isAdding={addTagMutation.isPending}
      />
    </Box>
  )
}

export default VideoManagement
