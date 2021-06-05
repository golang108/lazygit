package gui

import (
	"fmt"
	"sync"

	"github.com/jesseduffield/lazygit/pkg/commands"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	. "github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

// list panel functions

func (gui *Gui) getSelectedLocalCommit() *models.Commit {
	selectedLine := gui.State.Panels.Commits.SelectedLineIdx
	if selectedLine == -1 || selectedLine > len(gui.State.Commits)-1 {
		return nil
	}

	return gui.State.Commits[selectedLine]
}

func (gui *Gui) handleCommitSelect() error {
	state := gui.State.Panels.Commits
	if state.SelectedLineIdx > 290 && state.LimitCommits {
		state.LimitCommits = false
		go utils.Safe(func() {
			if err := gui.refreshCommitsWithLimit(); err != nil {
				_ = gui.SurfaceError(err)
			}
		})
	}

	gui.escapeLineByLinePanel()

	var task updateTask
	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		task = NewRenderStringTask(gui.Tr.NoCommitsThisBranch)
	} else {
		cmd := gui.OSCommand.ExecutableFromString(
			gui.GitCommand.ShowCmdStr(commit.Sha, gui.State.Modes.Filtering.GetPath()),
		)
		task = NewRunPtyTask(cmd)
	}

	return gui.refreshMainViews(refreshMainOpts{
		main: &viewUpdateOpts{
			title: "Patch",
			task:  task,
		},
		secondary: gui.secondaryPatchPanelUpdateOpts(),
	})
}

// during startup, the bottleneck is fetching the reflog entries. We need these
// on startup to sort the branches by recency. So we have two phases: INITIAL, and COMPLETE.
// In the initial phase we don't get any reflog commits, but we asynchronously get them
// and refresh the branches after that
func (gui *Gui) refreshReflogCommitsConsideringStartup() {
	switch gui.State.StartupStage {
	case INITIAL:
		go utils.Safe(func() {
			_ = gui.refreshReflogCommits()
			gui.refreshBranches()
			gui.State.StartupStage = COMPLETE
		})

	case COMPLETE:
		_ = gui.refreshReflogCommits()
	}
}

// whenever we change commits, we should update branches because the upstream/downstream
// counts can change. Whenever we change branches we should probably also change commits
// e.g. in the case of switching branches.
func (gui *Gui) refreshCommits() error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go utils.Safe(func() {
		gui.refreshReflogCommitsConsideringStartup()

		gui.refreshBranches()
		wg.Done()
	})

	go utils.Safe(func() {
		_ = gui.refreshCommitsWithLimit()
		context, ok := gui.State.Contexts.CommitFiles.GetParentContext()
		if ok && context.GetKey() == BRANCH_COMMITS_CONTEXT_KEY {
			// This makes sense when we've e.g. just amended a commit, meaning we get a new commit SHA at the same position.
			// However if we've just added a brand new commit, it pushes the list down by one and so we would end up
			// showing the contents of a different commit than the one we initially entered.
			// Ideally we would know when to refresh the commit files context and when not to,
			// or perhaps we could just pop that context off the stack whenever cycling windows.
			// For now the awkwardness remains.
			commit := gui.getSelectedLocalCommit()
			if commit != nil {
				gui.State.Panels.CommitFiles.refName = commit.RefName()
				_ = gui.refreshCommitFilesView()
			}
		}
		wg.Done()
	})

	wg.Wait()

	return nil
}

func (gui *Gui) refreshCommitsWithLimit() error {
	gui.Mutexes.BranchCommitsMutex.Lock()
	defer gui.Mutexes.BranchCommitsMutex.Unlock()

	builder := commands.NewCommitListBuilder(gui.Log, gui.GitCommand, gui.OSCommand, gui.Tr)

	commits, err := builder.GetCommits(
		commands.GetCommitsOptions{
			Limit:                gui.State.Panels.Commits.LimitCommits,
			FilterPath:           gui.State.Modes.Filtering.GetPath(),
			IncludeRebaseCommits: true,
			RefName:              "HEAD",
		},
	)
	if err != nil {
		return err
	}
	gui.State.Commits = commits

	return gui.postRefreshUpdate(gui.State.Contexts.BranchCommits)
}

