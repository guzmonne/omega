package fsm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// LIGHTSWITCH
const (
	Off	StateType = "Off"
	On	StateType = "On"

	SwitchOff	EventType = "SwitchOff"
	SwitchOn	EventType = "SwitchOn"
)

type OffAction struct {}
type OnAction struct {}

func (a *OffAction) Execute(stateCtx StateContext, event Event) error {
	fmt.Println("OffAction")
	return nil
}

func (a *OnAction) Execute(stateCtx StateContext, event Event) error {
	fmt.Println("OnAction")
	return nil
}
// ORDER PROCESSING
var ErrInvalidCardNumber = errors.New("card number is invalid")
var ErrInsufficientNumberOfItems = errors.New("insufficient number of items in order")

const (
	CreatingOrder			StateType = "CreatingOrder"
	OrderFailed				StateType = "OrderFailed"
	OrderPlaced				StateType = "OrderPlaced"
	ChargingCard			StateType = "ChargingCard"
	TransactionFailed	StateType = "TransactionFailed"
	OrderShipped			StateType = "OrderShipped"

	CreateOrder				EventType = "CreateOrder"
	FailOrder					EventType = "FailOrder"
	PlaceOrder				EventType = "PlaceOrder"
	ChargeCard				EventType = "ChargeCard"
	FailTransaction		EventType = "FailTransaction"
	ShipOrder					EventType = "ShipOrder"
)

//
type OrderShipment struct {
	cardNumber		string
	address				string
}

func (o *OrderShipment) String() string {
	return fmt.Sprintf(
		"OrderShipment { cardNumber: %s, address: %s }",
		o.cardNumber,
		o.address,
	)
}

type OrderStateContext struct {
	items				[]string
	address 		string
	cardNumber	string
	history     []EventType
	err					error
}

func (state *OrderStateContext) String() string {
	return fmt.Sprintf(
		"OrderStateContext { items: [ %s ], address: %s, cardNumber: %s, err: %v }",
		strings.Join(state.items, ","),
		state.address,
		state.cardNumber,
		state.err,
	)
}

type LogEventAction struct {}

func (a *LogEventAction) Execute(stateCtx StateContext, event Event) (StateContext, error) {
	state := stateCtx.(OrderStateContext)

	fmt.Printf("Event %v\n", event.Type)

	state.history = append(state.history, event.Type)
	return state, nil
}

type CreatingOrderContext struct {
	items []string
}

type CreatingOrderAction struct {}

func (a *CreatingOrderAction) Execute(stateCtx StateContext, event Event) (StateContext, Event, error) {
	state := stateCtx.(OrderStateContext)
	order := event.Context.(CreatingOrderContext)

	fmt.Println("Validating order:", order)

	if len(order.items) == 0 {
		state.err = ErrInsufficientNumberOfItems
		return state, Event{Type: FailOrder, Context: order}, nil
	}

	state.items = order.items
	state.err = nil
	return state, Event{Type: PlaceOrder, Context: order}, nil
}

type OrderFailedAction struct {}

func (a *OrderFailedAction) Execute(stateCtx StateContext, event Event) error {
	state := stateCtx.(OrderStateContext)

	fmt.Println("Order failed; err:", state.err)
	return nil
}

type OrderPlacedAction struct {}

func (a *OrderPlacedAction) Execute(stateCtx StateContext, event Event) error {
	state := stateCtx.(OrderStateContext)

	fmt.Println("Order placed; items:", state.items)
	return nil
}

type ChargingCardAction struct {}

type ChargingCardContext struct {
	cardNumber string
	address string
}

func (a *ChargingCardAction) Execute(stateCtx StateContext, event Event) (StateContext, Event, error) {
	state := stateCtx.(OrderStateContext)
	order := event.Context.(ChargingCardContext)

	fmt.Println("Validating card; cardNumber:", order.cardNumber)

	if order.cardNumber == "" {
		state.err = ErrInvalidCardNumber
		return state, Event{Type: FailTransaction, Context: order}, nil
	}

	state.address = order.address
	state.cardNumber = order.cardNumber
	state.err = nil
	return state, Event{Type: ShipOrder, Context: order}, nil
}

