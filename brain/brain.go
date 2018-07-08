package brain

import (
	"github.com/makeitplay/client-player-go/Game"
	"github.com/makeitplay/commons"
	"github.com/makeitplay/commons/GameState"
	"math"
	"github.com/makeitplay/commons/Units"
	"github.com/makeitplay/commons/BasicTypes"
	"github.com/makeitplay/go-dummy/strategy"
	"fmt"
)

// distance considered "near" for a player to the ball
const DistanceNearBall = strategy.RegionWidth // units float
const ERROR_MARGIN_RUNNING = 20.0
const ERROR_MARGIN_PASSING = 20.0

var TeamState = strategy.Defensive

type Brain struct {
	*Game.Player
	State PlayerState
}

func (b *Brain) ResetPosition() {
	region := b.GetActiveRegion(strategy.Defensive)
	b.Coords = strategy.GetRegionCenter(region, b.TeamPlace)
}

func (b *Brain) ProcessAnn(msg Game.GameMessage) {
	b.UpdatePosition(msg.GameInfo)
	commons.LogBroadcast(string(msg.State))
	switch GameState.State(msg.State) {
	case GameState.GETREADY:
	case GameState.LISTENING:
		TeamState = strategy.DetermineTeamState(msg, b.TeamPlace)
		b.State = b.DetermineMyState()
		b.TakeAnAction()
	}
}

func (b *Brain) DetermineMyState() PlayerState {
	var isOnMyField bool
	var subState string
	var ballPossess string

	if b.LastMsg.GameInfo.Ball.Holder == nil {
		ballPossess = "dsp" //disputing
		subState = "fbl"    //far
		if int(math.Abs(b.Coords.DistanceTo(b.LastMsg.GameInfo.Ball.Coords))) <= DistanceNearBall {
			subState = "nbl" //near
		}
	} else if b.LastMsg.GameInfo.Ball.Holder.TeamPlace == b.TeamPlace {
		ballPossess = "atk" //attacking
		subState = "hlp"    //helping
		if b.LastMsg.GameInfo.Ball.Holder.Id == b.Id {
			subState = "hld" //holding
		}
	} else {
		ballPossess = "dfd" //defending
		subState = "org"
		if b.isItInMyActiveRegion(b.LastMsg.GameInfo.Ball.Coords, strategy.Defensive) {
			subState = "mrg"
		}
	}

	if b.TeamPlace == Units.HomeTeam {
		isOnMyField = b.LastMsg.GameInfo.Ball.Coords.PosX <= Units.CourtWidth/2
	} else {
		isOnMyField = b.LastMsg.GameInfo.Ball.Coords.PosX >= Units.CourtWidth/2
	}
	fieldState := "fr"
	if isOnMyField {
		fieldState = "hs"
	}
	return PlayerState(ballPossess + "-" + subState + "-" + fieldState)
}

func (b *Brain) TakeAnAction() {
	var orders []BasicTypes.Order
	var msg string

	switch b.State {
	case DsptNfblHse:
		msg, orders = b.orderForDsptNfblHse()
		orders = append(orders, b.CreateCatchOrder())
		//case DsptNfblFrg:
		//	msg, orders = b.orderForDsptNfblFrg()
		//	orders = append(orders, b.CreateCatchOrder())
	case DsptFrblHse:
		msg, orders = b.orderForDsptFrblHse()
		orders = append(orders, b.CreateCatchOrder())
	case DsptFrblFrg:
		msg, orders = b.orderForDsptNfblFrg()
		orders = append(orders, b.CreateCatchOrder())

	case AtckHoldHse:
		msg, orders = b.orderForAtckHoldHse()
	case AtckHoldFrg:
		msg, orders = b.orderForAtckHoldFrg()
		//case AtckHelpHse:
		//	msg, orders = b.orderForAtckHelpHse()
		//case AtckHelpFrg:
		//	msg, orders = b.orderForAtckHelpFrg()

		//case DefdMyrgHse:
		//	msg, orders = b.orderForDefdMyrgHse()
		//	orders = append(orders, b.CreateCatchOrder())
		//case DefdMyrgFrg:
		//	msg, orders = b.orderForDefdMyrgFrg()
		//	orders = append(orders, b.CreateCatchOrder())
		//case DefdOtrgHse:
		//	msg, orders = b.orderForDefdOtrgHse()
		//	orders = append(orders, b.CreateCatchOrder())
		//case DefdOtrgFrg:
		//	msg, orders = b.orderForDefdOtrgFrg()
		//	orders = append(orders, b.CreateCatchOrder())
	default:
		msg = "Freeze position"
		orders = []BasicTypes.Order{b.CreateStopOrder(*b.Velocity.Direction)}

	}

	b.SendOrders(fmt.Sprintf("[%s-%s] %s", b.State, TeamState, msg), orders...)

}

func (b *Brain) ShouldIDisputeForTheBall() bool {
	myDistance := b.Coords.DistanceTo(b.LastMsg.GameInfo.Ball.Coords)
	playerCloser := 0
	for _, teamMate := range b.GetMyTeam(b.LastMsg.GameInfo).Players {
		if teamMate.Number != b.Number && teamMate.Coords.DistanceTo(b.LastMsg.GameInfo.Ball.Coords) < myDistance {
			playerCloser++
			if playerCloser > 1 {
				return false
			}
		}
	}
	return true
}

