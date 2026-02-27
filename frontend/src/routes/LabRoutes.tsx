import React, { Suspense } from 'react';
import { Routes, Route } from 'react-router-dom';
import { useUser } from '../contexts/UserContext';
import { PERMISSIONS } from '../constants/permissions';
import { NotFoundPage } from '../pages/NotFoundPage';
import { RouteErrorBoundary } from '../components/RouteErrorBoundary';
import { LoadingState } from '../components/ui/LoadingState';

const LabPage = React.lazy(() => import('../pages/LabPage').then((m) => ({ default: m.LabPage })));
const ModelDetailPage = React.lazy(() =>
  import('../pages/Lab/ModelDetailPage').then((m) => ({ default: m.ModelDetailPage })),
);
const EntryDetailPage = React.lazy(() =>
  import('../pages/Lab/EntryDetailPage').then((m) => ({ default: m.EntryDetailPage })),
);
const EvaluationDetailPage = React.lazy(() =>
  import('../pages/Lab/EvaluationDetailPage').then((m) => ({ default: m.EvaluationDetailPage })),
);
const EntryProfilePage = React.lazy(() =>
  import('../pages/Lab/EntryProfilePage').then((m) => ({ default: m.EntryProfilePage })),
);

export default function LabRoutes() {
  const { user, hasPermission, permissionsLoading } = useUser();

  if (!user || permissionsLoading) {
    if (!user && !permissionsLoading) {
      return <NotFoundPage />;
    }
    return <LoadingState />;
  }

  if (!hasPermission(PERMISSIONS.LAB_READ)) {
    return <NotFoundPage />;
  }

  return (
    <Suspense fallback={<LoadingState />}>
      <Routes>
        <Route
          path="/"
          element={
            <RouteErrorBoundary>
              <LabPage />
            </RouteErrorBoundary>
          }
        />
        <Route
          path="/models/:modelId"
          element={
            <RouteErrorBoundary>
              <ModelDetailPage />
            </RouteErrorBoundary>
          }
        />
        <Route
          path="/models/:modelId/calcutta/:calcuttaId"
          element={
            <RouteErrorBoundary>
              <EntryDetailPage />
            </RouteErrorBoundary>
          }
        />
        <Route
          path="/models/:modelId/calcutta/:calcuttaId/evaluations/:evaluationId"
          element={
            <RouteErrorBoundary>
              <EvaluationDetailPage />
            </RouteErrorBoundary>
          }
        />
        <Route
          path="/models/:modelId/calcutta/:calcuttaId/entry-results/:entryResultId"
          element={
            <RouteErrorBoundary>
              <EntryProfilePage />
            </RouteErrorBoundary>
          }
        />
        <Route
          path="*"
          element={
            <RouteErrorBoundary>
              <NotFoundPage />
            </RouteErrorBoundary>
          }
        />
      </Routes>
    </Suspense>
  );
}
