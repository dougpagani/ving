package addons

import (
	"github.com/gizak/termui"
	"github.com/yittg/ving/types"
)

// UI out of main board
type UI interface {
	// Render represents how the add-on looks like
	Render() *termui.Row

	// Init add-on state
	Init()

	// Activate  this add-on
	Activate()

	// Deactivate this add-on
	Deactivate()

	// ToggleKey activate/deactivate this add-on
	ToggleKey() string

	// RespondEvents declare events can handle
	RespondEvents() []types.EventMeta

	// HandleKeyEvent handle key event
	HandleKeyEvent(ev termui.Event)

	// ActivateAfterStart represents whether this add-on should be activated after startup
	ActivateAfterStart() bool

	// UpdateState update state
	UpdateState(activeTargets map[int]bool)
}

// ConfirmAware can handle entry event when active
type ConfirmAware interface {
	// OnEnter handle enter event
	OnEnter()
}

// VerticalDirectionAware can handle up/down event when active
type VerticalDirectionAware interface {
	// OnUp handle up event
	OnUp()

	// OnDown handle up event
	OnDown()
}

// HorizontalDirectionAware can handle left/right event when active
type HorizontalDirectionAware interface {
	// OnLeft handle up event
	OnLeft()

	// OnRight handle up event
	OnRight()
}

// DirectionAware can handle all direction event when active
type DirectionAware interface {
	VerticalDirectionAware

	HorizontalDirectionAware
}
