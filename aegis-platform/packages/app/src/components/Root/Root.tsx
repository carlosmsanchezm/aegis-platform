import { PropsWithChildren } from 'react';
import clsx from 'clsx';
import { makeStyles } from '@material-ui/core';
import HomeIcon from '@material-ui/icons/Home';
import ExtensionIcon from '@material-ui/icons/Extension';
import LibraryBooks from '@material-ui/icons/LibraryBooks';
import CreateComponentIcon from '@material-ui/icons/AddCircleOutline';
import DashboardIcon from '@material-ui/icons/Dashboard';
import TimelineIcon from '@material-ui/icons/Timeline';
import SecurityIcon from '@material-ui/icons/Security';
import CloudQueueIcon from '@material-ui/icons/CloudQueue';
import LaptopMacIcon from '@material-ui/icons/LaptopMac';
import MenuIcon from '@material-ui/icons/Menu';
import SearchIcon from '@material-ui/icons/Search';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import Brightness4Icon from '@material-ui/icons/Brightness4';
import Brightness7Icon from '@material-ui/icons/Brightness7';
import SettingsIcon from '@material-ui/icons/Settings';
import GroupIcon from '@material-ui/icons/People';
import NotificationsNoneIcon from '@material-ui/icons/NotificationsNone';
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
  SidebarSubheader,
  useSidebarOpenState,
  Link,
} from '@backstage/core-components';
import { useAppTheme } from '@backstage/core-plugin-api';
import { MyGroupsSidebarItem } from '@backstage/plugin-org';
import { NotificationsSidebarItem } from '@backstage/plugin-notifications';

const useSidebarLogoStyles = makeStyles(theme => ({
  root: {
    width: sidebarConfig.drawerWidthClosed,
    height: 3 * sidebarConfig.logoHeight,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: theme.spacing(2, 0, 1.5, 0),
  },
  link: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '100%',
    padding: theme.spacing(0, 2),
    transition: 'padding 200ms ease',
  },
  linkOpen: {
    justifyContent: 'flex-start',
    paddingLeft: theme.spacing(3),
  },
}));

const SidebarLogo = () => {
  const classes = useSidebarLogoStyles();
  const { isOpen } = useSidebarOpenState();

  return (
    <div className={classes.root}>
      <Link
        to="/"
        underline="none"
        className={clsx(classes.link, { [classes.linkOpen]: isOpen })}
        aria-label="Home"
      >
        {isOpen ? <LogoFull /> : <LogoIcon />}
      </Link>
    </div>
  );
};

const ThemeToggleItem = () => {
  const { activeThemeId, setActiveThemeId } = useAppTheme();
  const isDark = activeThemeId !== 'aegis-light';
  const icon = isDark ? Brightness7Icon : Brightness4Icon;
  const label = isDark ? 'Light mode' : 'Dark mode';

  return (
    <SidebarItem
      icon={icon}
      text={label}
      onClick={() => setActiveThemeId(isDark ? 'aegis-light' : 'aegis-dark')}
    />
  );
};

export const Root = ({ children }: PropsWithChildren<{}>) => (
  <SidebarPage>
    <Sidebar>
      <SidebarLogo />
      <SidebarGroup label="Search" icon={<SearchIcon />} to="/search">
        <SidebarSearchModal />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="ÆGIS Console" icon={<DashboardIcon />}>
        <SidebarSubheader>Overview</SidebarSubheader>
        <SidebarItem
          icon={DashboardIcon}
          to="aegis/dashboard"
          text="Control Center"
        />
        <SidebarSubheader>Manage</SidebarSubheader>
        <SidebarItem
          icon={TimelineIcon}
          to="aegis/telemetry"
          text="Telemetry"
        />
        <SidebarItem
          icon={SecurityIcon}
          to="aegis/posture"
          text="Live Posture"
        />
        <SidebarItem icon={CloudQueueIcon} to="aegis/clusters" text="Clusters" />
        <SidebarItem
          icon={LaptopMacIcon}
          to="aegis/workloads"
          text="Workspaces"
        />
        <SidebarSubheader>Create</SidebarSubheader>
        <SidebarItem
          icon={LaptopMacIcon}
          to="aegis/workspaces/create"
          text="Launch Secure Workspace"
        />
        <SidebarItem
          icon={PlayCircleOutlineIcon}
          to="aegis"
          text="Launch Workload"
        />
        {/* TODO: Surface Agent Builder entry once the workflow is wired up. */}
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Platform" icon={<MenuIcon />}>
        <SidebarSubheader>Admin</SidebarSubheader>
        <SidebarItem icon={HomeIcon} to="catalog" text="Catalog" />
        <MyGroupsSidebarItem
          singularTitle="My Group"
          pluralTitle="My Groups"
          icon={GroupIcon}
        />
        <SidebarItem icon={ExtensionIcon} to="api-docs" text="APIs" />
        <SidebarItem icon={LibraryBooks} to="docs" text="Docs" />
        <NotificationsSidebarItem icon={NotificationsNoneIcon} />
        <SidebarItem icon={SettingsIcon} to="/settings" text="Settings" />
        <SidebarItem icon={CreateComponentIcon} to="create" text="Create" />
        <SidebarDivider />
        <SidebarScrollWrapper>
          {/* Reserved for future platform navigation extensions. */}
        </SidebarScrollWrapper>
      </SidebarGroup>
      <SidebarSpace />
      <SidebarDivider />
      <SidebarGroup label="Appearance" icon={<Brightness4Icon />}>
        <ThemeToggleItem />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup
        label="Account"
        icon={<UserSettingsSignInAvatar />}
        to="/settings"
      >
        <SidebarSettings />
      </SidebarGroup>
    </Sidebar>
    {children}
  </SidebarPage>
);
