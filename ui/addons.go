package ui

import (
	"github.com/gizak/termui"
	"github.com/yittg/ving/statistic"
)

// AddOn out of main board
type AddOn interface {
	// Render represents how the add-on looks like
	Render() *termui.Row

	// Init add-on state
	Init()

	// Reset add-on state
	Reset()

	// ToggleKey activate/deactivate this add-on
	ToggleKey() string

	// RespondEvents declare events can handle
	RespondEvents() []string

	// HandleKeyEvent handle key event
	HandleKeyEvent(ev termui.Event)

	// ActivateAfterStart represents whether this add-on should be activated after startup
	ActivateAfterStart() bool

	// UpdateState update state
	UpdateState(sts []*statistic.Detail, state interface{})
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
