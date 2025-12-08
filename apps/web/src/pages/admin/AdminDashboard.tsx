import React from 'react'
import { Box, Container, Typography, Paper, Grid, Card, CardContent, Chip } from '@mui/material'
import {
  People as PeopleIcon,
  VideoLibrary as VideoIcon,
  Settings as SettingsIcon,
  Security as SecurityIcon,
} from '@mui/icons-material'
import { useAuth } from '~/providers/AuthProvider'

/**
 * AdminDashboard Component
 * Admin control panel - only accessible by admin users
 */
const AdminDashboard: React.FC = () => {
  const { user } = useAuth()

  const adminCards = [
    {
      title: 'Quản lý người dùng',
      description: 'Xem, tạo, sửa, xóa tài khoản người dùng',
      icon: <PeopleIcon sx={{ fontSize: 40 }} />,
      path: '/admin/users',
      color: '#008080',
    },
    {
      title: 'Quản lý video',
      description: 'Quản lý video và transcript',
      icon: <VideoIcon sx={{ fontSize: 40 }} />,
      path: '/admin/videos',
      color: '#4caf50',
    },
    {
      title: 'Cài đặt hệ thống',
      description: 'Cấu hình hệ thống và tham số',
      icon: <SettingsIcon sx={{ fontSize: 40 }} />,
      path: '/admin/settings',
      color: '#ff9800',
    },
    {
      title: 'Bảo mật',
      description: 'Quản lý phiên đăng nhập và quyền truy cập',
      icon: <SecurityIcon sx={{ fontSize: 40 }} />,
      path: '/admin/security',
      color: '#f44336',
    },
  ]

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Paper
        elevation={0}
        sx={{
          p: 4,
          mb: 4,
          border: '1px solid',
          borderColor: 'divider',
          bgcolor: 'background.paper',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Typography variant="h4" fontWeight={700}>
            Admin Dashboard
          </Typography>
          <Chip label="Admin" color="error" size="small" sx={{ fontWeight: 600 }} />
        </Box>
        <Typography color="text.secondary">
          Xin chào, <strong>{user?.username}</strong>! Đây là trang quản trị hệ thống.
        </Typography>
      </Paper>

      <Grid container spacing={3}>
        {adminCards.map((card) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={card.title}>
            <Card
              sx={{
                height: '100%',
                cursor: 'pointer',
                transition: 'all 0.2s',
                border: '1px solid',
                borderColor: 'divider',
                '&:hover': {
                  borderColor: card.color,
                  transform: 'translateY(-4px)',
                  boxShadow: `0 4px 20px ${card.color}20`,
                },
              }}
            >
              <CardContent sx={{ textAlign: 'center', py: 4 }}>
                <Box sx={{ color: card.color, mb: 2 }}>{card.icon}</Box>
                <Typography variant="h6" fontWeight={600} gutterBottom>
                  {card.title}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {card.description}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Paper
        elevation={0}
        sx={{
          p: 3,
          mt: 4,
          border: '1px solid',
          borderColor: 'divider',
          bgcolor: 'warning.lighter',
        }}
      >
        <Typography variant="body2" color="text.secondary">
          <strong>Lưu ý:</strong> Các thao tác trong khu vực admin sẽ được ghi log. Vui lòng cẩn
          thận khi thực hiện các thay đổi.
        </Typography>
      </Paper>
    </Container>
  )
}

export default AdminDashboard
