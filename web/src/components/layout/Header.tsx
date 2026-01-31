import { Fragment } from 'react';
import { Menu, MenuButton, MenuItem, MenuItems, Transition } from '@headlessui/react';
import {
  BellIcon,
  UserCircleIcon,
  SunIcon,
  MoonIcon,
  ArrowRightOnRectangleIcon,
} from '@heroicons/react/24/outline';
import { useAuth, useTheme } from '@/context';
import { cn } from '@/utils';
import { MobileMenuButton } from './Sidebar';

interface HeaderProps {
  onMenuClick: () => void;
}

export function Header({ onMenuClick }: HeaderProps) {
  const { user, logout } = useAuth();
  const { resolvedTheme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(resolvedTheme === 'dark' ? 'light' : 'dark');
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 bg-white px-4 shadow-sm dark:border-gray-700 dark:bg-gray-900 sm:gap-x-6 sm:px-6 lg:px-8">
      <MobileMenuButton onClick={onMenuClick} />

      {/* Separator */}
      <div className="h-6 w-px bg-gray-200 dark:bg-gray-700 lg:hidden" aria-hidden="true" />

      <div className="flex flex-1 gap-x-4 self-stretch lg:gap-x-6">
        <div className="flex flex-1" />
        
        <div className="flex items-center gap-x-4 lg:gap-x-6">
          {/* Theme toggle */}
          <button
            type="button"
            onClick={toggleTheme}
            className="-m-2.5 p-2.5 text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
            aria-label={`Switch to ${resolvedTheme === 'dark' ? 'light' : 'dark'} mode`}
          >
            {resolvedTheme === 'dark' ? (
              <SunIcon className="h-6 w-6" aria-hidden="true" />
            ) : (
              <MoonIcon className="h-6 w-6" aria-hidden="true" />
            )}
          </button>

          {/* Notifications */}
          <button
            type="button"
            className="-m-2.5 p-2.5 text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
            aria-label="View notifications"
          >
            <BellIcon className="h-6 w-6" aria-hidden="true" />
          </button>

          {/* Separator */}
          <div
            className="hidden lg:block lg:h-6 lg:w-px lg:bg-gray-200 dark:lg:bg-gray-700"
            aria-hidden="true"
          />

          {/* Profile dropdown */}
          <Menu as="div" className="relative">
            <MenuButton className="-m-1.5 flex items-center p-1.5">
              <span className="sr-only">Open user menu</span>
              <UserCircleIcon className="h-8 w-8 text-gray-400" aria-hidden="true" />
              <span className="hidden lg:flex lg:items-center">
                <span
                  className="ml-4 text-sm font-medium text-gray-900 dark:text-white"
                  aria-hidden="true"
                >
                  {user?.first_name} {user?.last_name}
                </span>
              </span>
            </MenuButton>
            <Transition
              as={Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <MenuItems className="absolute right-0 z-10 mt-2.5 w-56 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none dark:bg-gray-800 dark:ring-gray-700">
                <div className="px-3 py-2 border-b border-gray-100 dark:border-gray-700">
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {user?.first_name} {user?.last_name}
                  </p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                    {user?.email}
                  </p>
                </div>
                <MenuItem>
                  {({ active }) => (
                    <a
                      href="/settings/profile"
                      className={cn(
                        active ? 'bg-gray-50 dark:bg-gray-700' : '',
                        'block px-3 py-2 text-sm text-gray-700 dark:text-gray-200'
                      )}
                    >
                      Your Profile
                    </a>
                  )}
                </MenuItem>
                <MenuItem>
                  {({ active }) => (
                    <a
                      href="/settings"
                      className={cn(
                        active ? 'bg-gray-50 dark:bg-gray-700' : '',
                        'block px-3 py-2 text-sm text-gray-700 dark:text-gray-200'
                      )}
                    >
                      Settings
                    </a>
                  )}
                </MenuItem>
                <MenuItem>
                  {({ active }) => (
                    <button
                      onClick={handleLogout}
                      className={cn(
                        active ? 'bg-gray-50 dark:bg-gray-700' : '',
                        'flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 dark:text-gray-200'
                      )}
                    >
                      <ArrowRightOnRectangleIcon className="h-4 w-4" />
                      Sign out
                    </button>
                  )}
                </MenuItem>
              </MenuItems>
            </Transition>
          </Menu>
        </div>
      </div>
    </div>
  );
}

export default Header;
