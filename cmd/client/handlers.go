package main

import (
	"fmt"
	"log"
	"time"

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

func ackMoveOutcome(outcome gamelogic.MoveOutcome) pubsub.Acktype {
	switch outcome {
	case gamelogic.MoveOutComeSafe:
		return pubsub.AcktypeAck
	case gamelogic.MoveOutcomeMakeWar:
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
		if outcome == gamelogic.MoveOutcomeMakeWar {
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: mv.Player,
					Defender: gs.GetPlayerSnap(),
				},
			); err != nil {
				log.Printf("%v\n", err)
				return pubsub.AcktypeNackRequeue
			}
		}
		return ackMoveOutcome(outcome)
	}
}

func ackWarOutcome(outcome gamelogic.WarOutcome) pubsub.Acktype {
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

func warLog(outcome gamelogic.WarOutcome, winner, loser string) string {
	switch outcome {
	case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
		return fmt.Sprintf("%s won a war against %s", winner, loser)
	case gamelogic.WarOutcomeDraw:
		return fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
	default:
		return "No war"
	}
}

func publishGameLog(ch *amqp.Channel, log string, attackerName string) error {
	gameLog := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     log,
		Username:    attackerName,
	}
	fmt.Println("Publishing game log")
	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+attackerName, gameLog)
	if err != nil {
		return err
	}
	return nil
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		logg := warLog(outcome, winner, loser)
		switch outcome {
		case gamelogic.WarOutcomeDraw, gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			err := publishGameLog(ch, logg, gs.GetUsername())
			if err != nil {
				return pubsub.AcktypeNackRequeue
			}
		}
		return ackWarOutcome(outcome)
	}
}
