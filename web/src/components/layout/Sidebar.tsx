import { Fragment } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Dialog, DialogPanel, Transition, TransitionChild } from '@headlessui/react';
import {
  Bars3Icon,
  XMarkIcon,
  HomeIcon,
  MapIcon,
  ServerIcon,
  ChartBarIcon,
  BellIcon,
  UsersIcon,
  Cog6ToothIcon,
  SignalIcon,
} from '@heroicons/react/24/outline';
import { cn } from '@/utils';
import { useAuth } from '@/context';

interface NavigationItem {
  name: string;
  href: string;
  icon: typeof HomeIcon;
}

const navigation: NavigationItem[] = [
  { name: 'Dashboard', href: '/', icon: HomeIcon },
  { name: 'Network Map', href: '/map', icon: MapIcon },
  { name: 'Routers', href: '/routers', icon: ServerIcon },
  { name: 'Metrics', href: '/metrics', icon: ChartBarIcon },
  { name: 'Alerts', href: '/alerts', icon: BellIcon },
  { name: 'Sessions', href: '/sessions', icon: SignalIcon },
];

const adminNavigation: NavigationItem[] = [
  { name: 'Users', href: '/admin/users', icon: UsersIcon },
  { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
];

interface SidebarProps {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
}

export function Sidebar({ sidebarOpen, setSidebarOpen }: SidebarProps) {
  const location = useLocation();
  const { user } = useAuth();

  const isActive = (href: string) => {
    if (href === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(href);
  };

  const NavLink = ({ item }: { item: NavigationItem }) => (
    <Link
      to={item.href}
      className={cn(
        'group flex items-center gap-x-3 rounded-md p-2 text-sm font-medium transition-colors',
        isActive(item.href)
          ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400'
          : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white'
      )}
      aria-current={isActive(item.href) ? 'page' : undefined}
    >
      <item.icon
        className={cn(
          'h-5 w-5 shrink-0',
          isActive(item.href)
            ? 'text-blue-600 dark:text-blue-400'
            : 'text-gray-400 group-hover:text-gray-500 dark:group-hover:text-gray-300'
        )}
        aria-hidden="true"
      />
      {item.name}
    </Link>
  );

  const SidebarContent = () => (
    <div className="flex h-full flex-col">
      <div className="flex h-16 shrink-0 items-center px-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-md bg-blue-600 flex items-center justify-center">
            <SignalIcon className="h-5 w-5 text-white" />
          </div>
          <span className="text-lg font-semibold text-gray-900 dark:text-white">
            ISP Monitor
          </span>
        </Link>
      </div>
      <nav className="flex flex-1 flex-col px-4 py-4">
        <ul role="list" className="flex flex-1 flex-col gap-y-1">
          {navigation.map((item) => (
            <li key={item.name}>
              <NavLink item={item} />
            </li>
          ))}
        </ul>
        {user?.roles?.includes('admin') && (
          <div className="mt-auto border-t border-gray-200 pt-4 dark:border-gray-700">
            <p className="mb-2 px-2 text-xs font-semibold uppercase text-gray-400">
              Admin
            </p>
            <ul role="list" className="space-y-1">
              {adminNavigation.map((item) => (
                <li key={item.name}>
                  <NavLink item={item} />
                </li>
              ))}
            </ul>
          </div>
        )}
      </nav>
    </div>
  );

  return (
    <>
      {/* Mobile sidebar */}
      <Transition show={sidebarOpen} as={Fragment}>
        <Dialog as="div" className="relative z-50 lg:hidden" onClose={setSidebarOpen}>
          <TransitionChild
            as={Fragment}
            enter="transition-opacity ease-linear duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity ease-linear duration-300"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-900/80" />
          </TransitionChild>

          <div className="fixed inset-0 flex">
            <TransitionChild
              as={Fragment}
              enter="transition ease-in-out duration-300 transform"
              enterFrom="-translate-x-full"
              enterTo="translate-x-0"
              leave="transition ease-in-out duration-300 transform"
              leaveFrom="translate-x-0"
              leaveTo="-translate-x-full"
            >
              <DialogPanel className="relative mr-16 flex w-full max-w-xs flex-1">
                <TransitionChild
                  as={Fragment}
                  enter="ease-in-out duration-300"
                  enterFrom="opacity-0"
                  enterTo="opacity-100"
                  leave="ease-in-out duration-300"
                  leaveFrom="opacity-100"
                  leaveTo="opacity-0"
                >
                  <div className="absolute left-full top-0 flex w-16 justify-center pt-5">
                    <button
                      type="button"
                      className="-m-2.5 p-2.5"
                      onClick={() => setSidebarOpen(false)}
                    >
                      <span className="sr-only">Close sidebar</span>
                      <XMarkIcon className="h-6 w-6 text-white" aria-hidden="true" />
                    </button>
                  </div>
                </TransitionChild>
                <div className="flex grow flex-col overflow-y-auto bg-white dark:bg-gray-900">
                  <SidebarContent />
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </Dialog>
      </Transition>

      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-64 lg:flex-col">
        <div className="flex grow flex-col overflow-y-auto border-r border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900">
          <SidebarContent />
        </div>
      </div>
    </>
  );
}

export function MobileMenuButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      className="-m-2.5 p-2.5 text-gray-700 dark:text-gray-200 lg:hidden"
      onClick={onClick}
    >
      <span className="sr-only">Open sidebar</span>
      <Bars3Icon className="h-6 w-6" aria-hidden="true" />
    </button>
  );
}

export default Sidebar;