type TransactionFailedAction struct {}

func (a *TransactionFailedAction) Execute(stateCtx StateContext, event Event) error {
	state := stateCtx.(OrderStateContext)

	fmt.Println("Transaction failed; err:", state.err)
	return nil
}

type OrderShippedAction struct {}

func (a *OrderShippedAction) Execute(stateCtx StateContext, event Event) error {
	state := stateCtx.(OrderStateContext)

	fmt.Println("Order shipped; address:", state.address)
	return nil
}

// FSM Suite
type FSMSuite struct {
	suite.Suite
	s *StateMachine
	o *StateMachine
}

func (suite *FSMSuite) newLightSwitch() *StateMachine {
	return &StateMachine{
		States: States{
			Default: State{
				Events: Events{
					SwitchOff: Off,
				},
			},
			Off: State{
				Action: &OffAction{},
				Events: Events{
					SwitchOn: On,
				},
			},
			On: State{
				Action: &OnAction{},
				Events: Events{
					SwitchOff: Off,
				},
			},
		},
	}
}

func (suite *FSMSuite) newOrderFSM() *StateMachine {
	return &StateMachine{
		Context: OrderStateContext{},
		States: States{
			Default: State{
				ExitAction: &LogEventAction{},
				Events: Events{
					CreateOrder: CreatingOrder,
				},
			},
			CreatingOrder: State{
				ExitAction: &LogEventAction{},
				EntryAction: &CreatingOrderAction{},
				Events: Events{
					FailOrder: OrderFailed,
					PlaceOrder: OrderPlaced,
				},
			},
			OrderFailed: State{
				ExitAction: &LogEventAction{},
				Action: &OrderFailedAction{},
				Events: Events{
					CreateOrder: CreatingOrder,
				},
			},
			OrderPlaced: State{
				ExitAction: &LogEventAction{},
				Action: &OrderPlacedAction{},
				Events: Events{
					ChargeCard: ChargingCard,
				},
			},
			ChargingCard: State{
				ExitAction: &LogEventAction{},
				EntryAction: &ChargingCardAction{},
				Events: Events{
					FailTransaction: TransactionFailed,
					ShipOrder: OrderShipped,
				},
			},
			TransactionFailed: State{
				ExitAction: &LogEventAction{},
				Action: &TransactionFailedAction{},
				Events: Events{
					ChargeCard: ChargingCard,
				},
			},
			OrderShipped: State{
				ExitAction: &LogEventAction{},
				Action: &OrderShippedAction{},
			},
		},
	}
}

func (suite *FSMSuite) SetupSuite() {
	suite.s = suite.newLightSwitch()
	suite.o = suite.newOrderFSM()
}

func (suite *FSMSuite) TestLightSwitchFSM() {
	var err error

	SwitchOffEvent := Event{Type: SwitchOff}
	SwitchOnEvent := Event{Type: SwitchOn}

	// Set the initial "off" state in the state machine.
	err = suite.s.SendEvent(SwitchOffEvent)
	suite.NoError(err, "Couldn't set the initial state of the machine; err: %v", err)

	// Send the switch-off event again and expect the state machine to return an error.
	err = suite.s.SendEvent(SwitchOffEvent)
	suite.EqualError(err, ErrEventRejected.Error(), "Expected the event rejexted error; got: %v", err)

	// Send the switch-on event and expect the state machine to transition to the "on" state.
	err = suite.s.SendEvent(SwitchOnEvent)
	suite.NoError(err, "Couldn't switch the light on; err: %v", err)

	// Send the switch-on event again and expect the state machine to return an error.
	err = suite.s.SendEvent(SwitchOnEvent)
	suite.EqualError(err, ErrEventRejected.Error(), "Expected the event rejexted error; got: %v", err)

	// Send the switch-off event and expect the state machine to transition back to the "off" state.
	err = suite.s.SendEvent(SwitchOffEvent)
	suite.NoError(err, "Couldn't switch the light off; err: %v", err)
}

