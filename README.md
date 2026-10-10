# Shouketsu server
Websocket server for managing card game rooms and logic

## Quick Start

```bash
git clone https://github.com/xashaysso/shouketsu-server
cd shouketsu-server
```

*the service currently doesn't use any db so run it directly:*

```bash
go run cmd/server/main.go
```

## 

## WebSocket API

Connection endpoint:
`ws://localhost:8080/ws?player={PLAYER_ID}&room={ROOM_ID}`

All client requests must be sent as JSON objects containing `type` and an optional `payload`.

### Client Actions

#### End Turn
Passes the turn to the opponent, triggers the end of turn combat phase, increments player mana, and draws a card for the next player.
```json
{
  "type": "END_TURN"
}
```
#### Play Card
Plays a card from hand into a specific board slot or targeting a specific entity (minion/player).

Play minion on slot:
```json
{
  "type": "PLAY_CARD",
  "payload": {
    "card_index": 0,
    "slot_index": 1
  }
}
```
Play spell on target
```json
{
  "type": "PLAY_CARD",
  "payload": {
    "card_index": 2,
    "target": {
      "type": "player",
      "side": "opponent"
    }
  }
}
```
