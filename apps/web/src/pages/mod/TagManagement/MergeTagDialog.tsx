import React, { useState, useCallback, useMemo } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Autocomplete,
  Typography,
  Box,
  Chip,
  Alert,
  CircularProgress,
} from '@mui/material'
import { MergeType as MergeIcon, LocalOffer as TagIcon } from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import type { TagResponse } from '~/types/tag'
import { searchTags } from './api'

interface MergeTagDialogProps {
  open: boolean
  onClose: () => void
  onMerge: (sourceId: string, targetId: string) => void
  sourceTag: TagResponse | null
  isMerging: boolean
}

export const MergeTagDialog: React.FC<MergeTagDialogProps> = ({
  open,
  onClose,
  onMerge,
  sourceTag,
  isMerging,
}) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTarget, setSelectedTarget] = useState<TagResponse | null>(null)

  // Search for tags to merge into
  const { data: searchResults, isLoading: isSearching } = useQuery({
    queryKey: ['tag-search', searchQuery],
    queryFn: () => searchTags(searchQuery, 20),
    enabled: searchQuery.length >= 1,
    staleTime: 30000,
  })

  // Filter out source tag from search results
  const filteredResults = useMemo(() => {
    if (!searchResults || !sourceTag) return []
    return searchResults.filter((tag) => tag.id !== sourceTag.id)
  }, [searchResults, sourceTag])

  const handleMerge = useCallback(() => {
    if (!sourceTag || !selectedTarget) return
    onMerge(sourceTag.id, selectedTarget.id)
  }, [sourceTag, selectedTarget, onMerge])

  const handleClose = useCallback(() => {
    setSearchQuery('')
    setSelectedTarget(null)
    onClose()
  }, [onClose])

  // Reset selected target when dialog opens with new source
  React.useEffect(() => {
    if (open) {
      setSelectedTarget(null)
      setSearchQuery('')
    }
  }, [open, sourceTag?.id])

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <MergeIcon color="primary" />
        Merge Tag
      </DialogTitle>
      <DialogContent>
        <Box sx={{ mb: 3 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Source tag (sẽ bị xóa sau khi merge):
          </Typography>
          {sourceTag && (
            <Chip
              icon={<TagIcon sx={{ fontSize: 16 }} />}
              label={sourceTag.name}
              color="error"
              variant="outlined"
              sx={{ mt: 1 }}
            />
          )}
        </Box>

        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Target tag (sẽ giữ lại, nhận aliases từ source):
          </Typography>
          <Autocomplete
            options={filteredResults}
            getOptionLabel={(option) => option.name}
            loading={isSearching}
            value={selectedTarget}
            onChange={(_, newValue) => setSelectedTarget(newValue)}
            inputValue={searchQuery}
            onInputChange={(_, newInputValue) => setSearchQuery(newInputValue)}
            isOptionEqualToValue={(option, value) => option.id === value.id}
            renderInput={(params) => (
              <TextField
                {...params}
                placeholder="Tìm kiếm tag..."
                size="small"
                margin="normal"
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
                  icon={<TagIcon sx={{ fontSize: 14 }} />}
                  label={option.name}
                  size="small"
                  color={option.is_approved ? 'success' : 'default'}
                  variant={option.is_approved ? 'filled' : 'outlined'}
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
        </Box>

        {selectedTarget && (
          <Alert severity="warning" sx={{ mt: 2 }}>
            <Typography variant="body2">
              Sau khi merge, tag <strong>{sourceTag?.name}</strong> sẽ trở thành alias của{' '}
              <strong>{selectedTarget.name}</strong> và source tag sẽ bị xóa.
            </Typography>
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={isMerging}>
          Hủy
        </Button>
        <Button
          variant="contained"
          color="primary"
          onClick={handleMerge}
          disabled={!selectedTarget || isMerging}
          startIcon={isMerging ? <CircularProgress size={16} /> : <MergeIcon />}
        >
          {isMerging ? 'Đang merge...' : 'Merge'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default MergeTagDialog
