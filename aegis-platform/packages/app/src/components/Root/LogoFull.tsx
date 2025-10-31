import { makeStyles } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    minHeight: theme.spacing(5),
    padding: theme.spacing(0.5, 0),
  },
  image: {
    height: 32,
    width: 'auto',
    display: 'block',
    transition: 'opacity 150ms ease',
  },
}));

const LogoFull = () => {
  const classes = useStyles();
  const theme = useTheme();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const src = isDark
    ? '/branding/aegis-logo-dark.svg'
    : '/branding/aegis-logo-light.svg';

  return (
    <span className={classes.root} aria-label="ÆGIS logo">
      <img src={src} alt="ÆGIS" className={classes.image} />
    </span>
  );
};

export default LogoFull;
