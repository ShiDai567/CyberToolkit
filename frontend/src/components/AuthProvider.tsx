'use client';

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import {
  AuthSession,
  AuthUser,
  clearSession,
  getClientAPIBaseURL,
  loadStoredSession,
  storeSession,
} from '@/lib/auth';

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  signIn: (session: AuthSession) => void;
  signOut: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function bootstrap() {
      const stored = loadStoredSession();
      if (!stored) {
        if (!cancelled) {
          setIsLoading(false);
        }
        return;
      }

      try {
        const response = await fetch(`${getClientAPIBaseURL()}/api/v1/admin/me`, {
          headers: {
            Authorization: `Bearer ${stored.accessToken}`,
          },
        });

        if (!response.ok) {
          throw new Error('session invalid');
        }

        const payload = (await response.json()) as { data: AuthUser };
        if (!cancelled) {
          setToken(stored.accessToken);
          setUser(payload.data);
        }
      } catch {
        clearSession();
        if (!cancelled) {
          setToken(null);
          setUser(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isLoading,
      isAuthenticated: Boolean(user && token),
      signIn: (session) => {
        storeSession(session);
        setUser(session.user);
        setToken(session.accessToken);
      },
      signOut: () => {
        clearSession();
        setUser(null);
        setToken(null);
      },
    }),
    [isLoading, token, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
