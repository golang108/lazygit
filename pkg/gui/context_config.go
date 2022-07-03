package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) contextTree() *context.ContextTree {
	return &context.ContextTree{
		Global: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:                  types.GLOBAL_CONTEXT,
				View:                  nil,
				WindowName:            "",
				Key:                   context.GLOBAL_CONTEXT_KEY,
				Focusable:             false,
				HasUncontrolledBounds: true, // setting to true because the global context doesn't even have a view
			}),
			context.ContextCallbackOpts{
				OnRenderToMain: OnFocusWrapper(gui.statusRenderToMain),
			},
		),
		Status: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.SIDE_CONTEXT,
				View:       gui.Views.Status,
				WindowName: "status",
				Key:        context.STATUS_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{
				OnRenderToMain: OnFocusWrapper(gui.statusRenderToMain),
			},
		),
		Files:          gui.filesListContext(),
		Submodules:     gui.submodulesListContext(),
		Menu:           gui.menuListContext(),
		Remotes:        gui.remotesListContext(),
		RemoteBranches: gui.remoteBranchesListContext(),
		LocalCommits:   gui.branchCommitsListContext(),
		CommitFiles:    gui.commitFilesListContext(),
		ReflogCommits:  gui.reflogCommitsListContext(),
		SubCommits:     gui.subCommitsListContext(),
		Branches:       gui.branchesListContext(),
		Tags:           gui.tagsListContext(),
		Stash:          gui.stashListContext(),
		Suggestions:    gui.suggestionsListContext(),
		Normal: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.Main,
				WindowName: "main",
				Key:        context.NORMAL_MAIN_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{
				OnFocus: func(opts ...types.OnFocusOpts) error {
					return nil // TODO: should we do something here? We should allow for scrolling the panel
				},
			},
		),
		NormalSecondary: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.Secondary,
				WindowName: "secondary",
				Key:        context.NORMAL_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		Staging: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.Staging,
				WindowName: "main",
				Key:        context.STAGING_MAIN_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{
				OnFocus: func(opts ...types.OnFocusOpts) error {
					forceSecondaryFocused := false
					selectedLineIdx := -1
					if len(opts) > 0 && opts[0].ClickedWindowName != "" {
						if opts[0].ClickedWindowName == "main" || opts[0].ClickedWindowName == "secondary" {
							selectedLineIdx = opts[0].ClickedViewLineIdx
						}
						if opts[0].ClickedWindowName == "secondary" {
							forceSecondaryFocused = true
						}
					}
					return gui.onStagingFocus(forceSecondaryFocused, selectedLineIdx)
				},
			},
		),
		StagingSecondary: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.StagingSecondary,
				WindowName: "secondary",
				Key:        context.STAGING_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		PatchBuilding: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.PatchBuilding,
				WindowName: "main",
				Key:        context.PATCH_BUILDING_MAIN_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{
				OnFocus: func(opts ...types.OnFocusOpts) error {
					selectedLineIdx := -1
					if len(opts) > 0 && (opts[0].ClickedWindowName == "main") {
						selectedLineIdx = opts[0].ClickedViewLineIdx
					}

					return gui.onPatchBuildingFocus(selectedLineIdx)
				},
			},
		),
		PatchBuildingSecondary: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				View:       gui.Views.PatchBuildingSecondary,
				WindowName: "secondary",
				Key:        context.PATCH_BUILDING_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		Merging: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:            types.MAIN_CONTEXT,
				View:            gui.Views.Merging,
				WindowName:      "main",
				Key:             context.MERGING_MAIN_CONTEXT_KEY,
				OnGetOptionsMap: gui.getMergingOptions,
				Focusable:       true,
			}),
			context.ContextCallbackOpts{
				OnFocus: OnFocusWrapper(func() error { return gui.renderConflictsWithLock(true) }),
			},
		),
		Confirmation: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:                  types.TEMPORARY_POPUP,
				View:                  gui.Views.Confirmation,
				WindowName:            "confirmation",
				Key:                   context.CONFIRMATION_CONTEXT_KEY,
				Focusable:             true,
				HasUncontrolledBounds: true,
			}),
			context.ContextCallbackOpts{
				OnFocus: OnFocusWrapper(gui.handleAskFocused),
			},
		),
		CommitMessage: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:                  types.PERSISTENT_POPUP,
				View:                  gui.Views.CommitMessage,
				WindowName:            "commitMessage",
				Key:                   context.COMMIT_MESSAGE_CONTEXT_KEY,
				Focusable:             true,
				HasUncontrolledBounds: true,
			}),
			context.ContextCallbackOpts{
				OnFocus: OnFocusWrapper(gui.handleCommitMessageFocused),
			},
		),
		Search: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.PERSISTENT_POPUP,
				View:       gui.Views.Search,
				WindowName: "search",
				Key:        context.SEARCH_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{},
		),
		CommandLog: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:            types.EXTRAS_CONTEXT,
				View:            gui.Views.Extras,
				WindowName:      "extras",
				Key:             context.COMMAND_LOG_CONTEXT_KEY,
				OnGetOptionsMap: gui.getMergingOptions,
				Focusable:       true,
			}),
			context.ContextCallbackOpts{
				OnFocusLost: func() error {
					gui.Views.Extras.Autoscroll = true
					return nil
				},
			},
		),
		// TODO: consider adding keys. Maybe they're not needed?
		Options:      context.NewDisplayContext("", gui.Views.Options, "options"),
		AppStatus:    context.NewDisplayContext("", gui.Views.AppStatus, "appStatus"),
		SearchPrefix: context.NewDisplayContext("", gui.Views.SearchPrefix, "searchPrefix"),
		Information:  context.NewDisplayContext("", gui.Views.Information, "information"),
		Limit:        context.NewDisplayContext("", gui.Views.Limit, "limit"),
	}
}

// using this wrapper for when an onFocus function doesn't care about any potential
// props that could be passed
func OnFocusWrapper(f func() error) func(opts ...types.OnFocusOpts) error {
	return func(opts ...types.OnFocusOpts) error {
		return f()
	}
}
