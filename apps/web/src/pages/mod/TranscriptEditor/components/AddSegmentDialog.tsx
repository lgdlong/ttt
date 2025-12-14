import React, { useState } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Box,
  Alert,
} from '@mui/material'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import axiosInstance from '~/lib/axios'

interface AddSegmentDialogProps {
  open: boolean
  onClose: () => void
  videoId: string
}

interface CreateSegmentRequest {
  start_time: number
  end_time: number
  text: string
}

export const AddSegmentDialog: React.FC<AddSegmentDialogProps> = ({ open, onClose, videoId }) => {
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [text, setText] = useState('')
  const [validationError, setValidationError] = useState('')

  const queryClient = useQueryClient()

  const createSegmentMutation = useMutation({
    mutationFn: async (data: CreateSegmentRequest) => {
      const response = await axiosInstance.post(`/mod/videos/${videoId}/transcript/segments`, data)
      return response.data
    },
    onSuccess: () => {
      // Invalidate transcript query để reload segments
      queryClient.invalidateQueries({ queryKey: ['transcript', videoId] })
      handleClose()
    },
    onError: (error: unknown) => {
      const err = error as { response?: { data?: { message?: string } } }
      setValidationError(err.response?.data?.message || 'Có lỗi xảy ra khi tạo segment')
    },
  })

  const handleClose = () => {
    setStartTime('')
    setEndTime('')
    setText('')
    setValidationError('')
    onClose()
  }

  const handleSubmit = () => {
    setValidationError('')

    // Validation
    const start = Number.parseInt(startTime, 10)
    const end = Number.parseInt(endTime, 10)

    if (Number.isNaN(start) || start < 0) {
      setValidationError('Start time phải là số >= 0')
      return
    }

    if (Number.isNaN(end) || end <= start) {
      setValidationError('End time phải lớn hơn start time')
      return
    }

    if (!text.trim()) {
      setValidationError('Text không được để trống')
      return
    }

    createSegmentMutation.mutate({
      start_time: start,
      end_time: end,
      text: text.trim(),
    })
  }

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Thêm Segment Mới</DialogTitle>
      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          {validationError && (
            <Alert severity="error" onClose={() => setValidationError('')}>
              {validationError}
            </Alert>
          )}

          <TextField
            label="Start Time (milliseconds)"
            type="number"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            fullWidth
            required
            helperText="Thời gian bắt đầu tính bằng milliseconds (1 giây = 1000ms)"
          />

          <TextField
            label="End Time (milliseconds)"
            type="number"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            fullWidth
            required
            helperText="Thời gian kết thúc tính bằng milliseconds"
          />

          <TextField
            label="Text"
            value={text}
            onChange={(e) => setText(e.target.value)}
            fullWidth
            required
            multiline
            rows={4}
            helperText="Nội dung phụ đề của segment"
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={createSegmentMutation.isPending}>
          Hủy
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={createSegmentMutation.isPending}
        >
          {createSegmentMutation.isPending ? 'Đang tạo...' : 'Tạo'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
