import React from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Chip,
  Stack,
  Autocomplete,
  TextField,
} from '@mui/material'
import type { Video } from '~types/video'
import type { TagResponse } from '~types/tag'

interface TagManagementDialogProps {
  open: boolean
  onClose: () => void
  video: Video | null
  videoTags: TagResponse[]
  availableTags: TagResponse[]
  tagToAdd: TagResponse | null
  onTagToAddChange: (tag: TagResponse | null) => void
  onAddTag: () => void
  onRemoveTag: (tagId: number) => void
  isAdding: boolean
}

export const TagManagementDialog: React.FC<TagManagementDialogProps> = ({
  open,
  onClose,
  video,
  videoTags,
  availableTags,
  tagToAdd,
  onTagToAddChange,
  onAddTag,
  onRemoveTag,
  isAdding,
}) => {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Quản lý Tags cho Video</DialogTitle>
      <DialogContent>
        {video && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              {video.title}
            </Typography>
          </Box>
        )}

        {/* Current tags */}
        <Typography variant="subtitle2" gutterBottom>
          Tags hiện tại:
        </Typography>
        <Box sx={{ mb: 2, minHeight: 40 }}>
          {videoTags.length > 0 ? (
            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              {videoTags.map((tag) => (
                <Chip
                  key={tag.id}
                  label={tag.name}
                  size="small"
                  onDelete={() => onRemoveTag(tag.id)}
                  sx={{ borderRadius: 0, mb: 1 }}
                />
              ))}
            </Stack>
          ) : (
            <Typography variant="body2" color="text.secondary">
              Chưa có tag nào
            </Typography>
          )}
        </Box>

        {/* Add tag */}
        <Typography variant="subtitle2" gutterBottom>
          Thêm tag:
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Autocomplete
            fullWidth
            options={availableTags}
            getOptionLabel={(option) => option.name}
            value={tagToAdd}
            onChange={(_, newValue) => onTagToAddChange(newValue)}
            renderInput={(params) => (
              <TextField {...params} size="small" placeholder="Chọn tag..." />
            )}
          />
          <Button
            variant="contained"
            onClick={onAddTag}
            disabled={!tagToAdd || isAdding}
            sx={{ borderRadius: 0 }}
          >
            Thêm
          </Button>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Đóng</Button>
      </DialogActions>
    </Dialog>
  )
}

export default TagManagementDialog
