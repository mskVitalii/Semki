import { api } from './client'

export interface LoginDto {
  email: string
  password: string
  organization: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface AuthErrorResponse {
  message: string
  statusCode?: number
}

export const login = async (data: LoginDto): Promise<AuthResponse> => {
  const response = await api.post('/api/v1/login', data)
  return response.data
}

export interface RegisterDto {
  name: string
  email: string
  password: string
  organization: string
}

export interface RegisterUserResponse {
  message: string
  tokens: {
    access_token: string
    token_type: string
    refresh_token: string
    expires_at: number
    created_at: number
  }
}

export const register = async (
  data: RegisterDto,
): Promise<RegisterUserResponse> => {
  const response = await api.post('/api/v1/user/register', data)
  return response.data
}

export interface SetPasswordDto {
  password: string
}

export const setPassword = async (
  data: SetPasswordDto,
): Promise<RegisterUserResponse> => {
  const response = await api.post('/api/v1/user/set_password', data)
  return response.data
}

export async function requestPasswordReset(email: string) {
  const res = await api.post('/api/v1/user/reset_password', { email })
  return res.data
}
