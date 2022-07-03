package gui

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/constants"
	"github.com/jesseduffield/lazygit/pkg/gui/controllers/helpers"
	"github.com/jesseduffield/lazygit/pkg/gui/keybindings"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) getKeyDisplay(name string) string {
	key := gui.getKey(name)
	return keybindings.GetKeyDisplay(key)
}

func (gui *Gui) getKey(key string) types.Key {
	runeCount := utf8.RuneCountInString(key)
	if runeCount > 1 {
		binding := keybindings.Keymap[strings.ToLower(key)]
		if binding == nil {
			log.Fatalf("Unrecognized key %s for keybinding. For permitted values see %s", strings.ToLower(key), constants.Links.Docs.CustomKeybindings)
		} else {
			return binding
		}
	} else if runeCount == 1 {
		return []rune(key)[0]
	}
	log.Fatal("Key empty for keybinding: " + strings.ToLower(key))
	return nil
}

func (gui *Gui) noPopupPanel(f func() error) func() error {
	return func() error {
		if gui.popupPanelFocused() {
			return nil
		}

		return f()
	}
}

// only to be called from the cheatsheet generate script. This mutates the Gui struct.
func (self *Gui) GetCheatsheetKeybindings() []*types.Binding {
	self.helpers = helpers.NewStubHelpers()
	self.State = &GuiRepoState{}
	self.State.Contexts = self.contextTree()
	self.resetControllers()
	bindings, _ := self.GetInitialKeybindings()
	return bindings
}

