import React, { useState, useEffect, useCallback } from 'react'
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
  TextField,
  InputAdornment,
  IconButton,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
  Alert,
  Tooltip,
} from '@mui/material'
import {
  Search as SearchIcon,
  Block as BlockIcon,
  CheckCircle as CheckCircleIcon,
  Refresh as RefreshIcon,
  FilterList as FilterIcon,
} from '@mui/icons-material'
import { listUsers, updateUser } from '~/api/userApi'
import type { UserResponse, UserRole, ListUserRequest } from '~/types/user'

/**
 * UserManagement Component
 * Manages users with search, ban, and role display
 */
const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<UserResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Pagination
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(10)
  const [totalItems, setTotalItems] = useState(0)

  // Filters
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  // Ban dialog
  const [banDialogOpen, setBanDialogOpen] = useState(false)
  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params: ListUserRequest = {
        page: page + 1, // API uses 1-based pagination
        limit: rowsPerPage,
      }

      if (searchQuery) params.q = searchQuery
      if (roleFilter) params.role = roleFilter as UserRole
      if (statusFilter === 'active') params.is_active = true
      if (statusFilter === 'banned') params.is_active = false

      const response = await listUsers(params)
      setUsers(response.data)
      setTotalItems(response.pagination.total_items)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không thể tải danh sách người dùng')
    } finally {
      setLoading(false)
    }
  }, [page, rowsPerPage, searchQuery, roleFilter, statusFilter])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setPage(0) // Reset to first page on search
    }, 300)
    return () => clearTimeout(timer)
  }, [searchQuery])

  const handleChangePage = (_: unknown, newPage: number) => {
    setPage(newPage)
  }

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10))
    setPage(0)
  }

  const handleBanClick = (user: UserResponse) => {
    setSelectedUser(user)
    setBanDialogOpen(true)
  }

  const handleBanConfirm = async () => {
    if (!selectedUser) return

    setActionLoading(true)
    try {
      await updateUser({ id: selectedUser.id, data: { is_active: !selectedUser.is_active } })
      // Refresh list
      await fetchUsers()
      setBanDialogOpen(false)
      setSelectedUser(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Thao tác thất bại')
    } finally {
      setActionLoading(false)
    }
  }

  const getRoleChip = (role: UserRole) => {
    const config = {
      admin: { label: 'Admin', color: 'error' as const },
      mod: { label: 'Mod', color: 'warning' as const },
      user: { label: 'User', color: 'default' as const },
    }
    const { label, color } = config[role] || config.user
    return <Chip label={label} color={color} size="small" sx={{ fontWeight: 600 }} />
  }

  const getStatusChip = (isActive: boolean) => {
    return isActive ? (
      <Chip
        icon={<CheckCircleIcon />}
        label="Hoạt động"
        color="success"
        size="small"
        variant="outlined"
      />
    ) : (
      <Chip icon={<BlockIcon />} label="Bị khóa" color="error" size="small" variant="outlined" />
    )
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('vi-VN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            Quản lý người dùng
          </Typography>
          <Typography color="text.secondary">Tổng cộng {totalItems} người dùng</Typography>
        </Box>
        <IconButton onClick={fetchUsers} disabled={loading}>
          <RefreshIcon />
        </IconButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Filters */}
      <Paper
        elevation={0}
        sx={{
          p: 2,
          mb: 3,
          border: '1px solid',
          borderColor: 'divider',
          display: 'flex',
          gap: 2,
          flexWrap: 'wrap',
          alignItems: 'center',
        }}
      >
        <TextField
          placeholder="Tìm kiếm theo username, email, tên..."
          size="small"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          sx={{ minWidth: 300, flex: 1 }}
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

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Vai trò</InputLabel>
          <Select
            value={roleFilter}
            label="Vai trò"
            onChange={(e) => {
              setRoleFilter(e.target.value)
              setPage(0)
            }}
          >
            <MenuItem value="">Tất cả</MenuItem>
            <MenuItem value="admin">Admin</MenuItem>
            <MenuItem value="mod">Mod</MenuItem>
            <MenuItem value="user">User</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Trạng thái</InputLabel>
          <Select
            value={statusFilter}
            label="Trạng thái"
            onChange={(e) => {
              setStatusFilter(e.target.value)
              setPage(0)
            }}
          >
            <MenuItem value="">Tất cả</MenuItem>
            <MenuItem value="active">Hoạt động</MenuItem>
            <MenuItem value="banned">Bị khóa</MenuItem>
          </Select>
        </FormControl>

        {(searchQuery || roleFilter || statusFilter) && (
          <Button
            variant="outlined"
            size="small"
            startIcon={<FilterIcon />}
            onClick={() => {
              setSearchQuery('')
              setRoleFilter('')
              setStatusFilter('')
              setPage(0)
            }}
          >
            Xóa bộ lọc
          </Button>
        )}
      </Paper>

      {/* Users Table */}
      <Paper elevation={0} sx={{ border: '1px solid', borderColor: 'divider' }}>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Người dùng</TableCell>
                <TableCell>Email</TableCell>
                <TableCell>Vai trò</TableCell>
                <TableCell>Trạng thái</TableCell>
                <TableCell>Ngày tạo</TableCell>
                <TableCell align="right">Hành động</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                    <CircularProgress size={32} />
                  </TableCell>
                </TableRow>
              ) : users.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">Không tìm thấy người dùng</Typography>
                  </TableCell>
                </TableRow>
              ) : (
                users.map((user) => (
                  <TableRow key={user.id} hover>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight={600}>
                          {user.username}
                        </Typography>
                        {user.full_name && (
                          <Typography variant="caption" color="text.secondary">
                            {user.full_name}
                          </Typography>
                        )}
                      </Box>
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>{getRoleChip(user.role)}</TableCell>
                    <TableCell>{getStatusChip(user.is_active)}</TableCell>
                    <TableCell>{formatDate(user.created_at)}</TableCell>
                    <TableCell align="right">
                      <Tooltip title={user.is_active ? 'Khóa tài khoản' : 'Mở khóa tài khoản'}>
                        <IconButton
                          size="small"
                          onClick={() => handleBanClick(user)}
                          color={user.is_active ? 'error' : 'success'}
                        >
                          {user.is_active ? <BlockIcon /> : <CheckCircleIcon />}
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={totalItems}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
          labelRowsPerPage="Hiển thị:"
          labelDisplayedRows={({ from, to, count }) =>
            `${from}-${to} trong ${count !== -1 ? count : `hơn ${to}`}`
          }
        />
      </Paper>

      {/* Ban/Unban Dialog */}
      <Dialog open={banDialogOpen} onClose={() => setBanDialogOpen(false)}>
        <DialogTitle>
          {selectedUser?.is_active ? 'Khóa tài khoản' : 'Mở khóa tài khoản'}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            Bạn có chắc chắn muốn {selectedUser?.is_active ? 'khóa' : 'mở khóa'} tài khoản{' '}
            <strong>{selectedUser?.username}</strong>?
            {selectedUser?.is_active && (
              <>
                <br />
                <br />
                Người dùng này sẽ không thể đăng nhập vào hệ thống cho đến khi được mở khóa.
              </>
            )}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBanDialogOpen(false)} disabled={actionLoading}>
            Hủy
          </Button>
          <Button
            onClick={handleBanConfirm}
            variant="contained"
            color={selectedUser?.is_active ? 'error' : 'success'}
            disabled={actionLoading}
          >
            {actionLoading ? (
              <CircularProgress size={20} color="inherit" />
            ) : selectedUser?.is_active ? (
              'Khóa'
            ) : (
              'Mở khóa'
            )}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default UserManagement
