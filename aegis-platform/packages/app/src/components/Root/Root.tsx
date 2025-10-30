import { PropsWithChildren } from 'react';
import { Box, makeStyles } from '@material-ui/core';
import ExtensionIcon from '@material-ui/icons/Extension';
import LibraryBooks from '@material-ui/icons/LibraryBooks';
import CreateComponentIcon from '@material-ui/icons/AddCircleOutline';
import LogoFull from './LogoFull';
import LogoIcon from './LogoIcon';
import {
  Settings as SidebarSettings,
  UserSettingsSignInAvatar,
} from '@backstage/plugin-user-settings';
import { SidebarSearchModal } from '@backstage/plugin-search';
import {
  Sidebar,
  sidebarConfig,
  SidebarDivider,
  SidebarGroup,
  SidebarItem,
  SidebarPage,
  SidebarScrollWrapper,
  SidebarSpace,
  useSidebarOpenState,
  Link,
} from '@backstage/core-components';
import MenuIcon from '@material-ui/icons/Menu';
import SearchIcon from '@material-ui/icons/Search';
import { MyGroupsSidebarItem } from '@backstage/plugin-org';
import GroupIcon from '@material-ui/icons/People';
import { NotificationsSidebarItem } from '@backstage/plugin-notifications';
import StorageIcon from '@material-ui/icons/Storage';
import DashboardIcon from '@material-ui/icons/Dashboard';
import DnsIcon from '@material-ui/icons/Dns';
import { alpha } from '@material-ui/core/styles/colorManipulator';

const useSidebarLogoStyles = makeStyles({
  root: {
    width: sidebarConfig.drawerWidthClosed,
    height: 3 * sidebarConfig.logoHeight,
    display: 'flex',
    flexFlow: 'row nowrap',
    alignItems: 'center',
    marginBottom: -14,
  },
  link: {
    width: sidebarConfig.drawerWidthClosed,
    marginLeft: 24,
  },
});

const useStyles = makeStyles(theme => ({
  root: {
    position: 'relative',
    minHeight: '100vh',
    background: 'transparent',
  },
  background: {
    position: 'fixed',
    inset: 0,
    pointerEvents: 'none',
    backgroundImage: `radial-gradient(circle at 10% 10%, ${alpha(
      theme.palette.primary.main,
      0.2,
    )} 0, transparent 55%), radial-gradient(circle at 90% 15%, ${alpha(
      theme.palette.secondary.main,
      0.18,
    )} 0, transparent 60%), radial-gradient(circle at 50% 80%, ${alpha(
      '#4b9eff',
      0.2,
    )} 0, transparent 55%)`,
    opacity: 0.65,
    filter: 'blur(42px)',
    zIndex: 0,
  },
  page: {
    position: 'relative',
    zIndex: 1,
  },
  sidebar: {
    background: alpha('#030914', 0.84),
    backdropFilter: 'blur(18px)',
    borderRight: `1px solid ${alpha(theme.palette.primary.main, 0.2)}`,
    boxShadow: '0 30px 80px rgba(16, 255, 219, 0.1)',
    '& .MuiListItem-root': {
      borderRadius: 12,
      margin: '4px 12px',
      padding: '10px 12px',
      transition: 'background 240ms ease, transform 240ms ease',
    },
    '& .MuiListItem-root.Mui-selected, & .MuiListItem-root:hover': {
      background: alpha(theme.palette.primary.main, 0.18),
      transform: 'translateX(4px)',
      boxShadow: '0 0 22px rgba(83, 251, 224, 0.25)',
    },
    '& .MuiTypography-root': {
      letterSpacing: '0.14em',
      textTransform: 'uppercase',
      fontSize: '0.72rem',
    },
  },
  content: {
    position: 'relative',
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    zIndex: 1,
  },
}));

const SidebarLogo = () => {
  const classes = useSidebarLogoStyles();
  const { isOpen } = useSidebarOpenState();

  return (
    <div className={classes.root}>
      <Link to="/" underline="none" className={classes.link} aria-label="Home">
        {isOpen ? <LogoFull /> : <LogoIcon />}
      </Link>
    </div>
  );
};

export const Root = ({ children }: PropsWithChildren<{}>) => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <div className={classes.background} />
      <SidebarPage className={classes.page}>
        <Sidebar className={classes.sidebar}>
          <SidebarLogo />
          <SidebarGroup label="Search" icon={<SearchIcon />} to="/search">
            <SidebarSearchModal />
          </SidebarGroup>
          <SidebarDivider />
          <SidebarGroup label="Mission" icon={<MenuIcon />}>
            {/* Global nav, not org-specific */}
            <SidebarItem icon={DashboardIcon} to="dashboard" text="Command" />
            <SidebarItem icon={DnsIcon} to="catalog" text="Catalog" />
            <MyGroupsSidebarItem
              singularTitle="My Group"
              pluralTitle="My Groups"
              icon={GroupIcon}
            />
            <SidebarItem icon={ExtensionIcon} to="api-docs" text="APIs" />
            <SidebarItem icon={LibraryBooks} to="docs" text="Docs" />
            <SidebarItem icon={CreateComponentIcon} to="create" text="Create" />
            <SidebarItem icon={StorageIcon} to="aegis" text="Submit Workload" />
            <SidebarItem
              icon={StorageIcon}
              to="aegis/workloads"
              text="My Workloads"
            />
            {/* End global nav */}
            <SidebarDivider />
            <SidebarScrollWrapper>
              {/* Items in this group will be scrollable if they run out of space */}
            </SidebarScrollWrapper>
          </SidebarGroup>
          <SidebarSpace />
          <SidebarDivider />
          <NotificationsSidebarItem />
          <SidebarDivider />
          <SidebarGroup
            label="Settings"
            icon={<UserSettingsSignInAvatar />}
            to="/settings"
          >
            <SidebarSettings />
          </SidebarGroup>
        </Sidebar>
        <Box className={classes.content}>{children}</Box>
      </SidebarPage>
    </div>
  );
};