func (gui *Gui) refreshRebaseCommits() error {
	gui.Mutexes.BranchCommitsMutex.Lock()
	defer gui.Mutexes.BranchCommitsMutex.Unlock()

	builder := commands.NewCommitListBuilder(gui.Log, gui.GitCommand, gui.OSCommand, gui.Tr)

	updatedCommits, err := builder.MergeRebasingCommits(gui.State.Commits)
	if err != nil {
		return err
	}
	gui.State.Commits = updatedCommits

	return gui.postRefreshUpdate(gui.State.Contexts.BranchCommits)
}

// specific functions

func (gui *Gui) handleCommitSquashDown() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	if len(gui.State.Commits) <= 1 {
		return gui.CreateErrorPanel(gui.Tr.YouNoCommitsToSquash)
	}

	applied, err := gui.handleMidRebaseCommand("squash")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	return gui.Ask(AskOpts{
		Title:  gui.Tr.Squash,
		Prompt: gui.Tr.SureSquashThisCommit,
		HandleConfirm: func() error {
			return gui.WithWaitingStatus(gui.Tr.SquashingStatus, func() error {
				err := gui.GitCommand.WithSpan(gui.Tr.Spans.SquashCommitDown).InteractiveRebase(gui.State.Commits, gui.State.Panels.Commits.SelectedLineIdx, "squash")
				return gui.handleGenericMergeCommandResult(err)
			})
		},
	})
}

func (gui *Gui) handleCommitFixup() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	if len(gui.State.Commits) <= 1 {
		return gui.CreateErrorPanel(gui.Tr.YouNoCommitsToSquash)
	}

	applied, err := gui.handleMidRebaseCommand("fixup")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	return gui.Ask(AskOpts{
		Title:  gui.Tr.Fixup,
		Prompt: gui.Tr.SureFixupThisCommit,
		HandleConfirm: func() error {
			return gui.WithWaitingStatus(gui.Tr.FixingStatus, func() error {
				err := gui.GitCommand.WithSpan(gui.Tr.Spans.FixupCommit).InteractiveRebase(gui.State.Commits, gui.State.Panels.Commits.SelectedLineIdx, "fixup")
				return gui.handleGenericMergeCommandResult(err)
			})
		},
	})
}

func (gui *Gui) handleRenameCommit() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	applied, err := gui.handleMidRebaseCommand("reword")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	if gui.State.Panels.Commits.SelectedLineIdx != 0 {
		return gui.CreateErrorPanel(gui.Tr.OnlyRenameTopCommit)
	}

	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	message, err := gui.GitCommand.GetCommitMessage(commit.Sha)
	if err != nil {
		return gui.SurfaceError(err)
	}

	return gui.Prompt(PromptOpts{
		Title:          gui.Tr.LcRenameCommit,
		InitialContent: message,
		HandleConfirm: func(response string) error {
			if err := gui.GitCommand.WithSpan(gui.Tr.Spans.RewordCommit).RenameCommit(response); err != nil {
				return gui.SurfaceError(err)
			}

			return gui.RefreshSidePanels(RefreshOptions{Mode: ASYNC})
		},
	})
}

func (gui *Gui) handleRenameCommitEditor() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	applied, err := gui.handleMidRebaseCommand("reword")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	subProcess, err := gui.GitCommand.WithSpan(gui.Tr.Spans.RewordCommit).RewordCommit(gui.State.Commits, gui.State.Panels.Commits.SelectedLineIdx)
	if err != nil {
		return gui.SurfaceError(err)
	}
	if subProcess != nil {
		return gui.runSubprocessWithSuspenseAndRefresh(subProcess)
	}

	return nil
}

