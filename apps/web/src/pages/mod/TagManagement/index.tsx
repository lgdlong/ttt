import React, { useState, useCallback, useEffect } from 'react'
import { Box, Typography, Button, Paper, TextField, InputAdornment } from '@mui/material'
import { Add as AddIcon, Search as SearchIcon } from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import type { TagResponse } from '~/types/tag'
import { TagTable } from './TagTable'
import { TagFormDialog } from './TagFormDialog'
import { DeleteConfirmDialog } from './DeleteConfirmDialog'
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
  const [openEditDialog, setOpenEditDialog] = useState(false)
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false)
  const [selectedTag, setSelectedTag] = useState<TagResponse | null>(null)

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
    queryFn: () => fetchTags({ page: page + 1, pageSize, q: debouncedSearch }),
  })

  // Mutations
  const { createMutation, updateMutation, deleteMutation } = useTagMutations({
    onCreateSuccess: () => handleCloseCreateDialog(),
    onUpdateSuccess: () => handleCloseEditDialog(),
    onDeleteSuccess: () => handleCloseDeleteDialog(),
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

  const handleOpenEditDialog = useCallback((tag: TagResponse) => {
    setSelectedTag(tag)
    setOpenEditDialog(true)
  }, [])

  const handleCloseEditDialog = useCallback(() => {
    setOpenEditDialog(false)
    setSelectedTag(null)
  }, [])

  const handleOpenDeleteDialog = useCallback((tag: TagResponse) => {
    setSelectedTag(tag)
    setOpenDeleteDialog(true)
  }, [])

  const handleCloseDeleteDialog = useCallback(() => {
    setOpenDeleteDialog(false)
    setSelectedTag(null)
  }, [])

  const handleCreate = useCallback(
    (name: string, description?: string) => {
      createMutation.mutate({ name, description })
    },
    [createMutation]
  )

  const handleUpdate = useCallback(
    (name: string, description?: string) => {
      if (!selectedTag) return
      updateMutation.mutate({
        id: selectedTag.id,
        data: { name, description },
      })
    },
    [selectedTag, updateMutation]
  )

  const handleDelete = useCallback(() => {
    if (!selectedTag) return
    deleteMutation.mutate(selectedTag.id)
  }, [selectedTag, deleteMutation])

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
            Quản lý Tags
          </Typography>
          <Typography color="text.secondary">Tạo và quản lý tags cho video</Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenCreateDialog}
          sx={{ borderRadius: 0 }}
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
        onEdit={handleOpenEditDialog}
        onDelete={handleOpenDeleteDialog}
      />

      {/* Create Dialog */}
      <TagFormDialog
        open={openCreateDialog}
        onClose={handleCloseCreateDialog}
        onSave={handleCreate}
        tag={null}
        isSaving={createMutation.isPending}
        mode="create"
      />

      {/* Edit Dialog */}
      <TagFormDialog
        open={openEditDialog}
        onClose={handleCloseEditDialog}
        onSave={handleUpdate}
        tag={selectedTag}
        isSaving={updateMutation.isPending}
        mode="edit"
      />

      {/* Delete Dialog */}
      <DeleteConfirmDialog
        open={openDeleteDialog}
        onClose={handleCloseDeleteDialog}
        onConfirm={handleDelete}
        tag={selectedTag}
        isDeleting={deleteMutation.isPending}
      />
    </Box>
  )
}

export default TagManagement
