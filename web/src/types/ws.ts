/** NotifyType — mirrors game.NotifyType */
export type NotifyType =
  | 'game:start'
  | 'game:cut-wait'
  | 'game:dealt'
  | 'game:tong-ask'
  | 'game:tong-turn'
  | 'game:turn'
  | 'game:draw'
  | 'game:discard'
  | 'game:pair'
  | 'game:ganta'
  | 'game:win'
  | 'game:huang'
  | 'game:check'
  | 'player:hand'
  | 'game:phase'

/** WSMessage — mirrors ws.WSMessage */
export interface WSMessage {
  type: NotifyType
  game_id: string
  player: string
  data: unknown
  seq: number
}
