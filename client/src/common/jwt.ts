import { type OrganizationRole } from './types'

export type UserClaims = {
  _id: string
  organizationId: string
  organizationRole: OrganizationRole
}

export function decodeJwtPayload<T = unknown>(token: string): T | null {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return null
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const pad = payload.length % 4
    const padded = pad ? payload + '='.repeat(4 - pad) : payload
    const raw = atob(padded)
    const json = decodeURIComponent(
      Array.prototype.map
        .call(
          raw,
          (c: string) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2),
        )
        .join(''),
    )
    return JSON.parse(json) as T
  } catch {
    return null
  }
}

export function getUserClaims(token: string): UserClaims | null {
  return decodeJwtPayload<UserClaims>(token)
}
