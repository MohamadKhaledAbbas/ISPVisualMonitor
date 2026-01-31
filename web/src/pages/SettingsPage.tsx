import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth, useTheme } from '@/context';
import {
  Card,
  Button,
  Input,
  AlertBanner,
} from '@/components/common';
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react';
import { cn } from '@/utils';
import {
  UserCircleIcon,
  Cog6ToothIcon,
  BellIcon,
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
} from '@heroicons/react/24/outline';

const profileSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email address'),
});

type ProfileFormData = z.infer<typeof profileSchema>;

const passwordSchema = z
  .object({
    current_password: z.string().min(1, 'Current password is required'),
    new_password: z.string().min(8, 'Password must be at least 8 characters'),
    confirm_password: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Passwords don't match",
    path: ['confirm_password'],
  });

type PasswordFormData = z.infer<typeof passwordSchema>;

export function SettingsPage() {
  const { user } = useAuth();
  const { theme, setTheme } = useTheme();
  const [profileMessage, setProfileMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [passwordMessage, setPasswordMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const {
    register: registerProfile,
    handleSubmit: handleProfileSubmit,
    formState: { errors: profileErrors, isSubmitting: isProfileSubmitting },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: user?.first_name || '',
      last_name: user?.last_name || '',
      email: user?.email || '',
    },
  });

  const {
    register: registerPassword,
    handleSubmit: handlePasswordSubmit,
    reset: resetPassword,
    formState: { errors: passwordErrors, isSubmitting: isPasswordSubmitting },
  } = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  const onProfileSubmit = async (_data: ProfileFormData) => {
    try {
      // API call would go here
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setProfileMessage({ type: 'success', text: 'Profile updated successfully' });
    } catch {
      setProfileMessage({ type: 'error', text: 'Failed to update profile' });
    }
  };

  const onPasswordSubmit = async (_data: PasswordFormData) => {
    try {
      // API call would go here
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setPasswordMessage({ type: 'success', text: 'Password changed successfully' });
      resetPassword();
    } catch {
      setPasswordMessage({ type: 'error', text: 'Failed to change password' });
    }
  };

  const themeOptions = [
    { value: 'light', label: 'Light', icon: SunIcon },
    { value: 'dark', label: 'Dark', icon: MoonIcon },
    { value: 'system', label: 'System', icon: ComputerDesktopIcon },
  ];

  const tabs = [
    { name: 'Profile', icon: UserCircleIcon },
    { name: 'Preferences', icon: Cog6ToothIcon },
    { name: 'Notifications', icon: BellIcon },
  ];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Settings</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Manage your account and preferences
        </p>
      </div>

      <TabGroup>
        <div className="border-b border-gray-200 dark:border-gray-700">
          <TabList className="-mb-px flex space-x-8">
            {tabs.map((tab) => (
              <Tab
                key={tab.name}
                className={({ selected }) =>
                  cn(
                    'flex items-center gap-2 border-b-2 py-4 px-1 text-sm font-medium transition-colors focus:outline-none',
                    selected
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                  )
                }
              >
                <tab.icon className="h-5 w-5" />
                {tab.name}
              </Tab>
            ))}
          </TabList>
        </div>

        <TabPanels className="mt-6">
          {/* Profile Tab */}
          <TabPanel className="space-y-6">
            {/* Profile form */}
            <Card title="Profile Information" description="Update your account information">
              {profileMessage && (
                <AlertBanner
                  type={profileMessage.type}
                  message={profileMessage.text}
                  onClose={() => setProfileMessage(null)}
                  className="mb-4"
                />
              )}
              <form onSubmit={handleProfileSubmit(onProfileSubmit)} className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <Input
                    label="First Name"
                    {...registerProfile('first_name')}
                    error={profileErrors.first_name?.message}
                  />
                  <Input
                    label="Last Name"
                    {...registerProfile('last_name')}
                    error={profileErrors.last_name?.message}
                  />
                </div>
                <Input
                  label="Email Address"
                  type="email"
                  {...registerProfile('email')}
                  error={profileErrors.email?.message}
                  disabled
                  helperText="Email cannot be changed"
                />
                <div className="flex justify-end">
                  <Button type="submit" isLoading={isProfileSubmitting}>
                    Save Changes
                  </Button>
                </div>
              </form>
            </Card>

            {/* Password form */}
            <Card title="Change Password" description="Update your password">
              {passwordMessage && (
                <AlertBanner
                  type={passwordMessage.type}
                  message={passwordMessage.text}
                  onClose={() => setPasswordMessage(null)}
                  className="mb-4"
                />
              )}
              <form onSubmit={handlePasswordSubmit(onPasswordSubmit)} className="space-y-4">
                <Input
                  label="Current Password"
                  type="password"
                  {...registerPassword('current_password')}
                  error={passwordErrors.current_password?.message}
                />
                <Input
                  label="New Password"
                  type="password"
                  {...registerPassword('new_password')}
                  error={passwordErrors.new_password?.message}
                />
                <Input
                  label="Confirm New Password"
                  type="password"
                  {...registerPassword('confirm_password')}
                  error={passwordErrors.confirm_password?.message}
                />
                <div className="flex justify-end">
                  <Button type="submit" isLoading={isPasswordSubmitting}>
                    Change Password
                  </Button>
                </div>
              </form>
            </Card>
          </TabPanel>

          {/* Preferences Tab */}
          <TabPanel className="space-y-6">
            <Card title="Appearance" description="Customize the look and feel">
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Theme
                  </label>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    Select your preferred color scheme
                  </p>
                  <div className="mt-3 flex gap-3">
                    {themeOptions.map((option) => (
                      <button
                        key={option.value}
                        onClick={() => setTheme(option.value as 'light' | 'dark' | 'system')}
                        className={cn(
                          'flex flex-col items-center gap-2 rounded-lg border p-4 transition-colors',
                          theme === option.value
                            ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                            : 'border-gray-200 hover:border-gray-300 dark:border-gray-700 dark:hover:border-gray-600'
                        )}
                      >
                        <option.icon className="h-6 w-6 text-gray-600 dark:text-gray-400" />
                        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                          {option.label}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </Card>

            <Card title="Dashboard" description="Configure dashboard settings">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Auto-refresh
                    </p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      Automatically refresh dashboard data
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    defaultChecked
                    className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Refresh Interval
                    </p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      How often to refresh data
                    </p>
                  </div>
                  <select className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm dark:border-gray-600 dark:bg-gray-800">
                    <option value="30">30 seconds</option>
                    <option value="60">1 minute</option>
                    <option value="300">5 minutes</option>
                  </select>
                </div>
              </div>
            </Card>
          </TabPanel>

          {/* Notifications Tab */}
          <TabPanel className="space-y-6">
            <Card title="Email Notifications" description="Manage email notification preferences">
              <div className="space-y-4">
                {[
                  { label: 'Critical alerts', description: 'Receive emails for critical alerts', default: true },
                  { label: 'Warning alerts', description: 'Receive emails for warning alerts', default: true },
                  { label: 'Info alerts', description: 'Receive emails for informational alerts', default: false },
                  { label: 'Daily summary', description: 'Receive a daily summary of network status', default: true },
                  { label: 'Weekly report', description: 'Receive a weekly report with metrics and trends', default: false },
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                        {item.label}
                      </p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{item.description}</p>
                    </div>
                    <input
                      type="checkbox"
                      defaultChecked={item.default}
                      className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                  </div>
                ))}
              </div>
            </Card>

            <Card title="Browser Notifications" description="Manage browser push notifications">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Enable push notifications
                    </p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      Receive real-time alerts in your browser
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                </div>
              </div>
            </Card>
          </TabPanel>
        </TabPanels>
      </TabGroup>
    </div>
  );
}

export default SettingsPage;
