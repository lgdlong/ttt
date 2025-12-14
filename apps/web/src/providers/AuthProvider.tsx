import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import type { UserResponse, LoginRequest, SignupRequest } from '~/types/user'
import {
  getCurrentUser,
  getMe,
  login as apiLogin,
  signup as apiSignup,
  logout as apiLogout,
  loginWithGoogle,
  getRedirectPathByRole,
} from '~/api/authApi'
import { useNavigate } from 'react-router-dom'

interface AuthContextType {
  user: UserResponse | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (req: LoginRequest) => Promise<void>
  signup: (req: SignupRequest) => Promise<void>
  logout: () => Promise<void>
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

  // Refresh user data from server on mount
  useEffect(() => {
    const initAuth = async () => {
      try {
        // If we have user in localStorage, validate with server
        if (user) {
          const serverUser = await getMe()
          setUser(serverUser)
        }
      } catch {
        // Token invalid or expired, clear local state
        localStorage.removeItem('user')
        setUser(null)
      } finally {
        setIsLoading(false)
      }
    }

    initAuth()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const login = useCallback(
    async (req: LoginRequest) => {
      const response = await apiLogin(req)
      setUser(response.user)
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
      // New users always go to home
      navigate('/')
    },
    [navigate]
  )

  const logout = useCallback(async () => {
    await apiLogout()
    setUser(null)
    // apiLogout already redirects to /login
  }, [])

  const handleLoginWithGoogle = useCallback(async () => {
    await loginWithGoogle()
    // This redirects to Google, callback will handle the rest
  }, [])

  const refreshUser = useCallback(async () => {
    try {
      const serverUser = await getMe()
      setUser(serverUser)
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
