import { makeStyles, Typography, useTheme } from '@material-ui/core';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(1.5),
    color: theme.palette.text.primary,
  },
  glyph: {
    width: 34,
    height: 34,
  },
  wordmark: {
    fontWeight: 700,
    fontSize: 20,
    letterSpacing: '0.24em',
    textTransform: 'uppercase',
  },
  accent: {
    color:
      theme.palette.type === 'dark'
        ? '#F4F4F3'
        : theme.palette.primary.dark || theme.palette.primary.main,
  },
}));

const Glyph = ({ className }: { className?: string }) => {
  const theme = useTheme();
  const stroke = theme.palette.type === 'dark' ? '#F4F4F3' : '#050505';
  const fill = theme.palette.type === 'dark' ? '#F4F4F3' : '#050505';

  return (
    <svg
      className={className}
      viewBox="0 0 40 40"
      role="presentation"
      aria-hidden
      focusable="false"
    >
      <path
        d="M20 5.5l12.5 21.65H7.5L20 5.5z"
        fill="none"
        stroke={stroke}
        strokeWidth={2.4}
        strokeLinejoin="round"
      />
      <circle cx={20} cy={16.2} r={2.6} fill={fill} />
      <circle cx={11.2} cy={29.8} r={2.8} fill="none" stroke={stroke} strokeWidth={2.4} />
      <circle cx={28.8} cy={29.8} r={2.8} fill="none" stroke={stroke} strokeWidth={2.4} />
    </svg>
  );
};

const LogoFull = () => {
  const classes = useStyles();

  return (
    <div className={classes.root} aria-label="ÆGIS home">
      <Glyph className={classes.glyph} />
      <Typography
        component="span"
        className={`${classes.wordmark} ${classes.accent}`}
      >
        ÆGIS
      </Typography>
    </div>
  );
};

export default LogoFull;
