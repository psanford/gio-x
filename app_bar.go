package materials

import (
	"image"
	"image/color"
	"sync"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var moreIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.NavigationMoreVert)
	return icon
}()

var cancelIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentClear)
	return icon
}()

// AppBar implements the material design App Bar documented here:
// https://material.io/components/app-bars-top
//
// TODO(whereswaldon): implement support for RTL layouts
type AppBar struct {
	// init ensures that AppBars constructed using struct literal
	// syntax still have their fields initialized before use.
	init sync.Once

	*material.Theme

	NavigationButton       widget.Clickable
	NavigationIcon         *widget.Icon
	Title, ContextualTitle string

	normalActions, contextualActions actionGroup
	overflowMenu
	contextualAnim VisibilityAnimation
}

// actionGroup is a logical set of actions that might be displayed
// by an App Bar.
type actionGroup struct {
	actions           []AppBarAction
	actionAnims       []VisibilityAnimation
	overflow          []OverflowAction
	overflowState     []widget.Clickable
	lastOverflowCount int
}

func (a *actionGroup) setActions(actions []AppBarAction, overflows []OverflowAction) {
	a.actions = actions
	a.actionAnims = make([]VisibilityAnimation, len(actions))
	for i := range a.actionAnims {
		a.actionAnims[i].Duration = actionAnimationDuration
	}
	a.overflow = overflows
	a.overflowState = make([]widget.Clickable, len(a.actions)+len(a.overflow))
}

func (a *actionGroup) layout(gtx C, th *material.Theme, overflowBtn *widget.Clickable, background color.RGBA) D {
	overflowedActions := len(a.actions)
	gtx.Constraints.Min.Y = 0
	widthDp := float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp
	visibleActionItems := int((widthDp / 48) - 1)
	if visibleActionItems < 0 {
		visibleActionItems = 0
	}
	visibleActionItems = min(visibleActionItems, len(a.actions))
	overflowedActions -= visibleActionItems
	var actions []layout.FlexChild
	for i := range a.actions {
		action := a.actions[i]
		anim := &a.actionAnims[i]
		switch anim.State {
		case Visible:
			if i >= visibleActionItems {
				anim.Disappear(gtx.Now)
			}
		case Invisible:
			if i < visibleActionItems {
				anim.Appear(gtx.Now)
			}
		}
		actions = append(actions, layout.Rigid(func(gtx C) D {
			return action.layout(background, th.Color.InvText, anim, gtx)
		}))
	}
	if len(a.overflow)+overflowedActions > 0 {
		actions = append(actions, layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			btn := material.IconButton(th, overflowBtn, moreIcon)
			btn.Size = unit.Dp(24)
			btn.Background = background
			btn.Inset = layout.UniformInset(unit.Dp(6))
			return overflowButtonInset.Layout(gtx, btn.Layout)
		}))
	}
	a.lastOverflowCount = overflowedActions
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, actions...)
}

// overflowMenu holds the state for an overflow menu in an app bar.
type overflowMenu struct {
	VisibilityAnimation
	scrim Scrim
	list  layout.List
	// the button that triggers the overflow menu
	widget.Clickable
	selectedTag interface{}
}

func (o *overflowMenu) updateState(gtx layout.Context, actions *actionGroup) {
	if o.Clicked() && !o.Visible() {
		o.Appear(gtx.Now)
	}
	if o.scrim.Clicked() {
		o.Disappear(gtx.Now)
	}
	for i := range actions.overflowState {
		if actions.overflowState[i].Clicked() {
			o.Disappear(gtx.Now)
			o.selectedTag = o.actionForIndex(i, actions).Tag
		}
	}
}

func (o overflowMenu) overflowLen(actions *actionGroup) int {
	return len(actions.overflow) + actions.lastOverflowCount
}

func (o overflowMenu) actionForIndex(index int, actions *actionGroup) OverflowAction {
	if index < actions.lastOverflowCount {
		return actions.actions[len(actions.actions)-actions.lastOverflowCount+index].OverflowAction
	}
	return actions.overflow[index-actions.lastOverflowCount]
}

