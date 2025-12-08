import React from 'react'
import {
  Box,
  Container,
  Typography,
  Grid,
  Card,
  CardContent,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Button,
  IconButton,
  Tooltip,
} from '@mui/material'
import {
  Flag as FlagIcon,
  Visibility as VisibilityIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
  Report as ReportIcon,
  VideoLibrary as VideoIcon,
} from '@mui/icons-material'
import { useAuth } from '~/providers/AuthProvider'

interface ReportedContent {
  id: string
  type: 'video' | 'comment'
  title: string
  reportReason: string
  reportedAt: string
  reportCount: number
  status: 'pending' | 'reviewed' | 'resolved'
}

// Placeholder data - replace with actual API calls
const mockReportedContent: ReportedContent[] = [
  {
    id: '1',
    type: 'video',
    title: 'Video về Tiếng Anh giao tiếp',
    reportReason: 'Nội dung không phù hợp',
    reportedAt: '2024-01-15T10:30:00Z',
    reportCount: 3,
    status: 'pending',
  },
  {
    id: '2',
    type: 'comment',
    title: 'Comment trên video "Học IELTS"',
    reportReason: 'Spam/Quảng cáo',
    reportedAt: '2024-01-14T15:45:00Z',
    reportCount: 5,
    status: 'pending',
  },
  {
    id: '3',
    type: 'video',
    title: 'Video phát âm tiếng Anh',
    reportReason: 'Vi phạm bản quyền',
    reportedAt: '2024-01-13T09:20:00Z',
    reportCount: 1,
    status: 'reviewed',
  },
]

const ModDashboard: React.FC = () => {
  const { user } = useAuth()

  const stats = {
    pendingReports: 8,
    reviewedToday: 12,
    totalReports: 156,
    resolvedThisWeek: 45,
  }

  const getStatusChip = (status: ReportedContent['status']) => {
    switch (status) {
      case 'pending':
        return <Chip label="Chờ xử lý" color="warning" size="small" />
      case 'reviewed':
        return <Chip label="Đã xem" color="info" size="small" />
      case 'resolved':
        return <Chip label="Đã xử lý" color="success" size="small" />
      default:
        return <Chip label={status} size="small" />
    }
  }

  const getTypeIcon = (type: ReportedContent['type']) => {
    switch (type) {
      case 'video':
        return <VideoIcon fontSize="small" color="primary" />
      case 'comment':
        return <ReportIcon fontSize="small" color="secondary" />
      default:
        return null
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('vi-VN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <Container maxWidth="xl" sx={{ py: 4 }}>
      {/* Header */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" fontWeight="bold" gutterBottom>
          Mod Dashboard
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Xin chào, {user?.full_name || user?.username}! Đây là trang quản lý nội dung.
        </Typography>
      </Box>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    Báo cáo chờ xử lý
                  </Typography>
                  <Typography variant="h4" fontWeight="bold" color="warning.main">
                    {stats.pendingReports}
                  </Typography>
                </Box>
                <FlagIcon sx={{ fontSize: 40, color: 'warning.light' }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    Đã xem hôm nay
                  </Typography>
                  <Typography variant="h4" fontWeight="bold" color="info.main">
                    {stats.reviewedToday}
                  </Typography>
                </Box>
                <VisibilityIcon sx={{ fontSize: 40, color: 'info.light' }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    Tổng báo cáo
                  </Typography>
                  <Typography variant="h4" fontWeight="bold">
                    {stats.totalReports}
                  </Typography>
                </Box>
                <ReportIcon sx={{ fontSize: 40, color: 'grey.400' }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    Đã xử lý tuần này
                  </Typography>
                  <Typography variant="h4" fontWeight="bold" color="success.main">
                    {stats.resolvedThisWeek}
                  </Typography>
                </Box>
                <CheckCircleIcon sx={{ fontSize: 40, color: 'success.light' }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Reported Content Table */}
      <Card>
        <CardContent>
          <Box
            sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}
          >
            <Typography variant="h6" fontWeight="bold">
              Nội dung bị báo cáo
            </Typography>
            <Button variant="outlined" size="small">
              Xem tất cả
            </Button>
          </Box>

          <TableContainer component={Paper} variant="outlined">
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Loại</TableCell>
                  <TableCell>Nội dung</TableCell>
                  <TableCell>Lý do</TableCell>
                  <TableCell align="center">Số lượng</TableCell>
                  <TableCell>Thời gian</TableCell>
                  <TableCell>Trạng thái</TableCell>
                  <TableCell align="center">Hành động</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {mockReportedContent.map((report) => (
                  <TableRow key={report.id} hover>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {getTypeIcon(report.type)}
                        <Typography variant="body2" sx={{ textTransform: 'capitalize' }}>
                          {report.type === 'video' ? 'Video' : 'Bình luận'}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{
                          maxWidth: 200,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {report.title}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" color="error.main">
                        {report.reportReason}
                      </Typography>
                    </TableCell>
                    <TableCell align="center">
                      <Chip
                        label={report.reportCount}
                        size="small"
                        color={report.reportCount >= 5 ? 'error' : 'default'}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" color="text.secondary">
                        {formatDate(report.reportedAt)}
                      </Typography>
                    </TableCell>
                    <TableCell>{getStatusChip(report.status)}</TableCell>
                    <TableCell align="center">
                      <Box sx={{ display: 'flex', justifyContent: 'center', gap: 0.5 }}>
                        <Tooltip title="Xem chi tiết">
                          <IconButton size="small" color="primary">
                            <VisibilityIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Chấp nhận báo cáo">
                          <IconButton size="small" color="success">
                            <CheckCircleIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Từ chối báo cáo">
                          <IconButton size="small" color="error">
                            <CancelIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Box>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Container>
  )
}

export default ModDashboard
