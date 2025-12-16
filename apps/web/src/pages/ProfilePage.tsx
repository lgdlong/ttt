import React, { useEffect, useState } from 'react'
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  CircularProgress,
  Snackbar,
  Alert,
} from '@mui/material'
import { useForm, Controller, type SubmitHandler } from 'react-hook-form'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMe, updateMe, type UpdateMePayload } from '~/api/authApi'

// Form values type
interface IFormInput {
  email: string
  full_name: string
}

const ProfilePage: React.FC = () => {
  const queryClient = useQueryClient()
  const [snackbar, setSnackbar] = useState<{
    open: boolean
    message: string
    severity: 'success' | 'error'
  } | null>(null)

  // Data fetching using react-query
  const {
    data: user,
    isLoading: isLoadingUser,
    isError,
  } = useQuery({
    queryKey: ['me'],
    queryFn: getMe,
  })

  // Form management using react-hook-form
  const {
    control,
    handleSubmit,
    reset,
    formState: { isDirty, errors, isSubmitting },
  } = useForm<IFormInput>({
    defaultValues: {
      email: '',
      full_name: '',
    },
  })

  // Populate form with fetched user data
  useEffect(() => {
    if (user) {
      reset({
        email: user.email,
        full_name: user.full_name,
      })
    }
  }, [user, reset])

  // Data mutation using react-query
  const { mutate: updateUser } = useMutation({
    mutationFn: updateMe,
    onSuccess: (updatedUser) => {
      // Invalidate and refetch 'me' query to get fresh data
      queryClient.setQueryData(['me'], updatedUser)
      setSnackbar({ open: true, message: 'Hồ sơ đã được cập nhật thành công!', severity: 'success' })
    },
    onError: (error) => {
      setSnackbar({ open: true, message: error.message, severity: 'error' })
    },
  })

  const onSubmit: SubmitHandler<IFormInput> = (data) => {
    const payload: UpdateMePayload = {}
    if (data.full_name !== user?.full_name) {
      payload.full_name = data.full_name
    }
    if (data.email !== user?.email) {
      payload.email = data.email
    }
    updateUser(payload)
  }

  const handleCloseSnackbar = () => {
    setSnackbar(null)
  }

  if (isLoadingUser) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (isError) {
    return <Typography color="error">Lỗi khi tải thông tin người dùng.</Typography>
  }

  return (
    <Box sx={{ p: 3, maxWidth: 700, mx: 'auto' }}>
      <Typography variant="h4" component="h1" gutterBottom fontWeight={600}>
        Hồ sơ cá nhân
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Quản lý thông tin cá nhân của bạn.
      </Typography>

      <Paper component="form" onSubmit={handleSubmit(onSubmit)} elevation={0} sx={{ p: 3, border: '1px solid', borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          <Controller
            name="full_name"
            control={control}
            rules={{
              maxLength: { value: 100, message: 'Họ và tên không được vượt quá 100 ký tự' },
            }}
            render={({ field }) => (
              <TextField
                {...field}
                label="Họ và tên"
                fullWidth
                error={!!errors.full_name}
                helperText={errors.full_name?.message}
              />
            )}
          />

          <Controller
            name="email"
            control={control}
            rules={{
              required: 'Email là bắt buộc',
              pattern: {
                value: /^\S+@\S+\.\S+$/,
                message: 'Địa chỉ email không hợp lệ',
              },
            }}
            render={({ field }) => (
              <TextField
                {...field}
                label="Email"
                type="email"
                fullWidth
                error={!!errors.email}
                helperText={errors.email?.message}
              />
            )}
          />
          
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
            <Button
              type="submit"
              variant="contained"
              disabled={!isDirty || isSubmitting}
              startIcon={isSubmitting ? <CircularProgress size={20} color="inherit" /> : null}
            >
              {isSubmitting ? 'Đang lưu...' : 'Lưu thay đổi'}
            </Button>
          </Box>
        </Box>
      </Paper>

      {snackbar && (
        <Snackbar
          open={snackbar.open}
          autoHideDuration={6000}
          onClose={handleCloseSnackbar}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
          <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
            {snackbar.message}
          </Alert>
        </Snackbar>
      )}
    </Box>
  )
}

export default ProfilePage
