import { makeStyles } from '@material-ui/core';
import { useTheme } from '@material-ui/core/styles';

const useStyles = makeStyles(() => ({
  icon: {
    height: 28,
    width: 28,
    display: 'block',
  },
}));

const LogoIcon = () => {
  const classes = useStyles();
  const theme = useTheme();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';

  const gradientId = isDark ? 'aegisMarkDark' : 'aegisMarkLight';
  const outerGradient = isDark
    ? ['#22D3EE', '#9B87FF']
    : ['#4338CA', '#4F46E5'];
  const innerFill = isDark ? '#0F172A' : '#FFFFFF';

  return (
    <svg
      className={classes.icon}
      viewBox="0 0 48 48"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label="ÆGIS mark"
    >
      <defs>
        <linearGradient id={gradientId} x1="8" y1="6" x2="40" y2="42" gradientUnits="userSpaceOnUse">
          <stop stopColor={outerGradient[0]} />
          <stop offset="1" stopColor={outerGradient[1]} />
        </linearGradient>
      </defs>
      <path d="M24 5L40.5 39H7.5L24 5Z" fill={`url(#${gradientId})`} />
      <path d="M24 15L33.5 34H14.5L24 15Z" fill={innerFill} />
    </svg>
  );
};

export default LogoIcon;
