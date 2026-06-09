import type { Tile, TileName } from './tile'

/** GameMode — mirrors model.GameMode */
export enum GameMode {
  Mode4Player = 0,
  Mode3Player = 1,
}

/** GamePhase — mirrors model.GamePhase */
export enum GamePhase {
  PhaseWaiting = 0,
  PhaseShuffle = 1,
  PhaseCut = 2,
  PhaseDeal = 3,
  PhaseTongAsk = 4,
  PhasePlay = 5,
  PhaseCheck = 6,
  PhaseFinished = 7,
}

/** WinType — mirrors model.WinType */
export enum WinType {
  WinZiMo = 0,
  WinDianPao = 1,
  WinTianHu = 2,
  WinHaiDi = 3,
}

/** PlayerRole — mirrors model.PlayerRole */
export enum PlayerRole {
  RoleDealer = 0,
  RoleIdle1 = 1,
  RoleIdle2 = 2,
  RoleRest = 3,
}

/** CombType — mirrors model.CombType */
export enum CombType {
  CombWord = 0,
  CombNumeric = 1,
  CombSingle = 2,
}

/** CombCompleteness — mirrors model.CombCompleteness */
export enum CombCompleteness {
  CombComplete = 0,
  CombIncomplete = 1,
}

/** Combination — mirrors model.Combination */
export interface Combination {
  type: CombType
  completeness: CombCompleteness
  tiles: Tile[]
  bie_mapping: Record<number, TileName>
}

/** Player — mirrors model.Player */
export interface Player {
  id: string
  name: string
  role: PlayerRole
  is_ai: boolean
  hand: Tile[]
  open_tiles: Tile[]
  open_combs: Combination[]
  dang_jing: TileName | ''
  pair_count: number
  online: boolean
}

/** GameState — mirrors model.GameState */
export interface GameState {
  id: string
  phase: GamePhase
  mode: GameMode
  players: (Player | null)[]
  draw_pile: Tile[]
  discard_pile: Tile[]
  current_turn: number
  dealer_index: number
  cut_player_index: number
  tong_requested: boolean
  tong_order: number[]
  tong_current: number
  last_discard: Tile | null
  last_discarder: number
  winner: number
  win_type: WinType
}