// handleMidRebaseCommand sees if the selected commit is in fact a rebasing
// commit meaning you are trying to edit the todo file rather than actually
// begin a rebase. It then updates the todo file with that action
func (gui *Gui) handleMidRebaseCommand(action string) (bool, error) {
	selectedCommit := gui.State.Commits[gui.State.Panels.Commits.SelectedLineIdx]
	if selectedCommit.Status != "rebasing" {
		return false, nil
	}

	// for now we do not support setting 'reword' because it requires an editor
	// and that means we either unconditionally wait around for the subprocess to ask for
	// our input or we set a lazygit client as the EDITOR env variable and have it
	// request us to edit the commit message when prompted.
	if action == "reword" {
		return true, gui.CreateErrorPanel(gui.Tr.LcRewordNotSupported)
	}

	gui.OnRunCommand(oscommands.NewCmdLogEntry(
		fmt.Sprintf("Updating rebase action of commit %s to '%s'", selectedCommit.ShortSha(), action),
		"Update rebase TODO",
		false,
	))

	if err := gui.GitCommand.EditRebaseTodo(gui.State.Panels.Commits.SelectedLineIdx, action); err != nil {
		return false, gui.SurfaceError(err)
	}

	return true, gui.refreshRebaseCommits()
}

func (gui *Gui) handleCommitDelete() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	applied, err := gui.handleMidRebaseCommand("drop")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	return gui.Ask(AskOpts{
		Title:  gui.Tr.DeleteCommitTitle,
		Prompt: gui.Tr.DeleteCommitPrompt,
		HandleConfirm: func() error {
			return gui.WithWaitingStatus(gui.Tr.DeletingStatus, func() error {
				err := gui.GitCommand.WithSpan(gui.Tr.Spans.DropCommit).InteractiveRebase(gui.State.Commits, gui.State.Panels.Commits.SelectedLineIdx, "drop")
				return gui.handleGenericMergeCommandResult(err)
			})
		},
	})
}

func (gui *Gui) handleCommitMoveDown() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	span := gui.Tr.Spans.MoveCommitDown

	index := gui.State.Panels.Commits.SelectedLineIdx
	selectedCommit := gui.State.Commits[index]
	if selectedCommit.Status == "rebasing" {
		if gui.State.Commits[index+1].Status != "rebasing" {
			return nil
		}

		// logging directly here because MoveTodoDown doesn't have enough information
		// to provide a useful log
		gui.OnRunCommand(oscommands.NewCmdLogEntry(
			fmt.Sprintf("Moving commit %s down", selectedCommit.ShortSha()),
			span,
			false,
		))

		if err := gui.GitCommand.MoveTodoDown(index); err != nil {
			return gui.SurfaceError(err)
		}
		gui.State.Panels.Commits.SelectedLineIdx++
		return gui.refreshRebaseCommits()
	}

	return gui.WithWaitingStatus(gui.Tr.MovingStatus, func() error {
		err := gui.GitCommand.WithSpan(span).MoveCommitDown(gui.State.Commits, index)
		if err == nil {
			gui.State.Panels.Commits.SelectedLineIdx++
		}
		return gui.handleGenericMergeCommandResult(err)
	})
}

func (gui *Gui) handleCommitMoveUp() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	index := gui.State.Panels.Commits.SelectedLineIdx
	if index == 0 {
		return nil
	}

	span := gui.Tr.Spans.MoveCommitUp

	selectedCommit := gui.State.Commits[index]
	if selectedCommit.Status == "rebasing" {
		// logging directly here because MoveTodoDown doesn't have enough information
		// to provide a useful log
		gui.OnRunCommand(oscommands.NewCmdLogEntry(
			fmt.Sprintf("Moving commit %s up", selectedCommit.ShortSha()),
			span,
			false,
		))

		if err := gui.GitCommand.MoveTodoDown(index - 1); err != nil {
			return gui.SurfaceError(err)
		}
		gui.State.Panels.Commits.SelectedLineIdx--
		return gui.refreshRebaseCommits()
	}

	return gui.WithWaitingStatus(gui.Tr.MovingStatus, func() error {
		err := gui.GitCommand.WithSpan(span).MoveCommitDown(gui.State.Commits, index-1)
		if err == nil {
			gui.State.Panels.Commits.SelectedLineIdx--
		}
		return gui.handleGenericMergeCommandResult(err)
	})
}

