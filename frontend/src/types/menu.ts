export type MenuNodeType = 'directory' | 'menu' | 'button'

export interface MenuNode {
  id: string
  type: MenuNodeType
  title: string
  path?: string
  icon?: string
  permissionCode?: string
  dataPermMode?: 'NONE' | 'ORG' | 'BU' | 'ORG_BU'
  enabled?: boolean
  children?: MenuNode[]
}