func (o *overflowMenu) layoutOverflow(gtx C, th *material.Theme, actions *actionGroup) D {
	o.selectedTag = nil
	o.updateState(gtx, actions)
	if !o.Visible() {
		return layout.Dimensions{}
	}
	o.scrim.Layout(gtx, &o.VisibilityAnimation)
	defer op.Push(gtx.Ops).Pop()
	width := gtx.Constraints.Max.X / 2
	gtx.Constraints.Min.X = width
	op.Offset(f32.Pt(float32(width), 0)).Add(gtx.Ops)
	var menuMacro op.MacroOp
	menuMacro = op.Record(gtx.Ops)
	dims := layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			gtx.Constraints.Min.X = width
			paintRect(gtx, gtx.Constraints.Min, th.Color.Hint)
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx C) D {
			return o.list.Layout(gtx, o.overflowLen(actions), func(gtx C, index int) D {
				action := o.actionForIndex(index, actions)
				state := &actions.overflowState[index]
				return material.Clickable(gtx, state, func(gtx C) D {
					gtx.Constraints.Min.X = width
					return layout.Inset{
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
						Left:   unit.Dp(8),
					}.Layout(gtx, func(gtx C) D {
						label := material.Label(th, unit.Dp(18), action.Name)
						label.MaxLines = 1
						return label.Layout(gtx)
					})
				})
			})
		}),
	)
	menuOp := menuMacro.Stop()
	progress := o.Revealed(gtx)
	maxWidth := dims.Size.X
	rect := clip.Rect{
		Max: image.Point{
			X: maxWidth,
			Y: int(float32(dims.Size.Y) * progress),
		},
		Min: image.Point{
			X: maxWidth - int(float32(dims.Size.X)*progress),
			Y: 0,
		},
	}
	rect.Add(gtx.Ops)
	menuOp.Add(gtx.Ops)
	return dims
}

// NewAppBar creates and initializes an App Bar.
func NewAppBar(th *material.Theme) *AppBar {
	ab := &AppBar{
		Theme: th,
	}
	ab.initialize()
	return ab
}

func (a *AppBar) initialize() {
	a.init.Do(func() {
		a.overflowMenu.list.Axis = layout.Vertical
		a.overflowMenu.State = Invisible
		a.contextualAnim.State = Invisible
		a.overflowMenu.Duration = overflowAnimationDuration
		a.contextualAnim.Duration = contextualAnimationDuration
		a.overflowMenu.scrim.FinalAlpha = 82
	})
}

// AppBarAction configures an action in the App Bar's action items.
// The state and icon should not be nil.
type AppBarAction struct {
	OverflowAction
	Layout func(gtx layout.Context, bg, fg color.RGBA) layout.Dimensions
}

// SimpleIconAction configures an AppBarAction that functions as a simple
// IconButton. To receive events from the button, use the standard methods
// on the provided state parameter.
func SimpleIconAction(th *material.Theme, state *widget.Clickable, icon *widget.Icon, overflow OverflowAction) AppBarAction {
	a := AppBarAction{
		OverflowAction: overflow,
		Layout: func(gtx C, bg, fg color.RGBA) D {
			btn := SimpleIconButton(th, state, icon)
			btn.Background = bg
			btn.Color = fg
			return btn.Layout(gtx)
		},
	}
	return a
}

// SimpleIconButton creates an IconButtonStyle that is pre-configured to
// be the right size for use as an AppBarAction
func SimpleIconButton(th *material.Theme, state *widget.Clickable, icon *widget.Icon) material.IconButtonStyle {
	btn := material.IconButton(th, state, icon)
	btn.Size = unit.Dp(24)
	btn.Inset = layout.UniformInset(unit.Dp(12))
	return btn
}

const (
	actionAnimationDuration     = time.Millisecond * 250
	contextualAnimationDuration = time.Millisecond * 250
	overflowAnimationDuration   = time.Millisecond * 250
)

var actionButtonInset = layout.Inset{
	Top:    unit.Dp(4),
	Bottom: unit.Dp(4),
}

func (a AppBarAction) layout(bg, fg color.RGBA, anim *VisibilityAnimation, gtx layout.Context) layout.Dimensions {
	if !anim.Visible() {
		return layout.Dimensions{}
	}
	animating := anim.Animating()
	var macro op.MacroOp
	if animating {
		macro = op.Record(gtx.Ops)
	}
	if !animating {
		return a.Layout(gtx, bg, fg)
	}
	dims := actionButtonInset.Layout(gtx, func(gtx C) D {
		return a.Layout(gtx, bg, fg)
	})
	btnOp := macro.Stop()
	progress := anim.Revealed(gtx)
	dims.Size.X = int(progress * float32(dims.Size.X))
	// ensure this clip transformation stays local to this function
	defer op.Push(gtx.Ops).Pop()

	clip.Rect{
		Max: dims.Size,
	}.Add(gtx.Ops)
	btnOp.Add(gtx.Ops)
	return dims
}

var overflowButtonInset = layout.Inset{
	Top:    unit.Dp(10),
	Bottom: unit.Dp(10),
}

// OverflowAction is an action that is always in the overflow menu.
type OverflowAction struct {
	Name string
	Tag  interface{}
}

