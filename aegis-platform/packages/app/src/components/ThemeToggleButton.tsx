import { useEffect, useMemo, useState } from 'react';
import { IconButton, Tooltip } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import Brightness4Icon from '@material-ui/icons/Brightness4';
import Brightness7Icon from '@material-ui/icons/Brightness7';
import { appThemeApiRef, useApi } from '@backstage/core-plugin-api';

const useStyles = makeStyles(theme => ({
  root: {
    position: 'fixed',
    top: theme.spacing(2),
    right: theme.spacing(2),
    zIndex: theme.zIndex.tooltip + 1,
    backgroundColor: theme.palette.background.paper,
    borderRadius: theme.shape.borderRadius,
    boxShadow:
      theme.palette.type === 'dark'
        ? '0 18px 40px rgba(0,0,0,0.45)'
        : '0 14px 30px rgba(15,23,42,0.18)',
    backdropFilter: 'blur(12px)',
    border: `1px solid ${theme.palette.divider}`,
    padding: theme.spacing(0.5),
  },
  button: {
    color: theme.palette.text.primary,
  },
}));

export const ThemeToggleButton = () => {
  const classes = useStyles();
  const themeApi = useApi(appThemeApiRef);
  const [activeThemeId, setActiveThemeId] = useState<string | undefined>(
    themeApi.getActiveThemeId(),
  );

  useEffect(() => {
    const subscription = themeApi.activeThemeId$().subscribe(id => {
      setActiveThemeId(id);
    });
    return () => {
      subscription?.unsubscribe?.();
    };
  }, [themeApi]);

  const installedThemes = useMemo(() => themeApi.getInstalledThemes(), [themeApi]);
  const darkTheme = installedThemes.find(theme => theme.variant === 'dark');
  const lightTheme = installedThemes.find(theme => theme.variant === 'light');

  const fallbackThemeId = useMemo(() => {
    const prefersLight =
      typeof window !== 'undefined' &&
      window.matchMedia?.('(prefers-color-scheme: light)').matches;
    if (prefersLight) {
      return lightTheme?.id ?? darkTheme?.id ?? installedThemes[0]?.id;
    }
    return darkTheme?.id ?? lightTheme?.id ?? installedThemes[0]?.id;
  }, [darkTheme, lightTheme, installedThemes]);

  useEffect(() => {
    if (!activeThemeId && fallbackThemeId) {
      themeApi.setActiveThemeId(fallbackThemeId);
    }
  }, [activeThemeId, fallbackThemeId, themeApi]);

  const effectiveThemeId = activeThemeId ?? fallbackThemeId;
  const isDark = effectiveThemeId === darkTheme?.id;

  const handleToggle = () => {
    const nextId = isDark ? lightTheme?.id : darkTheme?.id;
    if (nextId) {
      themeApi.setActiveThemeId(nextId);
    }
  };

  if (!darkTheme || !lightTheme) {
    return null;
  }

  const tooltipTitle = isDark ? `${lightTheme.title} theme` : `${darkTheme.title} theme`;

  return (
    <div className={classes.root} role="status" aria-live="polite">
      <Tooltip title={`Switch to ${tooltipTitle}`} arrow>
        <IconButton
          size="small"
          className={classes.button}
          onClick={handleToggle}
          aria-label="Toggle color theme"
        >
          {isDark ? <Brightness7Icon /> : <Brightness4Icon />}
        </IconButton>
      </Tooltip>
    </div>
  );
};
