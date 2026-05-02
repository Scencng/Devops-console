import {
  Briefcase,
  Connection,
  DataBoard,
  Document,
  DocumentCopy,
  Histogram,
  House,
  Menu,
  Monitor,
  OfficeBuilding,
  Search,
  Setting,
  User,
  UserFilled,
} from '@element-plus/icons-vue'

export const MENU_ICON_MAP = {
  Briefcase,
  Connection,
  DataBoard,
  Document,
  DocumentCopy,
  Histogram,
  House,
  Menu,
  Monitor,
  OfficeBuilding,
  Search,
  Setting,
  User,
  UserFilled,
}

export const resolveMenuIcon = (iconName) => MENU_ICON_MAP[iconName] || null
