package statemachine

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"os"
)

type State uint8

const (
	WELCOME_SCREEN State = iota
	OPTIONS_MENU
	GAME
	PAUSE
	GAME_OVER
	EXIT
)

func (s *State) String() string {
	return [...]string{"WELCOME_SCREEN", "OPTIONS_MENU", "GAME", "PAUSE", "GAME_OVER", "EXIT"}[*s]
}

func (s *State) EnumIndex() uint8 {
	return uint8(*s)
}

type StateMachine struct {
	States       []State
	CurrentState State
	Transitions  map[State][]State
}

func NewStateMachine() *StateMachine {
	states := make([]State, 6, 6)

	states[WELCOME_SCREEN] = WELCOME_SCREEN
	states[OPTIONS_MENU] = OPTIONS_MENU
	states[GAME] = GAME
	states[PAUSE] = PAUSE
	states[GAME_OVER] = GAME_OVER
	states[EXIT] = EXIT

	initialState := WELCOME_SCREEN

	sm := &StateMachine{
		States:       states,
		CurrentState: initialState,
		Transitions:  make(map[State][]State),
	}

	sm.Transitions[WELCOME_SCREEN] = make([]State, 3, 3)
	sm.Transitions[WELCOME_SCREEN][0] = EXIT
	sm.Transitions[WELCOME_SCREEN][1] = GAME
	sm.Transitions[WELCOME_SCREEN][2] = OPTIONS_MENU

	sm.Transitions[GAME] = make([]State, 4, 4)
	sm.Transitions[GAME][0] = EXIT
	sm.Transitions[GAME][1] = WELCOME_SCREEN
	sm.Transitions[GAME][2] = GAME_OVER
	sm.Transitions[GAME][3] = PAUSE

	sm.Transitions[PAUSE] = make([]State, 4, 4)
	sm.Transitions[PAUSE][0] = EXIT
	sm.Transitions[PAUSE][1] = GAME
	sm.Transitions[PAUSE][2] = OPTIONS_MENU
	sm.Transitions[PAUSE][3] = WELCOME_SCREEN

	sm.Transitions[GAME_OVER] = make([]State, 2, 2)
	sm.Transitions[GAME_OVER][0] = EXIT
	sm.Transitions[GAME_OVER][1] = WELCOME_SCREEN

	return sm
}

func (sm *StateMachine) DoTransition(from, to State) bool {
	for _, toAvailableStates := range sm.Transitions {
		for _, toState := range toAvailableStates {
			if toState == to {
				fromStateStr := from.String()
				toStateStr := toState.String()
				stateCase := fromStateStr + ":" + toStateStr

				switch stateCase {
				case "WELCOME_SCREEN:EXIT":
					log.Println(stateCase)
					sm.CurrentState = toState
					log.Println("State transition from", from, "to", sm.CurrentState)
					sdl.Quit()
					os.Exit(0)
					return true
				case "WELCOME_SCREEN:GAME":
					log.Println(stateCase)
					log.Printf("%v\n", *(&sm.CurrentState))
					sm.CurrentState = toState
					log.Printf("%v\n", *(&sm.CurrentState))
					//os.Exit(1)
					log.Println("State transition from", from, "to", sm.CurrentState)
					return true
				case "WELCOME_SCREEN:OPTIONS_MENU":
					log.Println(stateCase)
					sm.CurrentState = toState
					return true
				case "GAME:PAUSE":
					log.Println(stateCase)
					sm.CurrentState = toState
					log.Println("State transition from", from, "to", sm.CurrentState)
					return true
				case "PAUSE:GAME":
					log.Println(stateCase)
					sm.CurrentState = toState
					log.Println("State transition from", from, "to", sm.CurrentState)
					return true
				}
			}
		}
	}
	return false
}
