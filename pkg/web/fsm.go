package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/paulbellamy/ratecounter"
	"gux.codes/omega/pkg/fsm"
	"gux.codes/omega/pkg/utils"
)

const (
	Failure             fsm.StateType = "Failure"
	Idle                fsm.StateType = "Idle"
	Cancelling          fsm.StateType = "Cancelling"
	CapturingFrame      fsm.StateType = "CapturingFrame"
	UpdatingVirtualTime fsm.StateType = "UpdatingVirtualTime"
	Navigating          fsm.StateType = "Navigating"
	Recording           fsm.StateType = "Recording"
	Saving              fsm.StateType = "Saving"
	Success             fsm.StateType = "Success"

	Cancel fsm.EventType = "Cancel"
	Done   fsm.EventType = "Done"
	Init   fsm.EventType = "Init"
	Start  fsm.EventType = "Start"
	Stop   fsm.EventType = "Stop"
	Close  fsm.EventType = "Close"
	Error  fsm.EventType = "Error"
)

type RecordStateContext struct {
	url string
	err error
	virtualTime int
	frames [][]byte
	counter *ratecounter.RateCounter
}

type NavigateContext struct {
	ctx context.Context
	url string
}

type Cleanup struct {}

func (a *Cleanup) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, fsm.Event, error) {
	return stateCtx, fsm.Event{Type: Done}, nil
}

type Navigate struct {}

func (a *Navigate) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, fsm.Event, error) {
	ctx := event.Context.(NavigateContext).ctx
	url := event.Context.(NavigateContext).url
	state := stateCtx.(RecordStateContext)

	err := chromedp.Run(ctx, chromedp.Tasks{chromedp.Navigate(url)})
	if err != nil {
		state.err = err
		return state, fsm.Event{Type: Error}, nil
	}

	state.url = url
	return state, fsm.Event{Type: Done}, nil
}

type CaptureFrame struct {}

func (a *CaptureFrame) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, fsm.Event, error) {
	ctx := event.Context.(context.Context)
	state := stateCtx.(RecordStateContext)

	var frame []byte

	if err := screenshot(1920, 1080, &frame).Do(ctx); err != nil {
		state.err = err
		return state, fsm.Event{Type: Error}, nil
	}

	state.frames = append(state.frames, frame)
	return state, fsm.Event{Type: Done, Context: ctx}, nil
}

type UpdateVirtualTime struct {}

func (a *UpdateVirtualTime) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, fsm.Event, error) {
	ctx := event.Context.(context.Context)
	state := stateCtx.(RecordStateContext)

	virtualTime := state.virtualTime + 16

	var res []byte
	chromedp.Evaluate(fmt.Sprintf("timeweb.goTo(%d)", virtualTime), &res).Do(ctx)

	state.virtualTime = virtualTime
	return stateCtx, fsm.Event{Type: Done}, nil
}

type Dump struct {}

func (a *Dump) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, fsm.Event, error) {
	state := stateCtx.(RecordStateContext)

	bar := pb.StartNew(len(state.frames))

	// Write the files to the filesystem
	for index, value := range state.frames {
		err := ioutil.WriteFile(fmt.Sprintf("/tmp/%06d.png", index), value, 0644)
		if err != nil {
			state.err = err
			return state, fsm.Event{Type: Error}, nil
		}
		// Move the bar forward
		bar.Increment()
	}

	// Close the bar progress
	bar.Finish()

	return stateCtx, fsm.Event{Type: Done}, nil
}

type UpdateFramesCounter struct {}

func (a *UpdateFramesCounter) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, error) {
	state := stateCtx.(RecordStateContext)
	state.counter.Incr(1)

	fmt.Printf("%s %d fps | virtualTime: %d ms\r", utils.BoxBlue("info"), state.counter.Rate(), state.virtualTime)

	return state, nil
}

type ClosedBeforeRecording struct {}

func (a *ClosedBeforeRecording) Execute(stateCtx fsm.StateContext, event fsm.Event) (fsm.StateContext, error) {
	state := stateCtx.(RecordStateContext)

	if event.Type == Close {
		state.err = errors.New("closed the session before recording at state: " + string(event.Type))
	}

	return state, nil
}

type ExitAction struct {
	done chan bool
}

func (a *ExitAction) Execute(stateCtx fsm.StateContext, event fsm.Event) error {
	state := stateCtx.(RecordStateContext)

	if state.err != nil {
		utils.Error("recording failure: " + state.err.Error())
	} else {
		utils.Success("recording successful")
	}

	a.done <- true
	return nil
}


