package constants

// Git commands and arguments.
const (
	GitBin             = "git"
	GitClone           = "clone"
	GitPull            = "pull"
	GitRebase          = "rebase"
	GitBranchFlag      = "-b"
	GitDirFlag         = "-C"
	GitFFOnlyFlag      = "--ff-only"
	GitPullRebaseFlag  = "--rebase"
	GitRebaseAbortFlag = "--abort"
	GitConfigCmd       = "config"
	GitGetFlag         = "--get"
	GitRemoteOrigin    = "remote.origin.url"
	GitRevParse        = "rev-parse"
	GitAbbrevRef       = "--abbrev-ref"
	GitHEAD            = "HEAD"
	GitTag             = "tag"
	GitCheckout        = "checkout"
	GitPush            = "push"
	GitLsRemote        = "ls-remote"
	GitLsRemoteTags    = "--tags"
	GitOrigin          = "origin"
	GitOriginPrefix    = "origin/"
	GitCommitPrefix    = "commit:"
	GitTagAnnotateFlag = "-a"
	GitTagMessageFlag  = "-m"
	GitTagListFlag     = "--list"
	GitBranchListFlag  = "--list"
	GitCatFile         = "cat-file"
	GitCatFileTypeFlag = "-t"
	GitCommitType      = "commit"
	GitTagGlob         = "v*"
)

// Git arguments for latest-branch operations.
const (
	GitFetch              = "fetch"
	GitBranch             = "branch"
	GitLog                = "log"
	GitForEachRef         = "for-each-ref"
	GitArgAll             = "--all"
	GitArgPrune           = "--prune"
	GitArgRemote          = "-r"
	GitArgContains        = "--contains"
	GitArgInsideWorkTree  = "--is-inside-work-tree"
	GitLogTipFormat       = "--format=%cI|%H|%s"
	GitLogDelimiter       = "|"
	GitLogFieldCount      = 3
	GitPointsAtFmt        = "--points-at=%s"
	GitRefsRemotesFmt     = "refs/remotes/%s"
	GitFormatRefnameShort = "--format=%(refname:short)"
	GitForEachRefTagFmt   = "--format=%(refname:short)|%(creatordate:iso-strict)"
	GitRefsTagsPrefix     = "refs/tags/"
	HeadPointer           = " -> "
	ShaDisplayLength      = 7
)

// Clone instruction format.
const (
	CloneInstructionFmt = "git clone -b %s %s %s"
	HTTPSFromSSHFmt     = "https://%s/%s"
	SSHFromHTTPSFmt     = "git@%s:%s"
)
