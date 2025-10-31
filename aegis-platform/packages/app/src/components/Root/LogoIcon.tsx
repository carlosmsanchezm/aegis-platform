import { makeStyles, useTheme } from '@material-ui/core';
import { alpha } from '@material-ui/core/styles/colorManipulator';

const useStyles = makeStyles({
  glyph: {
    width: 32,
    height: 32,
  },
});

const LogoIcon = () => {
  const classes = useStyles();
  const theme = useTheme();
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const stroke = theme.palette.text.primary;
  const fill = theme.palette.text.primary;
  const accent = isDark
    ? alpha(theme.palette.text.primary, 0.12)
    : alpha(theme.palette.text.primary, 0.18);

  return (
    <svg
      className={classes.glyph}
      viewBox="0 0 40 40"
      role="presentation"
      aria-hidden
      focusable="false"
    >
      <path
        d="M20 5.5l12.5 21.65H7.5L20 5.5z"
        fill={accent}
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

export default LogoIcon;