func (gui *Gui) handleCommitEdit() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	applied, err := gui.handleMidRebaseCommand("edit")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	return gui.WithWaitingStatus(gui.Tr.RebasingStatus, func() error {
		err = gui.GitCommand.WithSpan(gui.Tr.Spans.EditCommit).InteractiveRebase(gui.State.Commits, gui.State.Panels.Commits.SelectedLineIdx, "edit")
		return gui.handleGenericMergeCommandResult(err)
	})
}

func (gui *Gui) handleCommitAmendTo() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	return gui.Ask(AskOpts{
		Title:  gui.Tr.AmendCommitTitle,
		Prompt: gui.Tr.AmendCommitPrompt,
		HandleConfirm: func() error {
			return gui.WithWaitingStatus(gui.Tr.AmendingStatus, func() error {
				err := gui.GitCommand.WithSpan(gui.Tr.Spans.AmendCommit).AmendTo(gui.State.Commits[gui.State.Panels.Commits.SelectedLineIdx].Sha)
				return gui.handleGenericMergeCommandResult(err)
			})
		},
	})
}

func (gui *Gui) handleCommitPick() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	applied, err := gui.handleMidRebaseCommand("pick")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	// at this point we aren't actually rebasing so we will interpret this as an
	// attempt to pull. We might revoke this later after enabling configurable keybindings
	return gui.handlePullFiles()
}

func (gui *Gui) handleCommitRevert() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	commit := gui.getSelectedLocalCommit()

	if commit.IsMerge() {
		return gui.createRevertMergeCommitMenu(commit)
	} else {
		if err := gui.GitCommand.WithSpan(gui.Tr.Spans.RevertCommit).Revert(commit.Sha); err != nil {
			return gui.SurfaceError(err)
		}
		return gui.afterRevertCommit()
	}
}

func (gui *Gui) createRevertMergeCommitMenu(commit *models.Commit) error {
	menuItems := make([]*menuItem, len(commit.Parents))
	for i, parentSha := range commit.Parents {
		i := i
		message, err := gui.GitCommand.GetCommitMessageFirstLine(parentSha)
		if err != nil {
			return gui.SurfaceError(err)
		}

		menuItems[i] = &menuItem{
			displayString: fmt.Sprintf("%s: %s", utils.SafeTruncate(parentSha, 8), message),
			onPress: func() error {
				parentNumber := i + 1
				if err := gui.GitCommand.WithSpan(gui.Tr.Spans.RevertCommit).RevertMerge(commit.Sha, parentNumber); err != nil {
					return gui.SurfaceError(err)
				}
				return gui.afterRevertCommit()
			},
		}
	}

	return gui.createMenu(gui.Tr.SelectParentCommitForMerge, menuItems, createMenuOptions{showCancel: true})
}

func (gui *Gui) afterRevertCommit() error {
	gui.State.Panels.Commits.SelectedLineIdx++
	return gui.RefreshSidePanels(RefreshOptions{Mode: BLOCK_UI, Scope: []RefreshableView{COMMITS, BRANCHES}})
}

func (gui *Gui) handleViewCommitFiles() error {
	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	return gui.switchToCommitFilesContext(commit.Sha, true, gui.State.Contexts.BranchCommits, "commits")
}

func (gui *Gui) handleCreateFixupCommit() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	prompt := utils.ResolvePlaceholderString(
		gui.Tr.SureCreateFixupCommit,
		map[string]string{
			"commit": commit.Sha,
		},
	)

	return gui.Ask(AskOpts{
		Title:  gui.Tr.CreateFixupCommit,
		Prompt: prompt,
		HandleConfirm: func() error {
			if err := gui.GitCommand.WithSpan(gui.Tr.Spans.CreateFixupCommit).CreateFixupCommit(commit.Sha); err != nil {
				return gui.SurfaceError(err)
			}

			return gui.RefreshSidePanels(RefreshOptions{Mode: ASYNC})
		},
	})
}

