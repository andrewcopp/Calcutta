export interface User {
  id: string;
  email?: string;
  firstName: string;
  lastName: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  firstName: string;
  lastName: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  accessToken: string;
}

export interface InvitePreview {
  firstName: string;
  calcuttaName: string;
  commissionerName: string;
  tournamentStartingAt?: string;
}