func StartRecording() error {
	// Create the parent context.
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// Create the state machine
	stateMachine := NewRecordStateMachine()

	// Set up a channel so we can block later while we get each animation frame
	done := make(chan bool, 1)

	// Setup our Ctrl+C handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func(){
		<- signals
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		fmt.Println("- Cleaning up...")
		if err := stateMachine.SendEvent(fsm.Event{Type: Cancel, Context: ctx}); err != nil {
			utils.Error(err.Error())
			done <- true
		}
		<- time.After(3 * time.Second)
		os.Exit(0)
	}()

	// Run the web server
	webServerOptions := NewWebServerOptions()
	go startWebServer(*webServerOptions)

	// Start to listen to the console.log() messages
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			if message, err := parseArgs(ev.Args); err == nil {
				switch message.Action {
				case "start":
					if err := stateMachine.SendEvent(fsm.Event{Type: Start, Context: ctx}); err != nil {
						utils.Error(err.Error())
						done <- true
					}
				case "stop":
					if err := stateMachine.SendEvent(fsm.Event{Type: Stop, Context: ctx}); err != nil {
						utils.Error(err.Error())
						done <- true
					}
				case "done":
					if err := stateMachine.SendEvent(fsm.Event{Type: Close, Context: ctx}); err != nil {
						utils.Error(err.Error())
						done <- true
					}
				}
			}
		}
	})

  // Setup the ExitAction for the success and failure states
	if state, ok := stateMachine.States[Success]; ok {
		state.Action = &ExitAction{done: done}
	}
	if state, ok := stateMachine.States[Failure]; ok {
		state.Action = &ExitAction{done: done}
	}

	// Navigate to the handler function
	url := fmt.Sprintf("http://localhost:%d/handler", webServerOptions.Port)
	init := NavigateContext{ctx: ctx, url: url}
	if err := stateMachine.SendEvent(fsm.Event{Type: Init, Context: init}); err != nil {
		utils.Error(err.Error())
		done <- true
	}

	<- done

	return stateMachine.Context.(RecordStateContext).err
}

func parseArgs(args []*runtime.RemoteObject) (*ConsoleMessage, error) {
	message := &ConsoleMessage{}

	if len(args) == 0 {
		return message, errors.New("empty args slice")
	}

	arg := args[0]
	data := strings.ReplaceAll(fmt.Sprintf("%s", arg.Value[1 : len(arg.Value) -1]), "\\", "")
	err := json.Unmarshal([]byte(data), &message);

	return message, err
}

func NewRecordStateMachine() *fsm.StateMachine {
	return &fsm.StateMachine{
		Context: RecordStateContext{
			virtualTime: 0,
			counter: ratecounter.NewRateCounter(1 * time.Second),
		},
		States: fsm.States{
			fsm.Default: fsm.State{
				ExitAction: &ClosedBeforeRecording{},
				Events: fsm.Events{
					Init: Navigating,
					Cancel: Cancelling,
					Close: Failure,
				},
			},
			Navigating: fsm.State{
				EntryAction: &Navigate{},
				ExitAction: &ClosedBeforeRecording{},
				Events: fsm.Events{
					Done: Idle,
					Start: CapturingFrame,
					Cancel: Cancelling,
					Error: Failure,
					Close: Failure,
				},
			},
			CapturingFrame: fsm.State{
				EntryAction: &CaptureFrame{},
				Events: fsm.Events{
					Done: UpdatingVirtualTime,
					Stop: Idle,
					Cancel: Cancelling,
					Error: Failure,
					Close: Saving,
				},
			},
			UpdatingVirtualTime: fsm.State{
				EntryAction: &UpdateVirtualTime{},
				ExitAction: &UpdateFramesCounter{},
				Events: fsm.Events{
					Done: CapturingFrame,
					Stop: Idle,
					Cancel: Cancelling,
					Error: Failure,
					Close: Saving,
				},
			},
			Idle: fsm.State{
				Events: fsm.Events{
					Start: Recording,
					Cancel: Cancelling,
					Close: Saving,
					Error: Failure,
				},
			},
			Cancelling: fsm.State{
				EntryAction: &Cleanup{},
				Events: fsm.Events{
					Done: Failure,
				},
			},
			Saving: fsm.State{
				EntryAction: &Dump{},
				Events: fsm.Events{
					Cancel: Cancelling,
					Done: Success,
					Error: Failure,
				},
			},
			Success: fsm.State{
				Events: fsm.Events{},
			},
			Failure: fsm.State{
				Events: fsm.Events{},
			},
		},
	}
}