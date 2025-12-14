import { createTheme } from '@mui/material/styles'

/**
 * TTT Archive Design System
 * Based on UI_Spec.md - Clean Utility Style
 * Font: Inter
 */
const theme = createTheme({
  palette: {
    primary: {
      main: '#008080', // Teal/Cyan đậm
      light: '#E0F2F1', // Highlight color
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#10B981', // Badge Uy tín - Emerald
    },
    background: {
      default: '#F8FAFC', // Nền phụ/Nền trang chủ
      paper: '#FFFFFF', // Nền chính
    },
    text: {
      primary: '#1E293B', // Đen xám - Tiêu đề
      secondary: '#475569', // Xám vừa - Nội dung
      disabled: '#94A3B8', // Xám nhạt - Metadata
    },
    divider: '#E2E8F0',
  },
  typography: {
    fontFamily: '"Inter", sans-serif',
    h1: {
      fontWeight: 700,
    },
    h2: {
      fontWeight: 600,
    },
    h3: {
      fontWeight: 600,
    },
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 700,
    },
    h6: {
      fontWeight: 600,
    },
    body1: {
      fontWeight: 400,
      lineHeight: 1.6,
    },
    body2: {
      fontWeight: 400,
      lineHeight: 1.6,
    },
    subtitle1: {
      fontWeight: 600,
      lineHeight: 1.3,
    },
  },
  shape: {
    borderRadius: 0,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 500,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 500,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: 'none',
          border: '1px solid #E2E8F0',
        },
      },
    },
  },
})

export default theme
