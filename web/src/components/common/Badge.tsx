import { cn } from '@/utils';

export interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
  size?: 'sm' | 'md';
  className?: string;
}

export function Badge({ children, variant = 'default', size = 'md', className }: BadgeProps) {
  const variants = {
    default: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
    success: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    danger: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
    info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  };

  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-0.5 text-sm',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full',
        variants[variant],
        sizes[size],
        className
      )}
    >
      {children}
    </span>
  );
}

export interface StatusBadgeProps {
  status: string;
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const getVariant = (status: string): BadgeProps['variant'] => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'online':
      case 'up':
      case 'resolved':
        return 'success';
      case 'inactive':
      case 'offline':
      case 'down':
      case 'critical':
        return 'danger';
      case 'maintenance':
      case 'degraded':
      case 'warning':
      case 'acknowledged':
        return 'warning';
      case 'info':
      case 'open':
        return 'info';
      default:
        return 'default';
    }
  };

  return (
    <Badge variant={getVariant(status)} className={className}>
      {status}
    </Badge>
  );
}

export default Badge;
