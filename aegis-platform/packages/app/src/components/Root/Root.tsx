import { PropsWithChildren } from 'react';
import { makeStyles } from '@material-ui/core';
import HomeIcon from '@material-ui/icons/Home';
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
import TimelineIcon from '@material-ui/icons/Timeline';
import FlightIcon from '@material-ui/icons/FlightTakeoff';
import LayersIcon from '@material-ui/icons/Layers';

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

const useRootStyles = makeStyles(theme => ({
  page: {
    background: 'transparent',
    color: theme.palette.text.primary,
    display: 'flex',
    minHeight: '100vh',
  },
  sidebarContainer: {
    '& .MuiDrawer-paper': {
      background:
        'linear-gradient(180deg, rgba(6, 11, 22, 0.95) 0%, rgba(8, 16, 34, 0.92) 60%, rgba(10, 18, 36, 0.88) 100%)',
      borderRight: '1px solid rgba(0, 245, 255, 0.15)',
      backdropFilter: 'blur(14px)',
      boxShadow: '0 0 48px rgba(0, 245, 255, 0.08)',
    },
  },
  main: {
    position: 'relative',
    overflow: 'hidden',
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh',
  },
  backdrop: {
    position: 'absolute',
    inset: 0,
    pointerEvents: 'none',
    backgroundImage:
      'radial-gradient(circle at 20% 25%, rgba(0, 245, 255, 0.12) 0, rgba(5, 9, 18, 0) 55%),\
       radial-gradient(circle at 80% 10%, rgba(255, 45, 149, 0.12) 0, rgba(5, 9, 18, 0) 50%),\
       linear-gradient(110deg, rgba(0, 245, 255, 0.06) 0%, rgba(8, 16, 34, 0) 40%),\
       repeating-linear-gradient(90deg, rgba(0, 245, 255, 0.04) 0 1px, transparent 1px 40px)',
    opacity: 0.8,
  },
  grid: {
    position: 'absolute',
    inset: 0,
    pointerEvents: 'none',
    backgroundImage:
      'linear-gradient(rgba(0, 245, 255, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 245, 255, 0.03) 1px, transparent 1px)',
    backgroundSize: '120px 120px',
    mixBlendMode: 'screen',
  },
  content: {
    position: 'relative',
    zIndex: 1,
    padding: theme.spacing(4, 6, 6),
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  sidebarGroup: {
    marginTop: theme.spacing(1.5),
    padding: theme.spacing(0, 1.5),
    '& .MuiTypography-root': {
      letterSpacing: '0.16em',
      fontSize: '0.7rem',
      color: 'rgba(124, 154, 196, 0.9)',
    },
  },
  sidebarItem: {
    margin: theme.spacing(0.5, 1.5),
    borderRadius: 14,
    overflow: 'hidden',
    '& a': {
      borderRadius: 14,
      padding: theme.spacing(1.4, 2),
      backgroundColor: 'rgba(0, 245, 255, 0.06)',
      transition: 'background-color 180ms ease, transform 180ms ease',
      letterSpacing: '0.08em',
      textTransform: 'uppercase',
      fontSize: '0.75rem',
      color: '#e5f4ff !important',
    },
    '& a:hover': {
      backgroundColor: 'rgba(0, 245, 255, 0.18)',
      transform: 'translateX(2px)',
    },
    '& .MuiListItemIcon-root': {
      color: '#00f5ff',
    },
  },
  divider: {
    backgroundColor: 'rgba(0, 245, 255, 0.12)',
    margin: theme.spacing(0.5, 1.5),
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
  const classes = useRootStyles();

  return (
    <div className={classes.page}>
      <SidebarPage>
        <div className={classes.sidebarContainer}>
          <Sidebar>
            <SidebarLogo />
            <div className={classes.sidebarGroup}>
              <SidebarGroup
                label="Quick Intel"
                icon={<SearchIcon />}
                to="/search"
              >
                <SidebarSearchModal />
              </SidebarGroup>
            </div>
            <SidebarDivider className={classes.divider} />
            <div className={classes.sidebarGroup}>
              <SidebarGroup label="Mission Modules" icon={<MenuIcon />}>
                <SidebarItem
                  icon={HomeIcon}
                  to="catalog"
                  text="Mission Control"
                  className={classes.sidebarItem}
                />
                <SidebarItem
                  icon={TimelineIcon}
                  to="catalog-graph"
                  text="Service Mesh"
                  className={classes.sidebarItem}
                />
                <div className={classes.sidebarItem}>
                  <MyGroupsSidebarItem
                    singularTitle="Strike Team"
                    pluralTitle="Strike Teams"
                    icon={GroupIcon}
                  />
                </div>
                <SidebarItem
                  icon={ExtensionIcon}
                  to="api-docs"
                  text="Signal APIs"
                  className={classes.sidebarItem}
                />
                <SidebarItem
                  icon={LibraryBooks}
                  to="docs"
                  text="Playbooks"
                  className={classes.sidebarItem}
                />
                <SidebarItem
                  icon={CreateComponentIcon}
                  to="create"
                  text="Forge Asset"
                  className={classes.sidebarItem}
                />
                <SidebarDivider className={classes.divider} />
                <SidebarItem
                  icon={FlightIcon}
                  to="aegis/workspaces/create"
                  text="Launch Workspace"
                  className={classes.sidebarItem}
                />
                <SidebarItem
                  icon={StorageIcon}
                  to="aegis/workloads"
                  text="Mission Logs"
                  className={classes.sidebarItem}
                />
                <SidebarItem
                  icon={LayersIcon}
                  to="aegis"
                  text="Legacy Form"
                  className={classes.sidebarItem}
                />
                <SidebarScrollWrapper />
              </SidebarGroup>
            </div>
            <SidebarSpace />
            <SidebarDivider className={classes.divider} />
            <NotificationsSidebarItem />
            <SidebarDivider className={classes.divider} />
            <div className={classes.sidebarGroup}>
              <SidebarGroup
                label="Settings"
                icon={<UserSettingsSignInAvatar />}
                to="/settings"
              >
                <SidebarSettings />
              </SidebarGroup>
            </div>
          </Sidebar>
        </div>
        <div className={classes.main}>
          <div className={classes.backdrop} />
          <div className={classes.grid} />
          <div className={classes.content}>{children}</div>
        </div>
      </SidebarPage>
    </div>
  );
};