// renaming receiver to 'self' to aid refactoring. Will probably end up moving all Gui handlers to this pattern eventually.
func (self *Gui) GetInitialKeybindings() ([]*types.Binding, []*gocui.ViewMouseBinding) {
	config := self.c.UserConfig.Keybinding

	guards := types.KeybindingGuards{
		OutsideFilterMode: self.outsideFilterMode,
		NoPopupPanel:      self.noPopupPanel,
	}

	opts := types.KeybindingsOpts{
		GetKey: self.getKey,
		Config: config,
		Guards: guards,
	}

	bindings := []*types.Binding{
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.Quit),
			Modifier: gocui.ModNone,
			Handler:  self.handleQuit,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.QuitWithoutChangingDirectory),
			Modifier: gocui.ModNone,
			Handler:  self.handleQuitWithoutChangingDirectory,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.QuitAlt1),
			Modifier: gocui.ModNone,
			Handler:  self.handleQuit,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.Return),
			Modifier: gocui.ModNone,
			Handler:  self.handleTopLevelReturn,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.OpenRecentRepos),
			Handler:     self.handleCreateRecentReposMenu,
			Description: self.c.Tr.SwitchRepo,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.ScrollUpMain),
			Handler:     self.scrollUpMain,
			Alternative: "fn+up/shift+k",
			Description: self.c.Tr.LcScrollUpMainPanel,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.ScrollDownMain),
			Handler:     self.scrollDownMain,
			Alternative: "fn+down/shift+j",
			Description: self.c.Tr.LcScrollDownMainPanel,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.ScrollUpMainAlt1),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.ScrollDownMainAlt1),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownMain,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.ScrollUpMainAlt2),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.ScrollDownMainAlt2),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownMain,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.CreateRebaseOptionsMenu),
			Handler:     self.helpers.MergeAndRebase.CreateRebaseOptionsMenu,
			Description: self.c.Tr.ViewMergeRebaseOptions,
			OpensMenu:   true,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.CreatePatchOptionsMenu),
			Handler:     self.handleCreatePatchOptionsMenu,
			Description: self.c.Tr.ViewPatchOptions,
			OpensMenu:   true,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.Refresh),
			Handler:     self.handleRefresh,
			Description: self.c.Tr.LcRefresh,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.OptionMenu),
			Handler:     self.handleCreateOptionsMenu,
			Description: self.c.Tr.LcOpenMenu,
			OpensMenu:   true,
		},
		{
			ViewName: "",
			Key:      opts.GetKey(opts.Config.Universal.OptionMenuAlt1),
			Modifier: gocui.ModNone,
			Handler:  self.handleCreateOptionsMenu,
		},
		{
			ViewName:    "status",
			Key:         opts.GetKey(opts.Config.Universal.Edit),
			Handler:     self.handleEditConfig,
			Description: self.c.Tr.EditConfig,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.NextScreenMode),
			Handler:     self.nextScreenMode,
			Description: self.c.Tr.LcNextScreenMode,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.PrevScreenMode),
			Handler:     self.prevScreenMode,
			Description: self.c.Tr.LcPrevScreenMode,
		},
		{
			ViewName:    "status",
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.handleOpenConfig,
			Description: self.c.Tr.OpenConfig,
		},
		{
			ViewName:    "status",
			Key:         opts.GetKey(opts.Config.Status.CheckForUpdate),
			Handler:     self.handleCheckForUpdate,
			Description: self.c.Tr.LcCheckForUpdate,
		},
		{
			ViewName:    "status",
			Key:         opts.GetKey(opts.Config.Status.RecentRepos),
			Handler:     self.handleCreateRecentReposMenu,
			Description: self.c.Tr.SwitchRepo,
		},
		{
			ViewName:    "status",
			Key:         opts.GetKey(opts.Config.Status.AllBranchesLogGraph),
			Handler:     self.handleShowAllBranchLogs,
			Description: self.c.Tr.LcAllBranchesLogGraph,
		},
		{
			ViewName:    "files",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyFileNameToClipboard,
		},
		{
			ViewName:    "localBranches",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyBranchNameToClipboard,
		},
		{
			ViewName:    "commits",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyCommitShaToClipboard,
		},
		{
			ViewName:    "commits",
			Key:         opts.GetKey(opts.Config.Commits.ResetCherryPick),
			Handler:     self.helpers.CherryPick.Reset,
			Description: self.c.Tr.LcResetCherryPick,
		},
		{
			ViewName:    "reflogCommits",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyCommitShaToClipboard,
		},
		{
			ViewName:    "subCommits",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyCommitShaToClipboard,
		},
		{
			ViewName: "information",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleInfoClick,
		},
		{
			ViewName:    "commitFiles",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopyCommitFileNameToClipboard,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.FilteringMenu),
			Handler:     self.handleCreateFilteringMenuPanel,
			Description: self.c.Tr.LcOpenFilteringMenu,
			OpensMenu:   true,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.DiffingMenu),
			Handler:     self.handleCreateDiffingMenuPanel,
			Description: self.c.Tr.LcOpenDiffingMenu,
			OpensMenu:   true,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.DiffingMenuAlt),
			Handler:     self.handleCreateDiffingMenuPanel,
			Description: self.c.Tr.LcOpenDiffingMenu,
			OpensMenu:   true,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.ExtrasMenu),
			Handler:     self.handleCreateExtrasMenuPanel,
			Description: self.c.Tr.LcOpenExtrasMenu,
			OpensMenu:   true,
		},
		{
			ViewName: "secondary",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpSecondary,
		},
		{
			ViewName: "secondary",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownSecondary,
		},
		{
			ViewName:    "main",
			Key:         gocui.MouseWheelDown,
			Handler:     self.scrollDownMain,
			Description: self.c.Tr.ScrollDown,
			Alternative: "fn+up",
		},
		{
			ViewName:    "main",
			Key:         gocui.MouseWheelUp,
			Handler:     self.scrollUpMain,
			Description: self.c.Tr.ScrollUp,
			Alternative: "fn+down",
		},
		{
			ViewName: "stagingSecondary",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleTogglePanelClick,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.Return),
			Handler:     self.handleStagingEscape,
			Description: self.c.Tr.ReturnToFilesPanel,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.Select),
			Handler:     self.handleToggleStagedSelection,
			Description: self.c.Tr.StageSelection,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.Remove),
			Handler:     self.handleResetSelection,
			Description: self.c.Tr.ResetSelection,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.TogglePanel),
			Handler:     self.handleTogglePanel,
			Description: self.c.Tr.TogglePanel,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.Return),
			Handler:     self.handleEscapePatchBuildingPanel,
			Description: self.c.Tr.ExitLineByLineMode,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.Edit),
			Handler:     self.handleLineByLineEdit,
			Description: self.c.Tr.LcEditFile,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.HandleOpenFile,
			Description: self.c.Tr.LcOpenFile,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.Select),
			Handler:     self.handleToggleSelectionForPatch,
			Description: self.c.Tr.ToggleSelectionForPatch,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Main.EditSelectHunk),
			Handler:     self.handleEditHunk,
			Description: self.c.Tr.EditHunk,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.handleOpenFileAtLine,
			Description: self.c.Tr.LcOpenFile,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.PrevItem),
			Handler:     self.handleSelectPrevLine,
			Description: self.c.Tr.PrevLine,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.NextItem),
			Handler:     self.handleSelectNextLine,
			Description: self.c.Tr.NextLine,
		},
		{
			ViewName: "staging",
			Key:      opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevLine,
		},
		{
			ViewName: "staging",
			Key:      opts.GetKey(opts.Config.Universal.NextItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextLine,
		},
		{
			ViewName: "staging",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName: "staging",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownMain,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler:     self.handleSelectPrevHunk,
			Description: self.c.Tr.PrevHunk,
		},
		{
			ViewName: "staging",
			Key:      opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevHunk,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.NextBlock),
			Handler:     self.handleSelectNextHunk,
			Description: self.c.Tr.NextHunk,
		},
		{
			ViewName: "staging",
			Key:      opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextHunk,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Modifier:    gocui.ModNone,
			Handler:     self.copySelectedToClipboard,
			Description: self.c.Tr.LcCopySelectedTexToClipboard,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.NextPage),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineNextPage,
			Description: self.c.Tr.LcNextPage,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.PrevPage),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLinePrevPage,
			Description: self.c.Tr.LcPrevPage,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.GotoTop),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineGotoTop,
			Description: self.c.Tr.LcGotoTop,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.GotoBottom),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineGotoBottom,
			Description: self.c.Tr.LcGotoBottom,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.StartSearch),
			Handler:     func() error { return self.handleOpenSearch("staging") },
			Description: self.c.Tr.LcStartSearch,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelect),
			Handler:     self.handleToggleSelectRange,
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelectAlt),
			Handler:     self.handleToggleSelectRange,
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Main.ToggleSelectHunk),
			Handler:     self.handleToggleSelectHunk,
			Description: self.c.Tr.ToggleSelectHunk,
		},
		{
			ViewName: "staging",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleLBLMouseDown,
		},
		{
			ViewName: "staging",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModMotion,
			Handler:  self.handleMouseDrag,
		},
		{
			ViewName: "staging",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.handleOpenFileAtLine,
			Description: self.c.Tr.LcOpenFile,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.PrevItem),
			Handler:     self.handleSelectPrevLine,
			Description: self.c.Tr.PrevLine,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.NextItem),
			Handler:     self.handleSelectNextLine,
			Description: self.c.Tr.NextLine,
		},
		{
			ViewName: "patchBuilding",
			Key:      opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevLine,
		},
		{
			ViewName: "patchBuilding",
			Key:      opts.GetKey(opts.Config.Universal.NextItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextLine,
		},
		{
			ViewName: "patchBuilding",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName: "patchBuilding",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownMain,
		},
		{
			ViewName: "stagingSecondary",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownSecondary,
		},
		{
			ViewName: "secondary",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownSecondary,
		},
		{
			ViewName: "patchBuildingSecondary",
			Key:      gocui.MouseWheelDown,
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownSecondary,
		},
		{
			ViewName: "stagingSecondary",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpSecondary,
		},
		{
			ViewName: "patchBuildingSecondary",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpSecondary,
		},
		{
			ViewName: "secondary",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpSecondary,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler:     self.handleSelectPrevHunk,
			Description: self.c.Tr.PrevHunk,
		},
		{
			ViewName: "patchBuilding",
			Key:      opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevHunk,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.NextBlock),
			Handler:     self.handleSelectNextHunk,
			Description: self.c.Tr.NextHunk,
		},
		{
			ViewName: "patchBuilding",
			Key:      opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextHunk,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Modifier:    gocui.ModNone,
			Handler:     self.copySelectedToClipboard,
			Description: self.c.Tr.LcCopySelectedTexToClipboard,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.NextPage),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineNextPage,
			Description: self.c.Tr.LcNextPage,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.PrevPage),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLinePrevPage,
			Description: self.c.Tr.LcPrevPage,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.GotoTop),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineGotoTop,
			Description: self.c.Tr.LcGotoTop,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.GotoBottom),
			Modifier:    gocui.ModNone,
			Handler:     self.handleLineByLineGotoBottom,
			Description: self.c.Tr.LcGotoBottom,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.StartSearch),
			Handler:     func() error { return self.handleOpenSearch("patchBuilding") },
			Description: self.c.Tr.LcStartSearch,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelect),
			Handler:     self.handleToggleSelectRange,
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Main.ToggleDragSelectAlt),
			Handler:     self.handleToggleSelectRange,
			Description: self.c.Tr.ToggleDragSelect,
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Main.ToggleSelectHunk),
			Handler:     self.handleToggleSelectHunk,
			Description: self.c.Tr.ToggleSelectHunk,
		},
		{
			ViewName: "patchBuilding",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleLBLMouseDown,
		},
		{
			ViewName: "patchBuilding",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModMotion,
			Handler:  self.handleMouseDrag,
		},
		{
			ViewName: "patchBuilding",
			Key:      gocui.MouseWheelUp,
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpMain,
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.ScrollLeft),
			Handler:     self.scrollLeftMain,
			Description: self.c.Tr.LcScrollLeft,
			Tag:         "navigation",
		},
		{
			ViewName:    "staging",
			Key:         opts.GetKey(opts.Config.Universal.ScrollRight),
			Handler:     self.scrollRightMain,
			Description: self.c.Tr.LcScrollRight,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.ScrollLeft),
			Handler:     self.scrollLeftMain,
			Description: self.c.Tr.LcScrollLeft,
			Tag:         "navigation",
		},
		{
			ViewName:    "patchBuilding",
			Key:         opts.GetKey(opts.Config.Universal.ScrollRight),
			Handler:     self.scrollRightMain,
			Description: self.c.Tr.LcScrollRight,
			Tag:         "navigation",
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.ScrollLeft),
			Handler:     self.scrollLeftMain,
			Description: self.c.Tr.LcScrollLeft,
			Tag:         "navigation",
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.ScrollRight),
			Handler:     self.scrollRightMain,
			Description: self.c.Tr.LcScrollRight,
			Tag:         "navigation",
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.Return),
			Handler:     self.handleEscapeMerge,
			Description: self.c.Tr.ReturnToFilesPanel,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Files.OpenMergeTool),
			Handler:     self.helpers.WorkingTree.OpenMergeTool,
			Description: self.c.Tr.LcOpenMergeTool,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.Select),
			Handler:     self.handlePickHunk,
			Description: self.c.Tr.PickHunk,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Main.PickBothHunks),
			Handler:     self.handlePickAllHunks,
			Description: self.c.Tr.PickAllHunks,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler:     self.handleSelectPrevConflict,
			Description: self.c.Tr.PrevConflict,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.NextBlock),
			Handler:     self.handleSelectNextConflict,
			Description: self.c.Tr.NextConflict,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.PrevItem),
			Handler:     self.handleSelectPrevConflictHunk,
			Description: self.c.Tr.SelectPrevHunk,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.NextItem),
			Handler:     self.handleSelectNextConflictHunk,
			Description: self.c.Tr.SelectNextHunk,
		},
		{
			ViewName: "merging",
			Key:      opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevConflict,
		},
		{
			ViewName: "merging",
			Key:      opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextConflict,
		},
		{
			ViewName: "merging",
			Key:      opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectPrevConflictHunk,
		},
		{
			ViewName: "merging",
			Key:      opts.GetKey(opts.Config.Universal.NextItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.handleSelectNextConflictHunk,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.Edit),
			Handler:     self.handleMergeConflictEditFileAtLine,
			Description: self.c.Tr.LcEditFile,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.OpenFile),
			Handler:     self.handleMergeConflictOpenFileAtLine,
			Description: self.c.Tr.LcOpenFile,
		},
		{
			ViewName:    "merging",
			Key:         opts.GetKey(opts.Config.Universal.Undo),
			Handler:     self.handleMergeConflictUndo,
			Description: self.c.Tr.LcUndo,
		},
		{
			ViewName: "status",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleStatusClick,
		},
		{
			ViewName: "search",
			Key:      opts.GetKey(opts.Config.Universal.Confirm),
			Modifier: gocui.ModNone,
			Handler:  self.handleSearch,
		},
		{
			ViewName: "search",
			Key:      opts.GetKey(opts.Config.Universal.Return),
			Modifier: gocui.ModNone,
			Handler:  self.handleSearchEscape,
		},
		{
			ViewName: "confirmation",
			Key:      opts.GetKey(opts.Config.Universal.PrevItem),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      opts.GetKey(opts.Config.Universal.NextItem),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpConfirmationPanel,
		},
		{
			ViewName: "confirmation",
			Key:      opts.GetKey(opts.Config.Universal.NextItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownConfirmationPanel,
		},
		{
			ViewName:    "submodules",
			Key:         opts.GetKey(opts.Config.Universal.CopyToClipboard),
			Handler:     self.handleCopySelectedSideContextItemToClipboard,
			Description: self.c.Tr.LcCopySubmoduleNameToClipboard,
		},
		{
			ViewName:    "files",
			Key:         opts.GetKey(opts.Config.Universal.ToggleWhitespaceInDiffView),
			Handler:     self.toggleWhitespaceInDiffView,
			Description: self.c.Tr.ToggleWhitespaceInDiffView,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.IncreaseContextInDiffView),
			Handler:     self.IncreaseContextInDiffView,
			Description: self.c.Tr.IncreaseContextInDiffView,
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.DecreaseContextInDiffView),
			Handler:     self.DecreaseContextInDiffView,
			Description: self.c.Tr.DecreaseContextInDiffView,
		},
		{
			ViewName: "extras",
			Key:      gocui.MouseWheelUp,
			Handler:  self.scrollUpExtra,
		},
		{
			ViewName: "extras",
			Key:      gocui.MouseWheelDown,
			Handler:  self.scrollDownExtra,
		},
		{
			ViewName:    "extras",
			Key:         opts.GetKey(opts.Config.Universal.ExtrasMenu),
			Handler:     self.handleCreateExtrasMenuPanel,
			Description: self.c.Tr.LcOpenExtrasMenu,
			OpensMenu:   true,
		},
		{
			ViewName: "extras",
			Tag:      "navigation",
			Key:      opts.GetKey(opts.Config.Universal.PrevItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpExtra,
		},
		{
			ViewName: "extras",
			Tag:      "navigation",
			Key:      opts.GetKey(opts.Config.Universal.PrevItem),
			Modifier: gocui.ModNone,
			Handler:  self.scrollUpExtra,
		},
		{
			ViewName: "extras",
			Tag:      "navigation",
			Key:      opts.GetKey(opts.Config.Universal.NextItem),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownExtra,
		},
		{
			ViewName: "extras",
			Tag:      "navigation",
			Key:      opts.GetKey(opts.Config.Universal.NextItemAlt),
			Modifier: gocui.ModNone,
			Handler:  self.scrollDownExtra,
		},
		{
			ViewName: "extras",
			Tag:      "navigation",
			Key:      gocui.MouseLeft,
			Modifier: gocui.ModNone,
			Handler:  self.handleFocusCommandLog,
		},
	}

	mouseKeybindings := []*gocui.ViewMouseBinding{}
	for _, c := range self.State.Contexts.Flatten() {
		viewName := c.GetViewName()
		for _, binding := range c.GetKeybindings(opts) {
			// TODO: move all mouse keybindings into the mouse keybindings approach below
			binding.ViewName = viewName
			bindings = append(bindings, binding)
		}

		mouseKeybindings = append(mouseKeybindings, c.GetMouseKeybindings(opts)...)
	}

	for _, viewName := range []string{"status", "remotes", "tags", "localBranches", "remoteBranches", "files", "submodules", "reflogCommits", "commits", "commitFiles", "subCommits", "stash"} {
		bindings = append(bindings, []*types.Binding{
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.PrevBlock), Modifier: gocui.ModNone, Handler: self.previousSideWindow},
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.NextBlock), Modifier: gocui.ModNone, Handler: self.nextSideWindow},
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.PrevBlockAlt), Modifier: gocui.ModNone, Handler: self.previousSideWindow},
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.NextBlockAlt), Modifier: gocui.ModNone, Handler: self.nextSideWindow},
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.PrevBlockAlt2), Modifier: gocui.ModNone, Handler: self.previousSideWindow},
			{ViewName: viewName, Key: opts.GetKey(opts.Config.Universal.NextBlockAlt2), Modifier: gocui.ModNone, Handler: self.nextSideWindow},
		}...)
	}

	// Appends keybindings to jump to a particular sideView using numbers
	windows := []string{"status", "files", "localBranches", "commits", "stash"}

	if len(config.Universal.JumpToBlock) != len(windows) {
		log.Fatal("Jump to block keybindings cannot be set. Exactly 5 keybindings must be supplied.")
	} else {
		for i, window := range windows {
			bindings = append(bindings, &types.Binding{
				ViewName: "",
				Key:      opts.GetKey(opts.Config.Universal.JumpToBlock[i]),
				Modifier: gocui.ModNone,
				Handler:  self.goToSideWindow(window),
			})
		}
	}

	bindings = append(bindings, []*types.Binding{
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.NextTab),
			Handler:     self.handleNextTab,
			Description: self.c.Tr.LcNextTab,
			Tag:         "navigation",
		},
		{
			ViewName:    "",
			Key:         opts.GetKey(opts.Config.Universal.PrevTab),
			Handler:     self.handlePrevTab,
			Description: self.c.Tr.LcPrevTab,
			Tag:         "navigation",
		},
	}...)

	return bindings, mouseKeybindings
}

