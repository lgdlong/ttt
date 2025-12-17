import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import type { UserResponse, LoginRequest, SignupRequest } from '~/types/user'
import {
  getMe,
  login as apiLogin,
  signup as apiSignup,
  logout as apiLogout,
  loginWithGoogle,
} from '~/api/authApi'
import { eventBus } from '~/lib/apiClient'
import { getRedirectPathByRole } from '~/lib/authUtils'
import { useNavigate } from 'react-router-dom'

// This function is a local utility, not from the API
const getCurrentUser = (): UserResponse | null => {
  const userJson = localStorage.getItem('user')
  if (!userJson) return null
  try {
    return JSON.parse(userJson)
  } catch {
    return null
  }
}

interface AuthContextType {
  user: UserResponse | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (req: LoginRequest) => Promise<void>
  signup: (req: SignupRequest) => Promise<void>
  logout: (options?: { navigate: boolean }) => Promise<void>
  loginWithGoogle: () => Promise<void>
  refreshUser: () => Promise<void>
  hasRole: (roles: string[]) => boolean
  isAdmin: boolean
  isModerator: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<UserResponse | null>(() => getCurrentUser())
  const [isLoading, setIsLoading] = useState(true)
  const navigate = useNavigate()

  const handleLogout = useCallback(() => {
    setUser(null)
    localStorage.removeItem('user')
    navigate('/login')
  }, [navigate])

  useEffect(() => {
    eventBus.on('auth:logout', handleLogout)
    return () => {
      eventBus.off('auth:logout', handleLogout)
    }
  }, [handleLogout])
  // Refresh user data from server on mount
  useEffect(() => {
    const initAuth = async () => {
      setIsLoading(true)
      try {
        // We always try to fetch the user on initial load
        const serverUser = await getMe()
        setUser(serverUser)
        localStorage.setItem('user', JSON.stringify(serverUser))
      } catch {
        // If getMe fails, it means no valid token, so we are logged out.
        handleLogout()
      } finally {
        setIsLoading(false)
      }
    }

    initAuth()
  }, [handleLogout])

  const login = useCallback(
    async (req: LoginRequest) => {
      const response = await apiLogin(req)
      setUser(response.user)
      localStorage.setItem('user', JSON.stringify(response.user))
      // Navigate based on role
      const redirectPath = getRedirectPathByRole(response.user.role)
      navigate(redirectPath)
    },
    [navigate]
  )

  const signup = useCallback(
    async (req: SignupRequest) => {
      const response = await apiSignup(req)
      setUser(response.user)
      localStorage.setItem('user', JSON.stringify(response.user))
      // New users always go to home
      navigate('/')
    },
    [navigate]
  )

  const logout = useCallback(
    async (options = { navigate: true }) => {
      try {
        await apiLogout()
      } catch (error) {
        console.error('Logout failed, but clearing client state anyway.', error)
      } finally {
        // Ensure client-side logout always happens
        setUser(null)
        localStorage.removeItem('user')
        if (options.navigate) {
          navigate('/login')
        }
      }
    },
    [navigate]
  )

  const handleLoginWithGoogle = useCallback(async () => {
    await loginWithGoogle()
    // This redirects to Google, callback will handle the rest
  }, [])

  const refreshUser = useCallback(async () => {
    try {
      const serverUser = await getMe()
      setUser(serverUser)
      localStorage.setItem('user', JSON.stringify(serverUser))
    } catch {
      setUser(null)
      localStorage.removeItem('user')
    }
  }, [])

  const hasRole = useCallback(
    (roles: string[]) => {
      if (!user) return false
      return roles.includes(user.role)
    },
    [user]
  )

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    signup,
    logout,
    loginWithGoogle: handleLoginWithGoogle,
    refreshUser,
    hasRole,
    isAdmin: user?.role === 'admin',
    isModerator: user?.role === 'admin' || user?.role === 'mod',
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
