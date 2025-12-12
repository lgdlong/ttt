import React from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
} from '@mui/material'
import type { TagResponse } from '~/types/tag'

interface DeleteConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  tag: TagResponse | null
  isDeleting: boolean
}

export const DeleteConfirmDialog: React.FC<DeleteConfirmDialogProps> = ({
  open,
  onClose,
  onConfirm,
  tag,
  isDeleting,
}) => {
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Xác nhận xóa</DialogTitle>
      <DialogContent>
        <Typography>
          Bạn có chắc muốn xóa tag <strong>{tag?.name}</strong>?
        </Typography>
        {!tag?.is_approved && (
          <Typography color="info.main" sx={{ mt: 1 }}>
            Tag này chưa được duyệt.
          </Typography>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Hủy</Button>
        <Button variant="contained" color="error" onClick={onConfirm} disabled={isDeleting}>
          {isDeleting ? 'Đang xóa...' : 'Xóa'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default DeleteConfirmDialog
