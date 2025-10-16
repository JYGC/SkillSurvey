import { type AuthModel } from 'pocketbase';

export interface IBackendClient {
  get isTokenValid(): boolean;
  get token(): string;
  loginAsync: (email: string, password: string) => Promise<string>;
  logoutAsync: () => void;
  authRefresh: () => void;
  get loggedInUser(): AuthModel;
}