import React from 'react'
import { ThemeProvider, CssBaseline } from '@mui/material'
import { BrowserRouter } from 'react-router-dom'
import theme from '~/theme'
import QueryProvider from '~/providers/QueryProvider'
import { AuthProvider } from '~/providers/AuthProvider'
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
        <BrowserRouter>
          <AuthProvider>
            <AppRouter />
          </AuthProvider>
        </BrowserRouter>
      </ThemeProvider>
    </QueryProvider>
  )
}

export default App
