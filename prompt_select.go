package main

import "burntsushi.net/go/x-go-binding/xgb"

import (
    "burntsushi.net/go/xgbutil"
    "burntsushi.net/go/xgbutil/keybind"
    "burntsushi.net/go/xgbutil/xevent"
)

type promptSelect struct {
    showing bool
    selected int
    top *window
    input *textInput
}

func (ps *promptSelect) Id() xgb.Id {
    return ps.top.id
}

func newPromptSelect() *promptSelect {
    top := createWindow(ROOT.id, 0)
    input := renderTextInputCreate(
        top, THEME.prompt.bgColor, THEME.prompt.font, THEME.prompt.fontSize,
        THEME.prompt.fontColor, 500)

    top.change(xgb.CWBackPixel, uint32(THEME.prompt.bgColor))

    ps := &promptSelect{
        showing: false,
        selected: 0,
        top: top,
        input: input,
    }

    xevent.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            if !ps.showing {
                return
            }

            _, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)
            logDebug.Println(keybind.KeycodeString(X, kc))

            if kc == CONF.cancelKey {
                println("Hiding selection prompt")
                ps.hide()
            }

            if kc == CONF.confirmKey {
                println("Confirming selection prompt")
                ps.hide()
            }
    }).Connect(X, X.Dummy())

    return ps
}

func (ps *promptSelect) show() bool {
    // Note that DummyGrab is smart and avoids races. Check it out
    // in xgbutil/keybind.go if you're interested.
    // This makes it impossible to press and release alt-tab too quickly
    // to have it not register.
    if err := keybind.DummyGrab(X); err != nil {
        logWarning.Println("Could not grab keyboard for prompt select: %v", err)
        return false
    }

    // To the top!
    if len(WM.stack) > 0 {
        ps.top.configure(DoSibling | DoStack, 0, 0, 0, 0,
                         WM.stack[0].Frame().ParentId(), xgb.StackModeAbove)
    }

    // get our screen geometry so we can position ourselves
    headGeom := WM.headActive()
    // maxWidth := int(float64(headGeom.Width()) * 0.8) 

    width, height := 500, 500

    // position the damn window based on its width/height (i.e., center it)
    posx := headGeom.Width() / 2 - width / 2
    posy := headGeom.Height() / 2 - height / 2

    // Issue the configure requests. We also need to adjust the borders.
    ps.top.moveresize(DoX | DoY | DoW | DoH, posx, posy, width, height)

    ps.showing = true
    ps.selected = -1
    ps.top.map_()

    return true
}

// hide stops the grab and hides the prompt.
func (ps *promptSelect) hide() {
    ps.top.unmap()
    keybind.DummyUngrab(X)
    ps.showing = false
}