func (suite *FSMSuite) TestOrderFSM() {
	var event Event
	var err error

	fsm := suite.o

	// Try to create an order with an invalid set of items.
	event = Event{Type: CreateOrder, Context: CreatingOrderContext{items: []string{}}}
	err = fsm.SendEvent(event)
	suite.NoError(err, "Should not return an error")
	suite.EqualError(fsm.Context.(OrderStateContext).err, ErrInsufficientNumberOfItems.Error(), "Failed to send create order event; err: %v", err)

	// The state machine should enter the OrderFail state.
	suite.Equal(OrderFailed, fsm.Current, "Expected the FSM to be in the OrderFailed state; actual: %s", fsm.Current)

	// This fsm context history should be of length 2 and its last item should be the last event
	suite.Len(fsm.Context.(OrderStateContext).history, 2)
	suite.Equal(fsm.Context.(OrderStateContext).history[0], event.Type)
	suite.Equal(fsm.Context.(OrderStateContext).history[1], FailOrder)

	// Retry the order with valid items.
	event = Event{Type: CreateOrder, Context: CreatingOrderContext{items: []string{"foo", "bar"}}}
	err = fsm.SendEvent(event)
	suite.NoError(err, "Should not return an error")
	suite.NoError(fsm.Context.(OrderStateContext).err, "Failed to send create order; err: %v", err)

	// This fsm context history should be of length 4 and its last item should be the last event
	suite.Len(fsm.Context.(OrderStateContext).history, 4)
	suite.Equal(fsm.Context.(OrderStateContext).history[2], event.Type)
	suite.Equal(fsm.Context.(OrderStateContext).history[3], PlaceOrder)

	// The state machine should enter the OrderPlaced state.
	suite.Equal(OrderPlaced, fsm.Current, "Expected the FSM to be in the OrderPlaced state; actual: %s", fsm.Current)

	// Try to charge the card using an invalid card number.
	event = Event{Type: ChargeCard, Context: ChargingCardContext{cardNumber: "", address: "123 Foo Street, Bar Baz, QU 12345, USA"}}
	err = fsm.SendEvent(event)
	suite.NoError(err, "Should not return an error")
	suite.EqualError(fsm.Context.(OrderStateContext).err, ErrInvalidCardNumber.Error(), "Failed to send charge card event; err: %v", err)

	// This fsm context history should be of length 6 and its last item should be the last event
	suite.Len(fsm.Context.(OrderStateContext).history, 6)
	suite.Equal(fsm.Context.(OrderStateContext).history[4], event.Type)
	suite.Equal(fsm.Context.(OrderStateContext).history[5], FailTransaction)

	// The state should enter the TransactionFailed state.
	suite.Equal(TransactionFailed, fsm.Current, "Expected the FSM to be in the TransactionFailed state; actual: %s", fsm.Current)

	// Retry the card transaction
	event = Event{Type: ChargeCard, Context: ChargingCardContext{cardNumber: "0000-0000-0000-0000", address: "123 Foo Street, Bar Baz, QU 12345, USA"}}
	err = fsm.SendEvent(event)
	suite.NoError(err, "Should not return an error")
	suite.NoError(err, "Failed to send charge card event; err: %v", err)

	// This fsm context history should be of length 8 and its last item should be the last event
	suite.Len(fsm.Context.(OrderStateContext).history, 8)
	suite.Equal(fsm.Context.(OrderStateContext).history[6], event.Type)
	suite.Equal(fsm.Context.(OrderStateContext).history[7], ShipOrder)

	// The state machine should enter the OrderShipped state.
	suite.Equal(OrderShipped, fsm.Current, "Expected the FSM to be in the OrderShipped state; actual: %s", fsm.Current)

	// Check that the order can't be charged more than once.
	err = fsm.SendEvent(event)
	suite.EqualError(err, ErrEventRejected.Error(), "Expected the FSM to return a rejected event error; actual %v", err)

	// The fsm context shouldn't have been updated.
	suite.Len(fsm.Context.(OrderStateContext).history, 8)
	suite.Equal(fsm.Context.(OrderStateContext).history[7], ShipOrder)
}

func TestFSMSuite(t *testing.T) {
	suite.Run(t, new(FSMSuite))
}
