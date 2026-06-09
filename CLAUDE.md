# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

f3moon (fair flowers and full moon, 花好月圆) is a digital implementation of the traditional Chinese local character card game 荆楚花牌 (Jingchu Flower Cards).

### Key Game Concepts

- **Deck**: 112 cards total
  - 10 types of number cards (乙, 二, 三, 四, 五, 六, 七, 八, 九, 十) - 5 copies each
  - 12 types of character cards (孔, 己, 化, 千, 土, 子, 上, 大, 人, 可, 知, 礼) - 5 copies each
  - 2 special "别" cards that can substitute as 花三, 花五, 花七

- **Card Colors**: Black and red with optional flower markings
- **Jing Cards (经牌)**: 三, 五, 七 - these have special scoring multipliers
- **Combinations**: 
  - Text combinations: 孔乙己, 化三千, 七十土, 八九子, 上大人, 可知礼
  - Number sequences: 乙二三, 二三四, 三四五, 四五六, 五六七, 六七八, 七八九, 八九十
  - Sets (坎): 3+ identical cards
  - Special combinations with "别" cards

- **Players**: 3-4 players (one may be a non-playing 歇家)
- **Winning (胡牌)**: Requires ≥17 胡 points and valid combination structure

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Gin v1.12.0
- **Dependencies**: See go.mod for full list

## Common Commands

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run a specific test
go test ./... -run TestName

# Build the project
go build ./...

# Format code
go fmt ./...
```

## Architecture

This is a Go project structured to implement the complete game logic for 荆楚花牌. The rules are fully documented in rules.md and should be consulted as the authoritative source for game mechanics.

Key architectural components will likely include:
- **Card models**: Representing individual cards with their attributes (type, color, flower marking)
- **Deck management**: Shuffling, cutting, dealing
- **Hand evaluation**: Combination validation, scoring (算胡), waiting hand (听牌) detection
- **Game state management**: Turn progression, player actions, winning detection

## Key Reference Files

- **rules.md**: Complete game rules in Chinese - authoritative source for all game mechanics
- **README.md**: Basic project overview
- **go.mod**: Go module dependencies
