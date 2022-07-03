package context

import (
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

const (
	GLOBAL_CONTEXT_KEY              types.ContextKey = "global"
	STATUS_CONTEXT_KEY              types.ContextKey = "status"
	FILES_CONTEXT_KEY               types.ContextKey = "files"
	LOCAL_BRANCHES_CONTEXT_KEY      types.ContextKey = "localBranches"
	REMOTES_CONTEXT_KEY             types.ContextKey = "remotes"
	REMOTE_BRANCHES_CONTEXT_KEY     types.ContextKey = "remoteBranches"
	TAGS_CONTEXT_KEY                types.ContextKey = "tags"
	LOCAL_COMMITS_CONTEXT_KEY       types.ContextKey = "commits"
	REFLOG_COMMITS_CONTEXT_KEY      types.ContextKey = "reflogCommits"
	SUB_COMMITS_CONTEXT_KEY         types.ContextKey = "subCommits"
	COMMIT_FILES_CONTEXT_KEY        types.ContextKey = "commitFiles"
	STASH_CONTEXT_KEY               types.ContextKey = "stash"
	MAIN_NORMAL_CONTEXT_KEY         types.ContextKey = "normal"
	MAIN_MERGING_CONTEXT_KEY        types.ContextKey = "merging"
	MAIN_PATCH_BUILDING_CONTEXT_KEY types.ContextKey = "patchBuilding"
	MAIN_STAGING_CONTEXT_KEY        types.ContextKey = "staging"
	MENU_CONTEXT_KEY                types.ContextKey = "menu"
	CONFIRMATION_CONTEXT_KEY        types.ContextKey = "confirmation"
	SEARCH_CONTEXT_KEY              types.ContextKey = "search"
	COMMIT_MESSAGE_CONTEXT_KEY      types.ContextKey = "commitMessage"
	SUBMODULES_CONTEXT_KEY          types.ContextKey = "submodules"
	SUGGESTIONS_CONTEXT_KEY         types.ContextKey = "suggestions"
	COMMAND_LOG_CONTEXT_KEY         types.ContextKey = "cmdLog"
)

var AllContextKeys = []types.ContextKey{
	GLOBAL_CONTEXT_KEY, // not focusable
	STATUS_CONTEXT_KEY,
	FILES_CONTEXT_KEY,
	LOCAL_BRANCHES_CONTEXT_KEY,
	REMOTES_CONTEXT_KEY,
	REMOTE_BRANCHES_CONTEXT_KEY,
	TAGS_CONTEXT_KEY,
	LOCAL_COMMITS_CONTEXT_KEY,
	REFLOG_COMMITS_CONTEXT_KEY,
	SUB_COMMITS_CONTEXT_KEY,
	COMMIT_FILES_CONTEXT_KEY,
	STASH_CONTEXT_KEY,
	MAIN_NORMAL_CONTEXT_KEY, // not focusable
	MAIN_MERGING_CONTEXT_KEY,
	MAIN_PATCH_BUILDING_CONTEXT_KEY,
	MAIN_STAGING_CONTEXT_KEY, // not focusable for secondary view
	MENU_CONTEXT_KEY,
	CONFIRMATION_CONTEXT_KEY,
	SEARCH_CONTEXT_KEY,
	COMMIT_MESSAGE_CONTEXT_KEY,
	SUBMODULES_CONTEXT_KEY,
	SUGGESTIONS_CONTEXT_KEY,
	COMMAND_LOG_CONTEXT_KEY,
}

type ContextTree struct {
	Global         types.Context
	Status         types.Context
	Files          *WorkingTreeContext
	Menu           *MenuContext
	Branches       *BranchesContext
	Tags           *TagsContext
	LocalCommits   *LocalCommitsContext
	CommitFiles    *CommitFilesContext
	Remotes        *RemotesContext
	Submodules     *SubmodulesContext
	RemoteBranches *RemoteBranchesContext
	ReflogCommits  *ReflogCommitsContext
	SubCommits     *SubCommitsContext
	Stash          *StashContext
	Suggestions    *SuggestionsContext
	Normal         types.Context
	Staging        types.Context
	PatchBuilding  types.Context
	Merging        types.Context
	Confirmation   types.Context
	CommitMessage  types.Context
	Search         types.Context
	CommandLog     types.Context
}

// the order of this decides which context is initially at the top of its window
func (self *ContextTree) Flatten() []types.Context {
	// TODO: add context for staging secondary, etc
	return []types.Context{
		self.Global,
		self.Status,
		self.Submodules,
		self.Files,
		self.SubCommits,
		self.Remotes,
		self.RemoteBranches,
		self.Tags,
		self.Branches,
		self.CommitFiles,
		self.ReflogCommits,
		self.LocalCommits,
		self.Stash,
		self.Menu,
		self.Confirmation,
		self.CommitMessage,
		self.Staging,
		self.Merging,
		self.PatchBuilding,
		self.Normal,
		self.Suggestions,
		self.CommandLog,
	}
}

type TabView struct {
	Tab      string
	ViewName string
}
