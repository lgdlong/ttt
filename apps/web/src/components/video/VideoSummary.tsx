import React from 'react'
import { Typography, Paper } from '@mui/material'

interface VideoSummaryProps {
  summary?: string
}

export const VideoSummary: React.FC<VideoSummaryProps> = ({ summary }) => {
  if (!summary) return null

  return (
    <Paper elevation={0} variant="outlined" sx={{ p: 3, mb: 3, borderRadius: 2, bgcolor: 'background.default' }}>
      <Typography variant="h6" gutterBottom fontWeight={700} fontFamily="Inter">
        Tóm tắt nội dung
      </Typography>
      <Typography variant="body1" lineHeight={1.8} color="text.primary">
        {summary}
      </Typography>
    </Paper>
  )
}
