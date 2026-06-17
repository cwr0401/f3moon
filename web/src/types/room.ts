import { GameMode } from './game'

/** RoomStatus — mirrors room.RoomStatus */
export enum RoomStatus {
  RoomWaiting = 0,
  RoomPlaying = 1,
  RoomFinished = 2,
}

/** RoomPlayer — mirrors room.RoomPlayer */
export interface RoomPlayer {
  id: string
  name: string
  is_ai: boolean
  seat: number
  ready: boolean
}

/** Room — mirrors room.Room */
export interface Room {
  id: string
  zone_id: string
  name: string
  mode: GameMode
  status: RoomStatus
  players: (RoomPlayer | null)[]
  owner: string
  max_players: number
  max_rounds: number
  current_round: number
}
