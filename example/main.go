package main

import (
	"flag"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"git.sr.ht/~whereswaldon/materials"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var MenuIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.NavigationMenu)
	return icon
}()

var HomeIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionHome)
	return icon
}()

var SettingsIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionSettings)
	return icon
}()

var OtherIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionHelp)
	return icon
}()

var HeartIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ActionFavorite)
	return icon
}()

var PlusIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentAdd)
	return icon
}()

var barOnBottom bool

func main() {
	flag.BoolVar(&barOnBottom, "bottom-bar", false, "place the app bar on the bottom of the screen instead of the top")
	flag.Parse()
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func LayoutAppBarPage(gtx C) D {
	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, material.Body1(th, `The app bar widget provides a consistent interface element for triggering navigation and page-specific actions.

The controls below allow you to see the various features available in our App Bar implementation.`).Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Contextual App Bar").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if bar.OverflowActionClicked() {
						log.Printf("Overflow clicked: %v", bar.SelectedOverflowAction())
					}
					if contextBtn.Clicked() {
						bar.SetContextualActions(
							[]materials.AppBarAction{
								materials.SimpleIconAction(th, &red, HeartIcon,
									materials.OverflowAction{
										Name: "House",
										Tag:  &red,
									},
								),
							},
							[]materials.OverflowAction{
								{
									Name: "foo",
									Tag:  &blue,
								},
								{
									Name: "bar",
									Tag:  &green,
								},
							},
						)
						bar.ToggleContextual(gtx.Now, "Contextual Title")
					}
					return material.Button(th, &contextBtn, "Trigger").Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Bottom App Bar").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if bottomBar.Changed() {
						if bottomBar.Value {
							nav.Anchor = materials.Bottom
						} else {
							nav.Anchor = materials.Top
						}
					}

					return inset.Layout(gtx, material.Switch(th, &bottomBar).Layout)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Custom Navigation Icon").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if customNavIcon.Changed() {
						if customNavIcon.Value {
							bar.NavigationIcon = HomeIcon
						} else {
							bar.NavigationIcon = MenuIcon
						}
					}
					return inset.Layout(gtx, material.Switch(th, &customNavIcon).Layout)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Animated Resize").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body2(th, "Resize the width of your screen to see app bar actions collapse into or emerge from the overflow menu (as size permits).").Layout)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Custom Action Buttons").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if heartBtn.Clicked() {
						favorited = !favorited
					}
					return inset.Layout(gtx, material.Body2(th, "Click the heart action to see custom button behavior.").Layout)
				}),
			)
		}),
	)
}

func LayoutNavDrawerPage(gtx C) D {
	return layout.Flex{
		Alignment: layout.Middle,
		Axis:      layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, material.Body1(th, `The nav drawer widget provides a consistent interface element for navigation.

The controls below allow you to see the various features available in our Navigation Drawer implementation.`).Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return inset.Layout(gtx, material.Body1(th, "Use non-modal drawer").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					if nonModalDrawer.Changed() {
						if nonModalDrawer.Value {
							navAnim.Appear(gtx.Now)
						} else {
							navAnim.Disappear(gtx.Now)
						}
					}
					return inset.Layout(gtx, material.Switch(th, &nonModalDrawer).Layout)
				}),
			)
		}),
	)
}

type Page struct {
	layout func(layout.Context) layout.Dimensions
	materials.NavItem
	Actions  []materials.AppBarAction
	Overflow []materials.OverflowAction
}

