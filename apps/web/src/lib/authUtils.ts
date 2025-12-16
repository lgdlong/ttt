/**
 * Returns the appropriate redirect path based on user role.
 * @param role - The user's role string.
 * @returns The path to redirect to.
 */
export const getRedirectPathByRole = (role: string): string => {
  switch (role) {
    case 'admin':
      return '/admin'
    case 'mod':
      return '/mod'
    default:
      return '/'
  }
}
