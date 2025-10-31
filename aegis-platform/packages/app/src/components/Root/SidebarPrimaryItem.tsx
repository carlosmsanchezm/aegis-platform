import { ComponentType, ReactNode } from 'react';
import { makeStyles } from '@material-ui/core';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import { NavLink } from 'react-router-dom';
import { useSidebarOpenState } from '@backstage/core-components';

type SidebarPrimaryItemProps = {
  icon: ComponentType<{ className?: string }>;
  text: string;
  to: string;
  secondaryAction?: ReactNode;
};

const useStyles = makeStyles(theme => {
  const hoverBackground =
    theme.palette.navigation.navItem?.hoverBackground ??
    alpha(theme.palette.text.primary, 0.08);

  return {
    link: {
      display: 'flex',
      alignItems: 'center',
      gap: theme.spacing(2),
      padding: theme.spacing(1.5, 2),
      borderRadius: theme.shape.borderRadius,
      color: theme.palette.navigation.color,
      textDecoration: 'none',
      minHeight: 56,
      fontWeight: 600,
      letterSpacing: '0.08em',
      textTransform: 'uppercase',
      transition: 'background-color 150ms ease, color 150ms ease, box-shadow 150ms ease',
      boxShadow: 'none',
    },
    collapsed: {
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      gap: theme.spacing(1),
      textAlign: 'center',
      padding: theme.spacing(1.25, 1),
      minHeight: 86,
      letterSpacing: '0.12em',
    },
    active: {
      color: theme.palette.navigation.selectedColor,
      backgroundColor: hoverBackground,
      boxShadow: `inset 4px 0 0 ${theme.palette.navigation.indicator}`,
    },
    icon: {
      fontSize: '1.4rem',
    },
    label: {
      flex: '1 1 auto',
      minWidth: 0,
      whiteSpace: 'normal',
      lineHeight: 1.3,
    },
    collapsedLabel: {
      width: '100%',
      textAlign: 'center',
      fontSize: '0.72rem',
      letterSpacing: '0.14em',
    },
    secondary: {
      marginLeft: 'auto',
      display: 'flex',
      alignItems: 'center',
    },
  };
});

export const SidebarPrimaryItem = ({
  icon: Icon,
  text,
  to,
  secondaryAction,
}: SidebarPrimaryItemProps) => {
  const classes = useStyles();
  const { isOpen } = useSidebarOpenState();

  return (
    <NavLink
      to={to}
      aria-label={text}
      className={({ isActive }) => {
        const classNames = [classes.link];
        if (!isOpen) {
          classNames.push(classes.collapsed);
        }
        if (isActive) {
          classNames.push(classes.active);
        }
        return classNames.join(' ');
      }}
    >
      <Icon className={classes.icon} />
      <span
        className={[classes.label, !isOpen ? classes.collapsedLabel : undefined]
          .filter(Boolean)
          .join(' ')}
      >
        {text}
      </span>
      {secondaryAction ? (
        <span className={classes.secondary}>{secondaryAction}</span>
      ) : null}
    </NavLink>
  );
};

