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
