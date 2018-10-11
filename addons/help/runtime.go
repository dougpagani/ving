package help

import (
	"fmt"
	"sync"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
)

type runtime struct {
	msg []string

	ui         *ui
	initUILock sync.Once
}

func NewHelp() addons.AddOn {
	return &runtime{}
}

func (h *runtime) Desc() string {
	return "help pane"
}

func (h *runtime) Init([]*protocol.NetworkTarget, chan bool, *options.Option, *net.NPing) {
}

func (h *runtime) Start() {
}

func (h *runtime) Collect() {
}

func (h *runtime) Activate() {
}

func (h *runtime) Deactivate() {
}

func (h *runtime) RenderState() interface{} {
	if h.msg == nil {
		h.msg = append(h.msg, "[Keys](fg-bold)")
		h.msg = append(h.msg, "    [Key](fg-underline)           [Description](fg-underline)")

		h.msg = append(h.msg, fmt.Sprintf("    [%-10s](fg-bold)    %s",
			"q,<C-c>", "quit"))
		h.msg = append(h.msg, "")
		for _, addon := range addons.All {
			h.msg = append(h.msg,
				fmt.Sprintf("    [%-10s](fg-bold)    %s", addon.GetUI().ToggleKey(), addon.Desc()))
		}
	}
	return h.msg
}

func (h *runtime) GetUI() addons.UI {
	if h.ui == nil {
		h.initUILock.Do(func() {
			h.ui = newUI(h)
		})
	}
	return h.ui
}