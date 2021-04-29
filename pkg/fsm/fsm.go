package fsm

import (
	"errors"
	"sync"
)

var ErrEventRejected = errors.New("event rejected")
var ErrConfiguration = errors.New("configuration error")

const (
	// Default represents the default state of the system.
	Default StateType = ""
	// NoOp represent a no-op event.
	NoOp EventType = "NoOp"
)

// StateType represents an extensible state type in the state machine.
type StateType string

// EventType represents an extensible event type in the state machine.
type EventType string

// StateContext represents the FSM internal context.
type StateContext interface {}

// EventContext represents the event context.
type EventContext interface {}

// ExitAction represents the action to be executed when exiting a given state.
type ExitAction interface {
	Execute(stateContext StateContext, event Event) (StateContext, error)
}

// EntryAction represents the action to be executed when entering a given state.
type EntryAction interface {
	Execute(stateContext StateContext, event Event) (StateContext, Event, error)
}

// Action represents an entry action that can't modify the internal context, nor define a next Event.
type Action interface {
	Execute(interalCtx StateContext, event Event) error
}

// Events represents a mapping of events and states.
type Events map[EventType]StateType

// State binds a state with an action and a set of events it can handle.
type State struct {
	Action        Action
	EntryAction		EntryAction
	ExitAction		ExitAction
	Events				Events
}

// Event represent an event that can potentially update the FSM.
type Event struct {
	Type EventType
	Context EventContext
}

// States represents a mapping of states and their implementations.
type States map[StateType]State

// StateMachine represents the state machine.
type StateMachine struct {
	// Previous represents the previous state.
	Previous StateType
	// Current represents the current state.
	Current StateType
	// States hold the configuration of states and events handled by the state machine.
	States States
	// context holds the context of the current FSM.
	Context StateContext
	// mutex ensures that only 1 event is processed by the state machine at any given time.
	mutex sync.Mutex
}

// rejectEvent returns the Default StateType and an ErrEventRejected.
func (s *StateMachine) rejectEvent() (StateType, error) {
	return Default, ErrEventRejected
}

// getNextState returns the next state for the event fiven the machine's current
// state, or an error if the event can't be handled in the given state.
func (s *StateMachine) getNextState(event EventType) (StateType, error) {
	// Get the current state
	state, ok := s.States[s.Current];
	if !ok {
		return s.rejectEvent()
	}
	// Check that the state has valid Events.
	if state.Events == nil {
		return s.rejectEvent()
	}
	// Get the nest state
	next, ok := state.Events[event]
	if !ok {
		return s.rejectEvent()
	}

	return next, nil
}

// SendEvent sends an event t the state machine.
func (s *StateMachine) SendEvent(event Event) error {
	// Apply a lock to the StateMachine.
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// Apply the event inside a loop to handle additional event changes sequentially.
	// The loop breaks after an action of any event returns a NoOp.
	for {
		var err error

		// Determine the next state for the event given the machine's current state.
		nextState, err := s.getNextState(event.Type)
		if err != nil {
			return err
		}

		// Get the state definition for the next state.
		state, ok := s.States[nextState]
		if !ok {
			return ErrConfiguration
		}

		// Execute the exit action if it is defined.
		if state.ExitAction != nil {
			nextInternalCtx, err := state.ExitAction.Execute(s.Context, event)
			if err != nil {
				return err
			}
			s.Context = nextInternalCtx
		}

		// Transition over to the next state.
		s.Previous = s.Current
		s.Current = nextState

		// Define the next event as a NoOp by default
		nextEvent := Event{Type: NoOp}

		// Execute the entry action if it is defined.
		if state.EntryAction != nil {
			var nextInternalCtx StateContext
			nextInternalCtx, nextEvent, err = state.EntryAction.Execute(s.Context, event)
			if err != nil {
				return err
			}
			s.Context = nextInternalCtx
		} else if state.Action != nil {
			err := state.Action.Execute(s.Context, event)
			if err != nil {
				return err
			}
		}

		// Return if the next event is a no-op.
		if nextEvent.Type == NoOp {
			return nil
		}

		// Set the nextEvent as the event and run the loop again.
		event = nextEvent
	}
}