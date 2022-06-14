package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) contextTree() *context.ContextTree {
	return &context.ContextTree{
		Global: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.GLOBAL_CONTEXT,
				ViewName:   "",
				WindowName: "",
				Key:        context.GLOBAL_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{
				OnRenderToMain: OnFocusWrapper(gui.statusRenderToMain),
			},
		),
		Status: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.SIDE_CONTEXT,
				ViewName:   "status",
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
				ViewName:   "main",
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
				ViewName:   "secondary",
				WindowName: "secondary",
				Key:        context.NORMAL_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		StagingSecondary: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				ViewName:   "stagingSecondary",
				WindowName: "secondary",
				Key:        context.STAGING_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		PatchBuildingSecondary: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				ViewName:   "patchBuildingsecondary",
				WindowName: "secondary",
				Key:        context.PATCH_BUILDING_SECONDARY_CONTEXT_KEY,
				Focusable:  false,
			}),
			context.ContextCallbackOpts{},
		),
		Staging: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				ViewName:   "staging",
				WindowName: "main",
				Key:        context.STAGING_MAIN_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{
				OnFocus: func(opts ...types.OnFocusOpts) error {
					forceSecondaryFocused := false
					selectedLineIdx := -1
					if len(opts) > 0 && opts[0].ClickedViewName != "" {
						if opts[0].ClickedViewName == "staging" || opts[0].ClickedViewName == "secondaryStaging" {
							selectedLineIdx = opts[0].ClickedViewLineIdx
						}
						if opts[0].ClickedViewName == "secondaryStaging" {
							forceSecondaryFocused = true
						}
					}
					return gui.onStagingFocus(forceSecondaryFocused, selectedLineIdx)
				},
			},
		),
		PatchBuilding: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:       types.MAIN_CONTEXT,
				ViewName:   "patchBuilding",
				WindowName: "main",
				Key:        context.PATCH_BUILDING_MAIN_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{
				OnFocus: func(opts ...types.OnFocusOpts) error {
					selectedLineIdx := -1
					if len(opts) > 0 && (opts[0].ClickedViewName == "patchBuilding" || opts[0].ClickedViewName == "secondaryPatchBuilding") {
						selectedLineIdx = opts[0].ClickedViewLineIdx
					}

					return gui.onPatchBuildingFocus(selectedLineIdx)
				},
			},
		),
		Merging: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:            types.MAIN_CONTEXT,
				ViewName:        "merging",
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
				ViewName:              "confirmation",
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
				ViewName:              "commitMessage",
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
				ViewName:   "search",
				WindowName: "search",
				Key:        context.SEARCH_CONTEXT_KEY,
				Focusable:  true,
			}),
			context.ContextCallbackOpts{},
		),
		CommandLog: context.NewSimpleContext(
			context.NewBaseContext(context.NewBaseContextOpts{
				Kind:            types.EXTRAS_CONTEXT,
				ViewName:        "extras",
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
		Options:      context.NewDisplayContext("", "options", "options"),
		AppStatus:    context.NewDisplayContext("", "appStatus", "appStatus"),
		SearchPrefix: context.NewDisplayContext("", "searchPrefix", "searchPrefix"),
		Information:  context.NewDisplayContext("", "information", "information"),
		Limit:        context.NewDisplayContext("", "limit", "limit"),
	}
}

// using this wrapper for when an onFocus function doesn't care about any potential
// props that could be passed
func OnFocusWrapper(f func() error) func(opts ...types.OnFocusOpts) error {
	return func(opts ...types.OnFocusOpts) error {
		return f()
	}
}
