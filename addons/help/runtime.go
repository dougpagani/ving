package help

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/yittg/ving/addons"
	gui "github.com/yittg/ving/ui"
)

type runtime struct {
	msg []string

	ui         *ui
	initUILock sync.Once
}

func newHelp() addons.AddOn {
	return &runtime{}
}

func (h *runtime) Desc() string {
	return "help pane"
}

func (h *runtime) Init(*addons.Envoy) {
}

func (h *runtime) Start(context.Context) {
}

func (h *runtime) Schedule() {
}

func (h *runtime) State() interface{} {
	if h.msg == nil {
		h.msg = append(h.msg, "[Keys](fg-bold)")
		h.msg = append(h.msg, "    [Key](fg-underline)           [Description](fg-underline)")

		for _, gk := range gui.GlobalKeys {
			h.msg = append(h.msg, fmt.Sprintf("    [%-10s](fg-bold)    %s",
				strings.Join(gk.Keys, ","), gk.Description))
		}
		h.msg = append(h.msg, "")
		for _, addOn := range addons.All {
			addOnUI := addOn.GetUI()
			h.msg = append(h.msg,
				fmt.Sprintf("    [%-10s](fg-bold)    %s", addOnUI.ToggleKey(), addOn.Desc()))
			for _, em := range addOnUI.RespondEvents() {
				h.msg = append(h.msg, fmt.Sprintf("    [%-10s](fg-bold)    %s",
					"  "+strings.Join(em.Keys, ","), em.Description))
			}
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
