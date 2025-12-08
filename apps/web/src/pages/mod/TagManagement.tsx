import React, { useState, useCallback, useEffect } from 'react'
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  IconButton,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Chip,
  InputAdornment,
  Tooltip,
} from '@mui/material'
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Search as SearchIcon,
  LocalOffer as TagIcon,
} from '@mui/icons-material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axiosInstance from '~/lib/axios'
import type { TagResponse, CreateTagRequest, UpdateTagRequest } from '~/types/tag'

interface TagListResponse {
  tags: TagResponse[]
  total: number
  page: number
  page_size: number
}

// API functions
const fetchTags = async (params: {
  page: number
  pageSize: number
  q?: string
}): Promise<TagListResponse> => {
  const response = await axiosInstance.get('/mod/tags', {
    params: {
      page: params.page,
      page_size: params.pageSize,
      q: params.q || undefined,
    },
  })
  return response.data
}

const createTag = async (data: CreateTagRequest): Promise<TagResponse> => {
  const response = await axiosInstance.post('/mod/tags', data)
  return response.data
}

const updateTag = async ({
  id,
  data,
}: {
  id: number
  data: UpdateTagRequest
}): Promise<TagResponse> => {
  const response = await axiosInstance.put(`/mod/tags/${id}`, data)
  return response.data
}

const deleteTag = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/mod/tags/${id}`)
}

/**
 * TagManagement Component
 * Manage tags - create, edit, delete
 */
const TagManagement: React.FC = () => {
  const queryClient = useQueryClient()

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
  const [tagName, setTagName] = useState('')
  const [tagDescription, setTagDescription] = useState('')

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
  const createMutation = useMutation({
    mutationFn: createTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      handleCloseCreateDialog()
    },
  })

  const updateMutation = useMutation({
    mutationFn: updateTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      handleCloseEditDialog()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteTag,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mod-tags'] })
      handleCloseDeleteDialog()
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
    setTagName('')
    setTagDescription('')
    setOpenCreateDialog(true)
  }, [])

  const handleCloseCreateDialog = useCallback(() => {
    setOpenCreateDialog(false)
    setTagName('')
    setTagDescription('')
  }, [])

  const handleOpenEditDialog = useCallback((tag: TagResponse) => {
    setSelectedTag(tag)
    setTagName(tag.name)
    setTagDescription(tag.description || '')
    setOpenEditDialog(true)
  }, [])

  const handleCloseEditDialog = useCallback(() => {
    setOpenEditDialog(false)
    setSelectedTag(null)
    setTagName('')
    setTagDescription('')
  }, [])

  const handleOpenDeleteDialog = useCallback((tag: TagResponse) => {
    setSelectedTag(tag)
    setOpenDeleteDialog(true)
  }, [])

  const handleCloseDeleteDialog = useCallback(() => {
    setOpenDeleteDialog(false)
    setSelectedTag(null)
  }, [])

  const handleCreate = useCallback(() => {
    if (!tagName.trim()) return
    createMutation.mutate({
      name: tagName.trim(),
      description: tagDescription.trim() || undefined,
    })
  }, [tagName, tagDescription, createMutation])

  const handleUpdate = useCallback(() => {
    if (!selectedTag || !tagName.trim()) return
    updateMutation.mutate({
      id: selectedTag.id,
      data: {
        name: tagName.trim(),
        description: tagDescription.trim() || undefined,
      },
    })
  }, [selectedTag, tagName, tagDescription, updateMutation])

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
      <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Tên tag</TableCell>
                <TableCell>Mô tả</TableCell>
                <TableCell>Số video</TableCell>
                <TableCell align="right">Thao tác</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <CircularProgress size={32} />
                  </TableCell>
                </TableRow>
              ) : data?.tags && data.tags.length > 0 ? (
                data.tags.map((tag) => (
                  <TableRow key={tag.id} hover>
                    <TableCell>{tag.id}</TableCell>
                    <TableCell>
                      <Chip
                        icon={<TagIcon sx={{ fontSize: 16 }} />}
                        label={tag.name}
                        size="small"
                        sx={{ borderRadius: 0 }}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        noWrap
                        sx={{ maxWidth: 300 }}
                      >
                        {tag.description || '—'}
                      </Typography>
                    </TableCell>
                    <TableCell>{tag.video_count || 0}</TableCell>
                    <TableCell align="right">
                      <Tooltip title="Sửa">
                        <IconButton size="small" onClick={() => handleOpenEditDialog(tag)}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Xóa">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleOpenDeleteDialog(tag)}
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
                      {debouncedSearch ? 'Không tìm thấy tag phù hợp' : 'Chưa có tag nào'}
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          component="div"
          count={data?.total || 0}
          page={page}
          onPageChange={handleChangePage}
          rowsPerPage={pageSize}
          onRowsPerPageChange={handleChangeRowsPerPage}
          rowsPerPageOptions={[5, 10, 25, 50]}
          labelRowsPerPage="Số hàng:"
          labelDisplayedRows={({ from, to, count }) => `${from}–${to} / ${count}`}
        />
      </Paper>

      {/* Create Dialog */}
      <Dialog open={openCreateDialog} onClose={handleCloseCreateDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Thêm Tag mới</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            label="Tên tag"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            margin="normal"
            required
          />
          <TextField
            fullWidth
            label="Mô tả (tùy chọn)"
            value={tagDescription}
            onChange={(e) => setTagDescription(e.target.value)}
            margin="normal"
            multiline
            rows={2}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseCreateDialog}>Hủy</Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={!tagName.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? 'Đang tạo...' : 'Tạo'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={openEditDialog} onClose={handleCloseEditDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Sửa Tag</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            label="Tên tag"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            margin="normal"
            required
          />
          <TextField
            fullWidth
            label="Mô tả (tùy chọn)"
            value={tagDescription}
            onChange={(e) => setTagDescription(e.target.value)}
            margin="normal"
            multiline
            rows={2}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseEditDialog}>Hủy</Button>
          <Button
            variant="contained"
            onClick={handleUpdate}
            disabled={!tagName.trim() || updateMutation.isPending}
          >
            {updateMutation.isPending ? 'Đang lưu...' : 'Lưu'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Dialog */}
      <Dialog open={openDeleteDialog} onClose={handleCloseDeleteDialog}>
        <DialogTitle>Xác nhận xóa</DialogTitle>
        <DialogContent>
          <Typography>
            Bạn có chắc muốn xóa tag <strong>{selectedTag?.name}</strong>?
          </Typography>
          {selectedTag && selectedTag.video_count && selectedTag.video_count > 0 && (
            <Typography color="warning.main" sx={{ mt: 1 }}>
              Tag này đang được gắn với {selectedTag.video_count} video.
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDeleteDialog}>Hủy</Button>
          <Button
            variant="contained"
            color="error"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Đang xóa...' : 'Xóa'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default TagManagement