var (
	// initialize modal layer to draw modal components
	modal   = materials.NewModal()
	navAnim = materials.VisibilityAnimation{
		Duration: time.Millisecond * 100,
		State:    materials.Invisible,
	}
	nav      = materials.NewNav(th, "Navigation Drawer", "This is an example.")
	modalNav = materials.ModalNavFrom(&nav, modal)

	bar = materials.NewAppBar(th, modal)

	inset = layout.UniformInset(unit.Dp(8))
	th    = material.NewTheme(gofont.Collection())

	heartBtn, plusBtn, exampleOverflowState widget.Clickable
	red, green, blue                        widget.Clickable
	contextBtn                              widget.Clickable
	bottomBar                               widget.Bool
	customNavIcon                           widget.Bool
	nonModalDrawer                          widget.Bool
	favorited                               bool

	pages = []Page{
		Page{
			NavItem: materials.NavItem{
				Name: "App Bar Features",
				Icon: HomeIcon,
			},
			layout: LayoutAppBarPage,
			Actions: []materials.AppBarAction{
				materials.AppBarAction{
					OverflowAction: materials.OverflowAction{
						Name: "Favorite",
						Tag:  &heartBtn,
					},
					Layout: func(gtx layout.Context, bg, fg color.RGBA) layout.Dimensions {
						btn := materials.SimpleIconButton(th, &heartBtn, HeartIcon)
						btn.Background = bg
						if favorited {
							btn.Color = color.RGBA{R: 200, A: 255}
						} else {
							btn.Color = fg
						}
						return btn.Layout(gtx)
					},
				},
				materials.SimpleIconAction(th, &plusBtn, PlusIcon,
					materials.OverflowAction{
						Name: "Create",
						Tag:  &plusBtn,
					},
				),
			},
			Overflow: []materials.OverflowAction{
				{
					Name: "Example 1",
					Tag:  &exampleOverflowState,
				},
				{
					Name: "Example 2",
					Tag:  &exampleOverflowState,
				},
			},
		},
		Page{
			NavItem: materials.NavItem{
				Name: "Nav Drawer Features",
				Icon: SettingsIcon,
			},
			layout: LayoutNavDrawerPage,
		},
		Page{
			NavItem: materials.NavItem{
				Name: "About this library",
				Icon: OtherIcon,
			},
			layout: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H3(th, "Elsewhere").Layout(gtx)
					}),
				)
			},
			Actions: []materials.AppBarAction{
				materials.SimpleIconAction(th, &heartBtn, HeartIcon,
					materials.OverflowAction{
						Name: "Favorite",
						Tag:  &heartBtn,
					},
				),
			},
		},
	}
)

func loop(w *app.Window) error {
	var ops op.Ops

	bar.NavigationIcon = MenuIcon
	if barOnBottom {
		bar.Anchor = materials.Bottom
		nav.Anchor = materials.Bottom
	}

	// assign navigation tags and configure navigation bar with all pages
	for i, page := range pages {
		page.NavItem.Tag = i
		nav.AddNavItem(page.NavItem)
	}

	// configure app bar initial state
	page := pages[nav.CurrentNavDestination().(int)]
	bar.Title = page.Name
	bar.SetActions(page.Actions, page.Overflow)

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			if bar.NavigationClicked(gtx) {
				if nonModalDrawer.Value {
					navAnim.ToggleVisibility(gtx.Now)
				} else {
					modalNav.Appear(gtx.Now)
					navAnim.Disappear(gtx.Now)
				}
			}
			if nav.NavDestinationChanged() {
				page := pages[nav.CurrentNavDestination().(int)]
				bar.Title = page.Name
				bar.SetActions(page.Actions, page.Overflow)
			}
			layout.Inset{
				Top:    e.Insets.Top,
				Bottom: e.Insets.Bottom,
				Left:   e.Insets.Left,
				Right:  e.Insets.Right,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				content := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Max.X /= 3
							return nav.Layout(gtx, &navAnim)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return pages[nav.CurrentNavDestination().(int)].layout(gtx)
							})
						}),
					)
				})
				bar := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return bar.Layout(gtx)
				})
				flex := layout.Flex{Axis: layout.Vertical}
				if bottomBar.Value {
					flex.Layout(gtx, content, bar)
				} else {
					flex.Layout(gtx, bar, content)
				}
				modal.Layout(gtx)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
			e.Frame(gtx.Ops)
		}
	}
}
