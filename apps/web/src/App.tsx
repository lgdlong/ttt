import React from 'react'
import { ThemeProvider, CssBaseline } from '@mui/material'
import theme from '~/theme'
import QueryProvider from '~/providers/QueryProvider'
import AppRouter from '~/router'

/**
 * App Component
 * Root component that sets up providers and routing
 */
const App: React.FC = () => {
  return (
    <QueryProvider>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <AppRouter />
      </ThemeProvider>
    </QueryProvider>
  )
}

export default App
