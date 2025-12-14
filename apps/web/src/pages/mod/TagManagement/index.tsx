import React, { useState, useCallback, useEffect } from 'react'
import {
  Box,
  Typography,
  Button,
  Paper,
  TextField,
  InputAdornment,
  Snackbar,
  Alert,
} from '@mui/material'
import { Add as AddIcon, Search as SearchIcon } from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import type { TagResponse } from '~/types/tag'
import { TagTable } from './TagTable'
import { TagFormDialog } from './TagFormDialog'
import { MergeTagDialog } from './MergeTagDialog'
import { fetchTags } from './api'
import { useTagMutations } from './useTagMutations'

const TagManagement: React.FC = () => {
  // State
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Dialog state
  const [openCreateDialog, setOpenCreateDialog] = useState(false)
  const [openMergeDialog, setOpenMergeDialog] = useState(false)
  const [selectedTag, setSelectedTag] = useState<TagResponse | null>(null)
  const [approvingTagId, setApprovingTagId] = useState<string | undefined>()

  // Snackbar state
  const [snackbar, setSnackbar] = useState<{
    open: boolean
    message: string
    severity: 'success' | 'error' | 'info'
  }>({ open: false, message: '', severity: 'success' })

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(0)
    }, 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  // Query
  const { data, isLoading, error } = useQuery({
    queryKey: ['mod-tags', page, pageSize, debouncedSearch],
    queryFn: () => fetchTags({ page: page + 1, pageSize, query: debouncedSearch }),
  })

  // Mutations
  const { createMutation, mergeMutation, approvalMutation } = useTagMutations({
    onCreateSuccess: () => {
      handleCloseCreateDialog()
      setSnackbar({ open: true, message: 'Tag đã được tạo thành công!', severity: 'success' })
    },
    onMergeSuccess: () => {
      handleCloseMergeDialog()
      setSnackbar({ open: true, message: 'Merge tag thành công!', severity: 'success' })
    },
    onApprovalSuccess: () => {
      setApprovingTagId(undefined)
    },
  })

  // Handlers
  const handleChangePage = useCallback((_: unknown, newPage: number) => {
    setPage(newPage)
  }, [])

  const handleChangeRowsPerPage = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setPageSize(parseInt(event.target.value, 10))
    setPage(0)
  }, [])

  const handleOpenCreateDialog = useCallback(() => {
    setSelectedTag(null)
    setOpenCreateDialog(true)
  }, [])

  const handleCloseCreateDialog = useCallback(() => {
    setOpenCreateDialog(false)
    setSelectedTag(null)
  }, [])

  const handleOpenMergeDialog = useCallback((tag: TagResponse) => {
    setSelectedTag(tag)
    setOpenMergeDialog(true)
  }, [])

  const handleCloseMergeDialog = useCallback(() => {
    setOpenMergeDialog(false)
    setSelectedTag(null)
  }, [])

  const handleCreate = useCallback(
    (name: string) => {
      createMutation.mutate({ name })
    },
    [createMutation]
  )

  const handleMerge = useCallback(
    (sourceId: string, targetId: string) => {
      mergeMutation.mutate({ source_id: sourceId, target_id: targetId })
    },
    [mergeMutation]
  )

  const handleApprovalChange = useCallback(
    (tag: TagResponse, isApproved: boolean) => {
      setApprovingTagId(tag.id)
      approvalMutation.mutate({
        id: tag.id,
        data: { is_approved: isApproved },
      })
    },
    [approvalMutation]
  )

  const handleCloseSnackbar = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, open: false }))
  }, [])

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography color="error">Lỗi khi tải danh sách tags</Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            Quản lý Tags (V2)
          </Typography>
          <Typography color="text.secondary">
            Tạo, duyệt và merge canonical tags cho video
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenCreateDialog}
          sx={{ borderRadius: 1 }}
        >
          Thêm Tag
        </Button>
      </Box>

      {/* Search */}
      <Paper elevation={0} sx={{ p: 2, mb: 3, border: '1px solid', borderColor: 'divider' }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Tìm kiếm tag..."
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

      {/* Tags Table */}
      <TagTable
        tags={data?.tags || []}
        isLoading={isLoading}
        page={page}
        pageSize={pageSize}
        total={data?.total || 0}
        debouncedSearch={debouncedSearch}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        onMerge={handleOpenMergeDialog}
        onApprovalChange={handleApprovalChange}
        isApprovingId={approvingTagId}
      />

      {/* Create Dialog */}
      <TagFormDialog
        open={openCreateDialog}
        onClose={handleCloseCreateDialog}
        onSave={handleCreate}
        isSaving={createMutation.isPending}
      />

      {/* Merge Dialog */}
      <MergeTagDialog
        open={openMergeDialog}
        onClose={handleCloseMergeDialog}
        onMerge={handleMerge}
        sourceTag={selectedTag}
        isMerging={mergeMutation.isPending}
      />

      {/* Snackbar */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  )
}

export default TagManagement
