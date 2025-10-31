import { PropsWithChildren } from 'react';
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
import LayersIcon from '@material-ui/icons/Layers';
import PlayCircleFilledIcon from '@material-ui/icons/PlayCircleFilled';
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
  SidebarSpace,
  useSidebarOpenState,
  Link,
} from '@backstage/core-components';
import MenuIcon from '@material-ui/icons/Menu';
import SearchIcon from '@material-ui/icons/Search';
import { MyGroupsSidebarItem } from '@backstage/plugin-org';
import GroupIcon from '@material-ui/icons/People';
import NotificationsActiveIcon from '@material-ui/icons/NotificationsActive';

const useSidebarLogoStyles = makeStyles(theme => ({
  root: {
    width: sidebarConfig.drawerWidthClosed,
    height: 3 * sidebarConfig.logoHeight,
    display: 'flex',
    flexFlow: 'row nowrap',
    alignItems: 'center',
    paddingLeft: theme.spacing(3),
    paddingRight: theme.spacing(2),
    paddingTop: theme.spacing(3),
    paddingBottom: theme.spacing(1),
  },
  link: {
    width: '100%',
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

export const Root = ({ children }: PropsWithChildren<{}>) => (
  <SidebarPage>
    <Sidebar>
      <SidebarLogo />
      <SidebarGroup label="Search" icon={<SearchIcon />} to="/search">
        <SidebarSearchModal />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Create" icon={<PlayCircleFilledIcon />}>
        <SidebarItem
          icon={LaptopMacIcon}
          to="aegis/workspaces/create"
          text="Launch Secure Workspace"
        />
        <SidebarItem icon={CreateComponentIcon} to="aegis" text="Launch Workload" />
        {/* TODO: Add Agent Builder route when available */}
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Manage" icon={<DashboardIcon />}>
        <SidebarItem
          icon={DashboardIcon}
          to="aegis/dashboard"
          text="Control Center"
        />
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
        <SidebarItem icon={LayersIcon} to="aegis/workloads" text="Workspaces" />
      </SidebarGroup>
      <SidebarDivider />
      <SidebarGroup label="Admin" icon={<MenuIcon />}>
        <SidebarItem icon={HomeIcon} to="catalog" text="Catalog" />
        <MyGroupsSidebarItem
          singularTitle="My Group"
          pluralTitle="My Groups"
          icon={GroupIcon}
        />
        <SidebarItem icon={ExtensionIcon} to="api-docs" text="APIs" />
        <SidebarItem icon={LibraryBooks} to="docs" text="Docs" />
        <SidebarItem icon={CreateComponentIcon} to="create" text="Create..." />
        <SidebarItem
          icon={NotificationsActiveIcon}
          to="notifications"
          text="Notifications"
        />
      </SidebarGroup>
      <SidebarSpace />
      <SidebarGroup
        label="Settings"
        icon={<UserSettingsSignInAvatar />}
        to="/settings"
      >
        <SidebarSettings />
      </SidebarGroup>
    </Sidebar>
    {children}
  </SidebarPage>
);
