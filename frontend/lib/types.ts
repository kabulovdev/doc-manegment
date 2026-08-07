export interface User {
  id: string;
  email: string;
  display_name: string;
}

export interface AuthResponse {
  access_token: string;
  user: User;
}

export interface ApiErrorBody {
  error: string;
  code?: string;
  message?: string;
}
