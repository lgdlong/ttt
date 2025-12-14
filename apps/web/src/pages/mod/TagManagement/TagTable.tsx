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
  Switch,
} from '@mui/material'
import { MergeType as MergeIcon, LocalOffer as TagIcon } from '@mui/icons-material'
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
  onMerge: (tag: TagResponse) => void
  onApprovalChange: (tag: TagResponse, isApproved: boolean) => void
  isApprovingId?: string // Currently approving tag ID
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
  onMerge,
  onApprovalChange,
  isApprovingId,
}) => {
  return (
    <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: 100 }}>ID</TableCell>
              <TableCell>Tên tag</TableCell>
              <TableCell align="center" sx={{ width: 120 }}>
                Duyệt
              </TableCell>
              <TableCell align="right" sx={{ width: 100 }}>
                Thao tác
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={32} />
                </TableCell>
              </TableRow>
            ) : tags.length > 0 ? (
              tags.map((tag) => (
                <TableRow key={tag.id} hover>
                  <TableCell>
                    <Typography
                      variant="caption"
                      color="text.secondary"
                      sx={{ fontFamily: 'monospace' }}
                    >
                      {tag.id.slice(0, 8)}...
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      icon={<TagIcon sx={{ fontSize: 16 }} />}
                      label={tag.name}
                      size="small"
                      color={tag.is_approved ? 'success' : 'default'}
                      variant={tag.is_approved ? 'filled' : 'outlined'}
                      sx={{ borderRadius: 1 }}
                    />
                  </TableCell>
                  <TableCell align="center">
                    <Tooltip title={tag.is_approved ? 'Đã duyệt' : 'Chưa duyệt'}>
                      <Switch
                        size="small"
                        checked={tag.is_approved}
                        disabled={isApprovingId === tag.id}
                        onChange={(e) => onApprovalChange(tag, e.target.checked)}
                        color="success"
                      />
                    </Tooltip>
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="Merge vào tag khác">
                      <IconButton size="small" onClick={() => onMerge(tag)} color="primary">
                        <MergeIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={4} align="center" sx={{ py: 4 }}>
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
