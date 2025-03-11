package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.AcktypeAck
	}
}

func ackMoveOutcome(outcome gamelogic.MoveOutcome, movingPlayer, currentPlayer gamelogic.Player, ch *amqp.Channel) pubsub.Acktype {
	switch outcome {
	case gamelogic.MoveOutComeSafe:
		return pubsub.AcktypeAck
	case gamelogic.MoveOutcomeMakeWar:
		if err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+currentPlayer.Username,
			gamelogic.RecognitionOfWar{
				Attacker: movingPlayer,
				Defender: currentPlayer,
			},
		); err != nil {
			return pubsub.AcktypeNackRequeue
		}
		return pubsub.AcktypeAck
	case gamelogic.MoveOutcomeSamePlayer:
		return pubsub.AcktypeNackDiscard
	default:
		log.Printf("Error: unkown move outcome: %v", outcome)
		return pubsub.AcktypeNackDiscard
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(mv gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		return ackMoveOutcome(outcome, mv.Player, gs.GetPlayerSnap(), ch)
	}
}

func ackTypeForWarOutcome(outcome gamelogic.WarOutcome) pubsub.Acktype {
	switch outcome {
	case gamelogic.WarOutcomeNotInvolved:
		return pubsub.AcktypeNackRequeue
	case gamelogic.WarOutcomeNoUnits:
		return pubsub.AcktypeNackDiscard
	case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
		return pubsub.AcktypeAck
	default:
		log.Printf("Error: unkown war outcome: %v", outcome)
		return pubsub.AcktypeNackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)
		return ackTypeForWarOutcome(outcome)
	}
}
