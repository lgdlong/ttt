import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Box,
  Container,
  Paper,
  Typography,
  TextField,
  Button,
  Stack,
  InputAdornment,
  IconButton,
  Alert,
  CircularProgress,
  Divider,
} from '@mui/material'
import { Visibility, VisibilityOff, Google as GoogleIcon } from '@mui/icons-material'
import { useAuth } from '~/providers/AuthProvider'

/**
 * RegisterPage Component
 * Registration form with username/password and Google OAuth
 * - Clean, utility-style design
 * - Sharp corners (borderRadius: 0)
 * - Teal primary color
 */
const RegisterPage: React.FC = () => {
  const { signup, loginWithGoogle } = useAuth()
  const [fullName, setFullName] = useState('')
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [errors, setErrors] = useState<{
    fullName?: string
    username?: string
    email?: string
    password?: string
    confirmPassword?: string
  }>({})
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isGoogleLoading, setIsGoogleLoading] = useState(false)

  const validateForm = () => {
    const newErrors: {
      fullName?: string
      username?: string
      email?: string
      password?: string
      confirmPassword?: string
    } = {}

    // Full name is optional but if provided, validate length
    if (fullName && fullName.length > 100) {
      newErrors.fullName = 'Họ tên không được quá 100 ký tự'
    }

    if (!username) {
      newErrors.username = 'Username không được để trống'
    } else if (username.length < 3) {
      newErrors.username = 'Username phải có ít nhất 3 ký tự'
    }

    if (!email) {
      newErrors.email = 'Email không được để trống'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Email không hợp lệ'
    }

    if (!password) {
      newErrors.password = 'Mật khẩu không được để trống'
    } else if (password.length < 6) {
      newErrors.password = 'Mật khẩu phải có ít nhất 6 ký tự'
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = 'Vui lòng xác nhận mật khẩu'
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Mật khẩu không khớp'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    setIsLoading(true)
    setApiError(null)

    try {
      await signup({ username, email, password, full_name: fullName || undefined })
      // Navigation is handled by AuthProvider
    } catch (error) {
      setApiError(error instanceof Error ? error.message : 'Đăng ký thất bại')
    } finally {
      setIsLoading(false)
    }
  }

  const handleGoogleLogin = async () => {
    setIsGoogleLoading(true)
    setApiError(null)

    try {
      await loginWithGoogle()
      // This redirects to Google OAuth
    } catch (error) {
      setApiError(error instanceof Error ? error.message : 'Đăng nhập Google thất bại')
      setIsGoogleLoading(false)
    }
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        bgcolor: 'background.default',
        py: 4,
      }}
    >
      <Container maxWidth="sm">
        <Paper
          elevation={0}
          sx={{
            p: 4,
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          {/* Logo */}
          <Typography
            variant="h4"
            fontWeight={800}
            color="primary.main"
            textAlign="center"
            sx={{ mb: 1, cursor: 'pointer' }}
            component={Link}
            to="/"
            style={{ textDecoration: 'none' }}
          >
            TTT
          </Typography>

          <Typography variant="body1" color="text.secondary" textAlign="center" sx={{ mb: 4 }}>
            Tạo tài khoản mới
          </Typography>

          {/* Google Signup Button */}
          <Button
            variant="outlined"
            size="large"
            fullWidth
            startIcon={<GoogleIcon />}
            sx={{
              py: 1.5,
              mb: 3,
              fontWeight: 600,
              color: 'text.primary',
              borderColor: 'divider',
              '&:hover': {
                borderColor: 'primary.main',
                bgcolor: 'action.hover',
              },
            }}
            onClick={handleGoogleLogin}
            disabled={isLoading || isGoogleLoading}
          >
            {isGoogleLoading ? (
              <CircularProgress size={24} color="inherit" />
            ) : (
              'Đăng ký với Google'
            )}
          </Button>

          <Divider sx={{ mb: 3 }}>
            <Typography variant="body2" color="text.secondary">
              hoặc
            </Typography>
          </Divider>

          <form onSubmit={handleSubmit}>
            <Stack spacing={3}>
              {apiError && (
                <Alert severity="error" sx={{ borderRadius: 0 }}>
                  {apiError}
                </Alert>
              )}

              <TextField
                fullWidth
                label="Họ và tên"
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                error={!!errors.fullName}
                helperText={errors.fullName || 'Không bắt buộc'}
                autoComplete="name"
                autoFocus
                disabled={isLoading || isGoogleLoading}
              />

              <TextField
                fullWidth
                label="Username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                error={!!errors.username}
                helperText={errors.username}
                autoComplete="username"
                disabled={isLoading || isGoogleLoading}
              />

              <TextField
                fullWidth
                label="Email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                error={!!errors.email}
                helperText={errors.email}
                autoComplete="email"
                disabled={isLoading || isGoogleLoading}
              />

              <TextField
                fullWidth
                label="Mật khẩu"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                error={!!errors.password}
                helperText={errors.password}
                autoComplete="new-password"
                disabled={isLoading || isGoogleLoading}
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowPassword(!showPassword)}
                          edge="end"
                          aria-label="toggle password visibility"
                          disabled={isLoading || isGoogleLoading}
                        >
                          {showPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  },
                }}
              />

              <TextField
                fullWidth
                label="Xác nhận mật khẩu"
                type={showConfirmPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                error={!!errors.confirmPassword}
                helperText={errors.confirmPassword}
                disabled={isLoading || isGoogleLoading}
                autoComplete="new-password"
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                          edge="end"
                          aria-label="toggle confirm password visibility"
                          disabled={isLoading || isGoogleLoading}
                        >
                          {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  },
                }}
              />

              <Button
                type="submit"
                variant="contained"
                size="large"
                fullWidth
                sx={{ py: 1.5, fontWeight: 600 }}
                disabled={isLoading || isGoogleLoading}
              >
                {isLoading ? <CircularProgress size={24} color="inherit" /> : 'Đăng ký'}
              </Button>
            </Stack>
          </form>

          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Đã có tài khoản?{' '}
              <Link
                to="/login"
                style={{
                  color: '#008080',
                  textDecoration: 'none',
                  fontWeight: 600,
                }}
              >
                Đăng nhập ngay
              </Link>
            </Typography>
          </Box>
        </Paper>
      </Container>
    </Box>
  )
}

export default RegisterPage
