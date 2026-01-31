import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AppShell } from '@/components/layout';
import { ProtectedRoute } from '@/components/auth';
import { PageLoading } from '@/components/common';

// Lazy load pages for code splitting
const DashboardPage = lazy(() => import('@/pages/DashboardPage'));
const MapPage = lazy(() => import('@/pages/MapPage'));
const RoutersPage = lazy(() => import('@/pages/RoutersPage'));
const AlertsPage = lazy(() => import('@/pages/AlertsPage'));
const SessionsPage = lazy(() => import('@/pages/SessionsPage'));
const MetricsPage = lazy(() => import('@/pages/MetricsPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
const LoginPage = lazy(() => import('@/components/auth/LoginForm'));

function SuspenseWrapper({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<PageLoading />}>{children}</Suspense>;
}

export function AppRoutes() {
  return (
    <Routes>
      {/* Public routes */}
      <Route
        path="/login"
        element={
          <SuspenseWrapper>
            <LoginPage />
          </SuspenseWrapper>
        }
      />

      {/* Protected routes */}
      <Route
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        <Route
          path="/"
          element={
            <SuspenseWrapper>
              <DashboardPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/map"
          element={
            <SuspenseWrapper>
              <MapPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/routers"
          element={
            <SuspenseWrapper>
              <RoutersPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/routers/:id"
          element={
            <SuspenseWrapper>
              <RoutersPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/metrics"
          element={
            <SuspenseWrapper>
              <MetricsPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/alerts"
          element={
            <SuspenseWrapper>
              <AlertsPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/sessions"
          element={
            <SuspenseWrapper>
              <SessionsPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/settings"
          element={
            <SuspenseWrapper>
              <SettingsPage />
            </SuspenseWrapper>
          }
        />
        <Route
          path="/settings/*"
          element={
            <SuspenseWrapper>
              <SettingsPage />
            </SuspenseWrapper>
          }
        />
      </Route>

      {/* Fallback route */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default AppRoutes;
