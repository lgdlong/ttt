import React, { useState, useCallback } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Alert,
} from '@mui/material'

interface TagFormDialogProps {
  open: boolean
  onClose: () => void
  onSave: (name: string) => void
  isSaving: boolean
}

export const TagFormDialog: React.FC<TagFormDialogProps> = ({
  open,
  onClose,
  onSave,
  isSaving,
}) => {
  const [tagName, setTagName] = useState('')

  // Reset form when dialog opens
  React.useEffect(() => {
    if (open) {
      setTagName('')
    }
  }, [open])

  const handleSave = useCallback(() => {
    if (!tagName.trim()) return
    onSave(tagName.trim())
  }, [tagName, onSave])

  const handleClose = useCallback(() => {
    setTagName('')
    onClose()
  }, [onClose])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' && tagName.trim() && !isSaving) {
        handleSave()
      }
    },
    [tagName, isSaving, handleSave]
  )

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Thêm Tag mới</DialogTitle>
      <DialogContent>
        <Alert severity="info" sx={{ mb: 2, mt: 1 }}>
          Hệ thống sẽ tự động kiểm tra và merge với tag tương tự nếu có.
        </Alert>
        <TextField
          autoFocus
          fullWidth
          label="Tên tag"
          value={tagName}
          onChange={(e) => setTagName(e.target.value)}
          onKeyDown={handleKeyDown}
          margin="normal"
          required
          placeholder="Nhập tên tag..."
          helperText="Tag sẽ được chuẩn hóa tự động (ví dụ: 'tiền' → 'Money')"
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={isSaving}>
          Hủy
        </Button>
        <Button variant="contained" onClick={handleSave} disabled={!tagName.trim() || isSaving}>
          {isSaving ? 'Đang tạo...' : 'Tạo'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default TagFormDialog
