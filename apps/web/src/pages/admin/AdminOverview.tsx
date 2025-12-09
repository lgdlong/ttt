import React from 'react'
import { Box, Typography, Paper, Grid, CircularProgress, Alert } from '@mui/material'
import {
  People as PeopleIcon,
  VideoLibrary as VideoIcon,
  LocalOffer as TagIcon,
  CheckCircle as ActiveIcon,
} from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import { getAdminStats } from '~/api/statsApi'

interface StatCardProps {
  title: string
  value: string | number
  icon: React.ReactNode
  color: string
}

const StatCard: React.FC<StatCardProps> = ({ title, value, icon, color }) => (
  <Paper
    elevation={0}
    sx={{
      p: 3,
      border: '1px solid',
      borderColor: 'divider',
      height: '100%',
    }}
  >
    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
      <Box>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          {title}
        </Typography>
        <Typography variant="h4" fontWeight={700}>
          {value}
        </Typography>
      </Box>
      <Box
        sx={{
          p: 1,
          bgcolor: `${color}15`,
          color: color,
          borderRadius: 0,
        }}
      >
        {icon}
      </Box>
    </Box>
  </Paper>
)

/**
 * AdminOverview Component
 * Dashboard overview with statistics from API
 */
const AdminOverview: React.FC = () => {
  const {
    data: statsData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: getAdminStats,
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">Không thể tải thống kê. Vui lòng thử lại sau.</Alert>
      </Box>
    )
  }

  const stats = [
    {
      title: 'Tổng người dùng',
      value: statsData?.total_users.toLocaleString() || '0',
      icon: <PeopleIcon />,
      color: '#008080',
    },
    {
      title: 'Người dùng hoạt động',
      value: statsData?.active_users.toLocaleString() || '0',
      icon: <ActiveIcon />,
      color: '#4caf50',
    },
    {
      title: 'Tổng video',
      value: statsData?.total_videos.toLocaleString() || '0',
      icon: <VideoIcon />,
      color: '#2196f3',
    },
    {
      title: 'Tổng tags',
      value: statsData?.total_tags.toLocaleString() || '0',
      icon: <TagIcon />,
      color: '#ff9800',
    },
  ]

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h5" fontWeight={600} gutterBottom>
        Tổng quan
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Thống kê hệ thống TTT Archive
      </Typography>

      <Grid container spacing={3}>
        {stats.map((stat) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={stat.title}>
            <StatCard {...stat} />
          </Grid>
        ))}
      </Grid>

      <Paper
        elevation={0}
        sx={{
          p: 3,
          mt: 3,
          border: '1px solid',
          borderColor: 'divider',
        }}
      >
        <Typography variant="h6" fontWeight={600} gutterBottom>
          Hoạt động gần đây
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Chức năng đang phát triển...
        </Typography>
      </Paper>
    </Box>
  )
}

export default AdminOverview
