package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.AcktypeAck
	}
}

func ackTypeForMoveOutcome(outcome gamelogic.MoveOutcome) pubsub.Acktype {
	switch outcome {
	case gamelogic.MoveOutComeSafe, gamelogic.MoveOutcomeMakeWar:
		return pubsub.AcktypeAck
	case gamelogic.MoveOutcomeSamePlayer:
		return pubsub.AcktypeNackDiscard
	default:
		return pubsub.AcktypeNackDiscard
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(mv gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		return ackTypeForMoveOutcome(outcome)
	}
}
