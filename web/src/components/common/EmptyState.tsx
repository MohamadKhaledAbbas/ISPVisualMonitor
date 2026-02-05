import { ExclamationTriangleIcon, XCircleIcon, InformationCircleIcon } from '@heroicons/react/24/outline';
import { cn } from '@/utils';
import Button from './Button';

export interface EmptyStateProps {
  title: string;
  description?: string;
  icon?: React.ReactNode;
  action?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
}

export function EmptyState({ title, description, icon, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
      {icon && <div className="mb-4 text-gray-400 dark:text-gray-500">{icon}</div>}
      <h3 className="text-lg font-medium text-gray-900 dark:text-white">{title}</h3>
      {description && (
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{description}</p>
      )}
      {action && (
        <Button onClick={action.onClick} className="mt-4">
          {action.label}
        </Button>
      )}
    </div>
  );
}

export interface ErrorStateProps {
  title?: string;
  message?: string;
  retry?: () => void;
  className?: string;
}

export function ErrorState({
  title = 'Something went wrong',
  message = 'An error occurred while loading this page. Please try again.',
  retry,
  className,
}: ErrorStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
      <XCircleIcon className="h-12 w-12 text-red-500" />
      <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">{title}</h3>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{message}</p>
      {retry && (
        <Button onClick={retry} className="mt-4">
          Try again
        </Button>
      )}
    </div>
  );
}

export interface AlertBannerProps {
  type: 'info' | 'warning' | 'error' | 'success';
  title?: string;
  message: string;
  onClose?: () => void;
  className?: string;
}

export function AlertBanner({ type, title, message, onClose, className }: AlertBannerProps) {
  const styles = {
    info: {
      bg: 'bg-blue-50 dark:bg-blue-900/20',
      border: 'border-blue-200 dark:border-blue-800',
      icon: <InformationCircleIcon className="h-5 w-5 text-blue-400" />,
      text: 'text-blue-800 dark:text-blue-200',
    },
    warning: {
      bg: 'bg-yellow-50 dark:bg-yellow-900/20',
      border: 'border-yellow-200 dark:border-yellow-800',
      icon: <ExclamationTriangleIcon className="h-5 w-5 text-yellow-400" />,
      text: 'text-yellow-800 dark:text-yellow-200',
    },
    error: {
      bg: 'bg-red-50 dark:bg-red-900/20',
      border: 'border-red-200 dark:border-red-800',
      icon: <XCircleIcon className="h-5 w-5 text-red-400" />,
      text: 'text-red-800 dark:text-red-200',
    },
    success: {
      bg: 'bg-green-50 dark:bg-green-900/20',
      border: 'border-green-200 dark:border-green-800',
      icon: <InformationCircleIcon className="h-5 w-5 text-green-400" />,
      text: 'text-green-800 dark:text-green-200',
    },
  };

  const style = styles[type];

  return (
    <div
      className={cn(
        'rounded-md border p-4',
        style.bg,
        style.border,
        className
      )}
      role="alert"
    >
      <div className="flex">
        <div className="flex-shrink-0">{style.icon}</div>
        <div className="ml-3">
          {title && <h3 className={cn('text-sm font-medium', style.text)}>{title}</h3>}
          <p className={cn('text-sm', style.text, title && 'mt-1')}>{message}</p>
        </div>
        {onClose && (
          <div className="ml-auto pl-3">
            <button
              type="button"
              onClick={onClose}
              className={cn(
                'inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2',
                style.text,
                'hover:bg-black/5 dark:hover:bg-white/5'
              )}
            >
              <span className="sr-only">Dismiss</span>
              <XCircleIcon className="h-5 w-5" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default EmptyState;
