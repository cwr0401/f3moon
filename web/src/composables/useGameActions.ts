import { useGameStore } from '../stores/game'
import type { TileName } from '../types/tile'

export function useGameActions() {
  const store = useGameStore()

  function doCut(position: number) {
    return store.cut(position)
  }

  function doDeal() {
    return store.deal()
  }

  function doTong(tileName: TileName, tongSize: number) {
    return store.tong(tileName, tongSize, false)
  }

  function doTongSkip() {
    return store.tong('' as TileName, 0, true)
  }

  function doDraw() {
    return store.draw()
  }

  function doDiscard(tileId: number) {
    store.selectTile(null)
    return store.discard(tileId)
  }

  function doPair(tileId: number, pairSize: number) {
    return store.pair(tileId, pairSize)
  }

  function doGanta() {
    return store.ganta()
  }

  function doWin() {
    return store.win()
  }

  function doPass() {
    return store.pass()
  }

  function doDangJing(jing: TileName) {
    return store.dangJing(jing)
  }

  return { doCut, doDeal, doTong, doTongSkip, doDraw, doDiscard, doPair, doGanta, doWin, doPass, doDangJing }
}
