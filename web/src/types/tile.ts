/** TileName — mirrors model.TileName */
export type TileName =
  // Numeric tiles 1-10
  | '乙' | '二' | '三' | '四' | '五'
  | '六' | '七' | '八' | '九' | '十'
  // Word tiles
  | '孔' | '己' | '化' | '千' | '土' | '子'
  | '上' | '大' | '人' | '可' | '知' | '礼'
  // Special
  | '别'

/** TileColor — mirrors model.TileColor */
export enum TileColor {
  Black = 0,
  Red = 1,
}

/** Tile — mirrors model.Tile */
export interface Tile {
  id: number
  name: TileName
  color: TileColor
  is_flower: boolean
  numeric: number
}

/** Jing tiles: 三, 五, 七 */
export const JING_TILES: TileName[] = ['三', '五', '七']

/** Red tile names */
export const RED_TILE_NAMES: Set<TileName> = new Set([
  '三', '五', '七', '上', '大', '人', '可', '知', '礼',
])

/** Black tile names */
export const BLACK_TILE_NAMES: Set<TileName> = new Set([
  '乙', '二', '四', '六', '八', '九', '十',
  '孔', '己', '化', '千', '土', '子',
])

/** Check if a tile name is jing */
export function isJingName(name: TileName): boolean {
  return JING_TILES.includes(name)
}
