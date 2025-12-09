import React, { useState, useCallback } from 'react'
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField } from '@mui/material'
import type { TagResponse } from '~/types/tag'

interface TagFormDialogProps {
  open: boolean
  onClose: () => void
  onSave: (name: string, description?: string) => void
  tag: TagResponse | null
  isSaving: boolean
  mode: 'create' | 'edit'
}

export const TagFormDialog: React.FC<TagFormDialogProps> = ({
  open,
  onClose,
  onSave,
  tag,
  isSaving,
  mode,
}) => {
  const [tagName, setTagName] = useState(tag?.name || '')
  const [tagDescription, setTagDescription] = useState(tag?.description || '')

  // Reset form when dialog opens/closes or tag changes
  React.useEffect(() => {
    if (open) {
      setTagName(tag?.name || '')
      setTagDescription(tag?.description || '')
    }
  }, [open, tag])

  const handleSave = useCallback(() => {
    if (!tagName.trim()) return
    onSave(tagName.trim(), tagDescription.trim() || undefined)
  }, [tagName, tagDescription, onSave])

  const handleClose = useCallback(() => {
    setTagName('')
    setTagDescription('')
    onClose()
  }, [onClose])

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{mode === 'create' ? 'Thêm Tag mới' : 'Sửa Tag'}</DialogTitle>
      <DialogContent>
        <TextField
          autoFocus
          fullWidth
          label="Tên tag"
          value={tagName}
          onChange={(e) => setTagName(e.target.value)}
          margin="normal"
          required
        />
        <TextField
          fullWidth
          label="Mô tả (tùy chọn)"
          value={tagDescription}
          onChange={(e) => setTagDescription(e.target.value)}
          margin="normal"
          multiline
          rows={2}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Hủy</Button>
        <Button variant="contained" onClick={handleSave} disabled={!tagName.trim() || isSaving}>
          {isSaving
            ? mode === 'create'
              ? 'Đang tạo...'
              : 'Đang lưu...'
            : mode === 'create'
              ? 'Tạo'
              : 'Lưu'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default TagFormDialog