func Interpolate(a, b color.RGBA, progress float32) color.RGBA {
	var out color.RGBA
	out.R = uint8(int16(a.R) - int16(float32(int16(a.R)-int16(b.R))*progress))
	out.G = uint8(int16(a.G) - int16(float32(int16(a.G)-int16(b.G))*progress))
	out.B = uint8(int16(a.B) - int16(float32(int16(a.B)-int16(b.B))*progress))
	out.A = uint8(int16(a.A) - int16(float32(int16(a.A)-int16(b.A))*progress))
	return out
}

// Layout renders the app bar. It will span all available horizontal
// space (gtx.Constraints.Max.X), but has a fixed height.
func (a *AppBar) Layout(gtx layout.Context) layout.Dimensions {
	a.initialize()
	originalMaxY := gtx.Constraints.Max.Y
	gtx.Constraints.Max.Y = gtx.Px(unit.Dp(56))
	fill := a.Theme.Color.Primary
	actionSet := &a.normalActions
	if a.contextualAnim.Visible() {
		if !a.contextualAnim.Animating() {
			fill = a.Theme.Color.Text
			fill.A = 255
		} else {
			fill = Interpolate(fill, a.Theme.Color.Text, a.contextualAnim.Revealed(gtx))
		}
		actionSet = &a.contextualActions
	}
	paintRect(gtx, gtx.Constraints.Max, fill)

	layout.Flex{
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if a.NavigationIcon == nil {
				return layout.Dimensions{}
			}
			icon := a.NavigationIcon
			if a.contextualAnim.Visible() {
				icon = cancelIcon
			}
			button := material.IconButton(a.Theme, &a.NavigationButton, icon)
			button.Size = unit.Dp(24)
			button.Background = fill
			button.Inset = layout.UniformInset(unit.Dp(16))
			return button.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx C) D {
				titleText := a.Title
				if a.contextualAnim.Visible() {
					titleText = a.ContextualTitle
				}
				title := material.Body1(a.Theme, titleText)
				title.Color = a.Theme.Color.InvText
				title.TextSize = unit.Dp(18)
				return title.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			return layout.E.Layout(gtx, func(gtx C) D {
				return actionSet.layout(gtx, a.Theme, &a.overflowMenu.Clickable, fill)
			})
		}),
	)
	{
		gtx := gtx
		gtx.Constraints.Max.Y = originalMaxY
		a.overflowMenu.layoutOverflow(gtx, a.Theme, actionSet)
	}
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NavigationClicked returns true when the navigation button has been
// clicked in the last frame.
func (a *AppBar) NavigationClicked(gtx layout.Context) bool {
	clicked := a.NavigationButton.Clicked()
	if clicked && a.contextualAnim.Visible() {
		a.contextualAnim.Disappear(gtx.Now)
		return false
	}
	return clicked
}

// OverflowActionClicked returns whether an overflow item was selected
// during the last frame
func (a *AppBar) OverflowActionClicked() bool {
	return a.overflowMenu.selectedTag != nil
}

// SelectedOverflowAction returns the Tag of the overflow action selected
// during the last frame (if any).
func (a *AppBar) SelectedOverflowAction() interface{} {
	return a.overflowMenu.selectedTag
}

// SetActions configures the set of actions available in the
// action item area of the app bar. They will be displayed
// in the order provided (from left to right) and will
// collapse into the overflow menu from right to left. The
// provided OverflowActions will always be in the overflow
// menu in the order provided.
func (a *AppBar) SetActions(actions []AppBarAction, overflows []OverflowAction) {
	a.normalActions.setActions(actions, overflows)
}

// SetContextualActions configures the actions that should be available in
// the next Contextual mode that this action bar enters.
func (a *AppBar) SetContextualActions(actions []AppBarAction, overflows []OverflowAction) {
	a.contextualActions.setActions(actions, overflows)
}

// StartContextual causes the AppBar to transform into a contextual
// App Bar with a different set of actions than normal. If the App Bar
// is already in contextual mode, this has no effect. The title will
// be used as the contextual app bar title.
func (a *AppBar) StartContextual(when time.Time, title string) {
	if !a.contextualAnim.Visible() {
		a.contextualAnim.Appear(when)
		a.ContextualTitle = title
	}
}

// StopContextual causes the AppBar to stop showing a contextual menu
// if one is currently being displayed.
func (a *AppBar) StopContextual(when time.Time) {
	if a.contextualAnim.Visible() {
		a.contextualAnim.Disappear(when)
	}
}

// ToggleContextual switches between contextual an noncontextual mode.
// If it transitions to contextual mode, the provided title is used.
func (a *AppBar) ToggleContextual(when time.Time, title string) {
	if !a.contextualAnim.Visible() {
		a.StartContextual(when, title)
	} else {
		a.StopContextual(when)
	}
}

// CloseOverflowMenu requests that the overflow menu be closed if it is
// open.
func (a *AppBar) CloseOverflowMenu(when time.Time) {
	if a.overflowMenu.Visible() {
		a.overflowMenu.Disappear(when)
	}
}