func (gui *Gui) handleSquashAllAboveFixupCommits() error {
	if ok, err := gui.validateNotInFilterMode(); err != nil || !ok {
		return err
	}

	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	prompt := utils.ResolvePlaceholderString(
		gui.Tr.SureSquashAboveCommits,
		map[string]string{
			"commit": commit.Sha,
		},
	)

	return gui.Ask(AskOpts{
		Title:  gui.Tr.SquashAboveCommits,
		Prompt: prompt,
		HandleConfirm: func() error {
			return gui.WithWaitingStatus(gui.Tr.SquashingStatus, func() error {
				err := gui.GitCommand.WithSpan(gui.Tr.Spans.SquashAllAboveFixupCommits).SquashAllAboveFixupCommits(commit.Sha)
				return gui.handleGenericMergeCommandResult(err)
			})
		},
	})
}

func (gui *Gui) handleTagCommit() error {
	// TODO: bring up menu asking if you want to make a lightweight or annotated tag
	// if annotated, switch to a subprocess to create the message

	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	return gui.handleCreateLightweightTag(commit.Sha)
}

func (gui *Gui) handleCreateLightweightTag(commitSha string) error {
	return gui.Prompt(PromptOpts{
		Title: gui.Tr.TagNameTitle,
		HandleConfirm: func(response string) error {
			if err := gui.GitCommand.WithSpan(gui.Tr.Spans.CreateLightweightTag).CreateLightweightTag(response, commitSha); err != nil {
				return gui.SurfaceError(err)
			}
			return gui.RefreshSidePanels(RefreshOptions{Mode: ASYNC, Scope: []RefreshableView{COMMITS, TAGS}})
		},
	})
}

func (gui *Gui) handleCheckoutCommit() error {
	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	return gui.Ask(AskOpts{
		Title:  gui.Tr.LcCheckoutCommit,
		Prompt: gui.Tr.SureCheckoutThisCommit,
		HandleConfirm: func() error {
			return gui.handleCheckoutRef(commit.Sha, handleCheckoutRefOptions{span: gui.Tr.Spans.CheckoutCommit})
		},
	})
}

func (gui *Gui) handleCreateCommitResetMenu() error {
	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return gui.CreateErrorPanel(gui.Tr.NoCommitsThisBranch)
	}

	return gui.createResetMenu(commit.Sha)
}

func (gui *Gui) handleOpenSearchForCommitsPanel(_viewName string) error {
	// we usually lazyload these commits but now that we're searching we need to load them now
	if gui.State.Panels.Commits.LimitCommits {
		gui.State.Panels.Commits.LimitCommits = false
		if err := gui.RefreshSidePanels(RefreshOptions{Mode: ASYNC, Scope: []RefreshableView{COMMITS}}); err != nil {
			return err
		}
	}

	return gui.handleOpenSearch("commits")
}

func (gui *Gui) handleGotoBottomForCommitsPanel() error {
	// we usually lazyload these commits but now that we're searching we need to load them now
	if gui.State.Panels.Commits.LimitCommits {
		gui.State.Panels.Commits.LimitCommits = false
		if err := gui.RefreshSidePanels(RefreshOptions{Mode: SYNC, Scope: []RefreshableView{COMMITS}}); err != nil {
			return err
		}
	}

	for _, context := range gui.getListContexts() {
		if context.ViewName == "commits" {
			return context.handleGotoBottom()
		}
	}

	return nil
}

func (gui *Gui) handleCopySelectedCommitMessageToClipboard() error {
	commit := gui.getSelectedLocalCommit()
	if commit == nil {
		return nil
	}

	message, err := gui.GitCommand.GetCommitMessage(commit.Sha)
	if err != nil {
		return gui.SurfaceError(err)
	}

	if err := gui.OSCommand.WithSpan(gui.Tr.Spans.CopyCommitMessageToClipboard).CopyToClipboard(message); err != nil {
		return gui.SurfaceError(err)
	}

	gui.raiseToast(gui.Tr.CommitMessageCopiedToClipboard)

	return nil
}
