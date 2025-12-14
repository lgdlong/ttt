import React, { useState } from 'react'
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
  CircularProgress,
} from '@mui/material'
import { useQuery } from '@tanstack/react-query'
import { searchApprovedTags } from '~/api/modApi'
import type { Video } from '~types/video'
import type { TagResponse } from '~types/tag'

interface TagManagementDialogProps {
  open: boolean
  onClose: () => void
  video: Video | null
  videoTags: TagResponse[]
  tagsToAdd: TagResponse[]
  onTagsToAddChange: (tags: TagResponse[]) => void
  onAddTags: () => void
  onRemoveTag: (tagId: string) => void
  isAdding: boolean
}

export const TagManagementDialog: React.FC<TagManagementDialogProps> = ({
  open,
  onClose,
  video,
  videoTags,
  tagsToAdd,
  onTagsToAddChange,
  onAddTags,
  onRemoveTag,
  isAdding,
}) => {
  const [searchQuery, setSearchQuery] = useState('')

  // Query approved tags với debounce
  const { data: searchResults = [], isLoading: isSearching } = useQuery({
    queryKey: ['approved-tags-search', searchQuery],
    queryFn: () => searchApprovedTags(searchQuery, 20),
    enabled: searchQuery.length >= 1,
    staleTime: 30000, // Cache 30s
  })

  // Filter out tags already added to video
  const availableTags = searchResults.filter((tag) => !videoTags.some((vTag) => vTag.id === tag.id))
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
                  onDelete={() => onRemoveTag(tag.id as unknown as string)}
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

        {/* Add tags - Multi-select with search */}
        <Typography variant="subtitle2" gutterBottom>
          Thêm tags (có thể chọn nhiều):
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, flexDirection: 'column' }}>
          <Autocomplete
            multiple
            fullWidth
            options={availableTags}
            getOptionLabel={(option) => option.name}
            value={tagsToAdd}
            onChange={(_, newValue) => onTagsToAddChange(newValue)}
            loading={isSearching}
            inputValue={searchQuery}
            onInputChange={(_, newInputValue) => setSearchQuery(newInputValue)}
            isOptionEqualToValue={(option, value) => option.id === value.id}
            filterOptions={(x) => x} // Disable client-side filtering (use server search)
            renderInput={(params) => (
              <TextField
                {...params}
                size="small"
                placeholder="Tìm kiếm tag (chỉ approved)..."
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {isSearching && <CircularProgress color="inherit" size={20} />}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
            renderOption={(props, option) => (
              <li {...props} key={option.id}>
                <Chip
                  label={option.name}
                  size="small"
                  color="success"
                  variant="outlined"
                  sx={{ borderRadius: 1, mr: 1 }}
                />
                {option.is_approved && (
                  <Typography variant="caption" color="success.main">
                    (đã duyệt)
                  </Typography>
                )}
              </li>
            )}
            noOptionsText={searchQuery.length < 1 ? 'Nhập để tìm kiếm...' : 'Không tìm thấy tag'}
          />
          <Button
            variant="contained"
            onClick={onAddTags}
            disabled={tagsToAdd.length === 0 || isAdding}
            sx={{ borderRadius: 0 }}
            fullWidth
          >
            {isAdding ? 'Đang thêm...' : `Thêm ${tagsToAdd.length} tag(s)`}
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
