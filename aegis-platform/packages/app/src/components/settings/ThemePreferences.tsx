import { useEffect, useMemo, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Grid,
  Typography,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import NightlightIcon from '@material-ui/icons/NightsStay';
import WbSunnyIcon from '@material-ui/icons/WbSunny';
import { appThemeApiRef, useApi } from '@backstage/core-plugin-api';

const useStyles = makeStyles(theme => ({
  card: {
    position: 'relative',
    borderRadius: theme.shape.borderRadius,
    border: `1px solid ${theme.palette.divider}`,
    backgroundColor: theme.palette.background.paper,
    transition: 'border-color 200ms ease, box-shadow 200ms ease, transform 200ms ease',
  },
  selected: {
    borderColor: theme.palette.primary.main,
    boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
    transform: 'translateY(-2px)',
  },
  content: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    gap: theme.spacing(1.5),
    minHeight: 180,
  },
  iconWrap: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: 44,
    height: 44,
    borderRadius: theme.shape.borderRadius,
    background:
      theme.palette.type === 'dark'
        ? 'rgba(155,135,255,0.14)'
        : 'rgba(79,70,229,0.14)',
    color: theme.palette.primary.main,
  },
  title: {
    fontWeight: 600,
  },
  description: {
    color: theme.palette.text.secondary,
    lineHeight: 1.6,
  },
  check: {
    position: 'absolute',
    top: theme.spacing(2),
    right: theme.spacing(2),
    color: theme.palette.primary.main,
  },
  actions: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: theme.spacing(2),
    marginTop: theme.spacing(3),
  },
}));

export const ThemePreferences = () => {
  const classes = useStyles();
  const themeApi = useApi(appThemeApiRef);
  const [activeThemeId, setActiveThemeId] = useState<string | undefined>(
    themeApi.getActiveThemeId(),
  );

  useEffect(() => {
    const subscription = themeApi.activeThemeId$().subscribe(themeId => {
      setActiveThemeId(themeId);
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

  const effectiveThemeId = activeThemeId ?? fallbackThemeId;

  const handleSelect = (themeId?: string) => {
    themeApi.setActiveThemeId(themeId);
  };

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Theme preferences
      </Typography>
      <Typography variant="body1" color="textSecondary" paragraph>
        Switch between ÆGIS Nightfall and ÆGIS Dawn, or follow your system preference.
      </Typography>
      <Grid container spacing={3}>
        {darkTheme ? (
          <Grid item xs={12} md={6}>
            <Card
              className={`${classes.card} ${
                effectiveThemeId === darkTheme.id ? classes.selected : ''
              }`}
              elevation={0}
            >
              <CardActionArea onClick={() => handleSelect(darkTheme.id)}>
                <CardContent className={classes.content}>
                  <span className={classes.iconWrap}>
                    <NightlightIcon fontSize="large" />
                  </span>
                  <Typography variant="h6" className={classes.title}>
                    {darkTheme.title}
                  </Typography>
                  <Typography variant="body2" className={classes.description}>
                    Immerse the console in a high-contrast, low-glare palette ideal for control room work.
                  </Typography>
                </CardContent>
              </CardActionArea>
              {effectiveThemeId === darkTheme.id ? (
                <CheckCircleIcon className={classes.check} />
              ) : null}
            </Card>
          </Grid>
        ) : null}
        {lightTheme ? (
          <Grid item xs={12} md={6}>
            <Card
              className={`${classes.card} ${
                effectiveThemeId === lightTheme.id ? classes.selected : ''
              }`}
              elevation={0}
            >
              <CardActionArea onClick={() => handleSelect(lightTheme.id)}>
                <CardContent className={classes.content}>
                  <span className={classes.iconWrap}>
                    <WbSunnyIcon fontSize="large" />
                  </span>
                  <Typography variant="h6" className={classes.title}>
                    {lightTheme.title}
                  </Typography>
                  <Typography variant="body2" className={classes.description}>
                    A bright, clean mode with softened whites for daylight and collaboration sessions.
                  </Typography>
                </CardContent>
              </CardActionArea>
              {effectiveThemeId === lightTheme.id ? (
                <CheckCircleIcon className={classes.check} />
              ) : null}
            </Card>
          </Grid>
        ) : null}
      </Grid>
      <div className={classes.actions}>
        <Typography variant="body2" color="textSecondary">
          Current selection:{' '}
          <strong>{
            effectiveThemeId === darkTheme?.id
              ? darkTheme?.title
              : effectiveThemeId === lightTheme?.id
              ? lightTheme?.title
              : 'System default'
          }</strong>
        </Typography>
        <Button onClick={() => handleSelect(undefined)} variant="outlined">
          Use system preference
        </Button>
      </div>
    </Box>
  );
};
