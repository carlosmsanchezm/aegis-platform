import { ReactNode } from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BehaviorSubject } from 'rxjs';
import { TestApiProvider } from '@backstage/test-utils';
import { appThemeApiRef } from '@backstage/core-plugin-api';
import { ThemeToggleButton } from './ThemeToggleButton';

describe('ThemeToggleButton', () => {
  const originalMatchMedia = window.matchMedia;

  beforeEach(() => {
    window.matchMedia = jest.fn().mockImplementation(() => ({
      matches: false,
      media: '',
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    }));
  });

  afterEach(() => {
    window.matchMedia = originalMatchMedia;
  });

  it('toggles between dark and light themes', async () => {
    const subject = new BehaviorSubject<string | undefined>(undefined);
    let currentTheme: string | undefined;

    const installedThemes = [
      {
        id: 'aegis-dark',
        title: 'ÆGIS Nightfall',
        variant: 'dark' as const,
        Provider: ({ children }: { children: ReactNode }) => <>{children}</>,
      },
      {
        id: 'aegis-light',
        title: 'ÆGIS Dawn',
        variant: 'light' as const,
        Provider: ({ children }: { children: ReactNode }) => <>{children}</>,
      },
    ];

    const setActiveThemeId = jest.fn((themeId?: string) => {
      currentTheme = themeId;
      subject.next(themeId);
    });

    const appThemeApi = {
      getInstalledThemes: () => installedThemes,
      activeThemeId$: () => subject,
      getActiveThemeId: () => currentTheme,
      setActiveThemeId,
    };

    render(
      <TestApiProvider apis={[[appThemeApiRef, appThemeApi]]}>
        <ThemeToggleButton />
      </TestApiProvider>,
    );

    await waitFor(() => {
      expect(setActiveThemeId).toHaveBeenCalledWith('aegis-dark');
    });

    const button = await screen.findByRole('button', { name: /toggle color theme/i });
    await userEvent.click(button);

    expect(setActiveThemeId).toHaveBeenLastCalledWith('aegis-light');

    await userEvent.click(button);
    expect(setActiveThemeId).toHaveBeenLastCalledWith('aegis-dark');
  });
});
