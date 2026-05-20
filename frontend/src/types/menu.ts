export type MenuNodeType = 'directory' | 'menu' | 'button'
export type TenantScope =
  | 'platform_only'
  | 'enterprise_only'
  | 'personal_only'
  | 'all'
  | 'platform_enterprise'
  | 'enterprise_personal'
  | 'platform_personal'

export interface MenuNode {
  id: string
  type: MenuNodeType
  title: string
  path?: string
  icon?: string
  permissionCode?: string
  isPlatformOnly?: boolean
  tenantScope?: TenantScope
  showInAdmin?: boolean
  dataPermMode?: 'NONE' | 'ORG' | 'BU' | 'ORG_BU'
  enabled?: boolean
  children?: MenuNode[]
}
