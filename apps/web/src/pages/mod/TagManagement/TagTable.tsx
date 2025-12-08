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
  Chip,
  Tooltip,
} from '@mui/material'
import { Edit as EditIcon, Delete as DeleteIcon, LocalOffer as TagIcon } from '@mui/icons-material'
import type { TagResponse } from '~/types/tag'

interface TagTableProps {
  tags: TagResponse[]
  isLoading: boolean
  page: number
  pageSize: number
  total: number
  debouncedSearch: string
  onPageChange: (_: unknown, newPage: number) => void
  onRowsPerPageChange: (event: React.ChangeEvent<HTMLInputElement>) => void
  onEdit: (tag: TagResponse) => void
  onDelete: (tag: TagResponse) => void
}

export const TagTable: React.FC<TagTableProps> = ({
  tags,
  isLoading,
  page,
  pageSize,
  total,
  debouncedSearch,
  onPageChange,
  onRowsPerPageChange,
  onEdit,
  onDelete,
}) => {
  return (
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
            ) : tags.length > 0 ? (
              tags.map((tag) => (
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
                      <IconButton size="small" onClick={() => onEdit(tag)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Xóa">
                      <IconButton size="small" color="error" onClick={() => onDelete(tag)}>
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

export default TagTable
