import React from 'react'
import { Box, Button, Alert, Paper, Typography } from '@mui/material'
import { Save, CheckCircle, Keyboard } from '@mui/icons-material'

interface EditorControlsProps {
  saving: boolean
  hasUnsavedChanges: boolean
  saveError: string | null
  showShortcuts: boolean
  onSaveDraft: () => void
  onVerify: () => void
  onDismissShortcuts: () => void
}

export const EditorControls: React.FC<EditorControlsProps> = ({
  saving,
  hasUnsavedChanges,
  saveError,
  showShortcuts,
  onSaveDraft,
  onVerify,
  onDismissShortcuts,
}) => {
  return (
    <>
      <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
        <Button
          variant="outlined"
          startIcon={<Save />}
          onClick={onSaveDraft}
          disabled={saving || !hasUnsavedChanges}
          fullWidth
        >
          {saving ? 'Đang lưu...' : 'Lưu nháp'}
        </Button>
        <Button
          variant="contained"
          startIcon={<CheckCircle />}
          onClick={onVerify}
          disabled={saving || hasUnsavedChanges}
          fullWidth
        >
          Duyệt
        </Button>
      </Box>

      {hasUnsavedChanges && (
        <Alert severity="warning" sx={{ mt: 1 }}>
          Có thay đổi chưa lưu
        </Alert>
      )}

      {saveError && (
        <Alert severity="error" sx={{ mt: 1 }}>
          {saveError}
        </Alert>
      )}

      {showShortcuts && (
        <Paper
          elevation={1}
          sx={{
            p: 2,
            mt: 2,
            backgroundColor: 'info.main',
            color: 'info.contrastText',
          }}
        >
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              mb: 1,
            }}
          >
            <Typography variant="subtitle2" sx={{ display: 'flex', gap: 1 }}>
              <Keyboard fontSize="small" />
              Phím tắt
            </Typography>
            <Button
              size="small"
              onClick={onDismissShortcuts}
              sx={{ color: 'inherit', minWidth: 'auto', p: 0 }}
            >
              ✕
            </Button>
          </Box>
          <Box sx={{ fontSize: '0.85rem', lineHeight: 1.6 }}>
            <Box>
              <strong>Enter:</strong> Dòng tiếp + Play
            </Box>
            <Box>
              <strong>Shift+Enter:</strong> Dòng trước
            </Box>
            <Box>
              <strong>Ctrl+Space:</strong> Pause/Play
            </Box>
            <Box>
              <strong>Ctrl+R:</strong> Replay dòng hiện tại
            </Box>
          </Box>
        </Paper>
      )}
    </>
  )
}