func (gui *Gui) resetKeybindings() error {
	gui.g.DeleteAllKeybindings()

	bindings, mouseBindings := gui.GetInitialKeybindings()

	// prepending because we want to give our custom keybindings precedence over default keybindings
	customBindings, err := gui.CustomCommandsClient.GetCustomCommandKeybindings()
	if err != nil {
		log.Fatal(err)
	}
	bindings = append(customBindings, bindings...)

	for _, binding := range bindings {
		if err := gui.SetKeybinding(binding); err != nil {
			return err
		}
	}

	for _, binding := range mouseBindings {
		if err := gui.SetMouseKeybinding(binding); err != nil {
			return err
		}
	}

	for _, values := range gui.viewTabMap() {
		for _, value := range values {
			viewName := value.ViewName
			tabClickCallback := func(tabIndex int) error { return gui.onViewTabClick(gui.windowForView(viewName), tabIndex) }

			if err := gui.g.SetTabClickBinding(viewName, tabClickCallback); err != nil {
				return err
			}
		}
	}

	return nil
}

func (gui *Gui) wrappedHandler(f func() error) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return f()
	}
}

func (gui *Gui) SetKeybinding(binding *types.Binding) error {
	handler := binding.Handler
	// TODO: move all mouse-ey stuff into new mouse approach
	if gocui.IsMouseKey(binding.Key) {
		handler = func() error {
			// we ignore click events on views that aren't popup panels, when a popup panel is focused
			if gui.popupPanelFocused() && gui.currentViewName() != binding.ViewName {
				return nil
			}

			return binding.Handler()
		}
	}

	return gui.g.SetKeybinding(binding.ViewName, binding.Key, binding.Modifier, gui.wrappedHandler(handler))
}

// warning: mutates the binding
func (gui *Gui) SetMouseKeybinding(binding *gocui.ViewMouseBinding) error {
	baseHandler := binding.Handler
	newHandler := func(opts gocui.ViewMouseBindingOpts) error {
		// we ignore click events on views that aren't popup panels, when a popup panel is focused
		if gui.popupPanelFocused() && gui.currentViewName() != binding.ViewName {
			return nil
		}

		return baseHandler(opts)
	}
	binding.Handler = newHandler

	return gui.g.SetViewClickBinding(binding)
}
