// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package tuf

import (
	"errors"

	"github.com/gittuf/gittuf/internal/common/set"
	"github.com/secure-systems-lab/go-securesystemslib/signerverifier"
)

const (
	// RootRoleName defines the expected name for the gittuf root of trust.
	RootRoleName = "root"

	// TargetsRoleName defines the expected name for the top level gittuf policy file.
	TargetsRoleName = "targets"

	// GitHubAppRoleName defines the expected name for the GitHub app role in the root of trust metadata.
	GitHubAppRoleName = "github-app"

	AllowRuleName          = "gittuf-allow-rule"
	ExhaustiveVerifierName = "gittuf-exhaustive-verifier"

	GittufPrefix           = "gittuf-"
	GittufControllerPrefix = "gittuf-controller"

	GlobalRuleThresholdType        = "threshold"
	GlobalRuleBlockForcePushesType = "block-force-pushes"
	RemoveGlobalRuleType           = "remove"
)

var (
	ErrInvalidRootMetadata                             = errors.New("invalid root metadata")
	ErrUnknownRootMetadataVersion                      = errors.New("unknown schema version for root metadata")
	ErrUnknownTargetsMetadataVersion                   = errors.New("unknown schema version for rule file metadata")
	ErrPrimaryRuleFileInformationNotFoundInRoot        = errors.New("root metadata does not contain primary rule file information")
	ErrGitHubAppInformationNotFoundInRoot              = errors.New("the special GitHub app role is not defined, but GitHub app approvals is set to trusted")
	ErrDuplicatedRuleName                              = errors.New("two rules with same name found in policy")
	ErrInvalidPrincipalID                              = errors.New("principal ID is invalid")
	ErrInvalidPrincipalType                            = errors.New("invalid principal type (do you have the right gittuf version?)")
	ErrPrincipalNotFound                               = errors.New("principal not found")
	ErrPrincipalStillInUse                             = errors.New("principal is still in use")
	ErrRuleNotFound                                    = errors.New("cannot find rule entry")
	ErrMissingRules                                    = errors.New("some rules are missing")
	ErrCannotManipulateRulesWithGittufPrefix           = errors.New("cannot add or change rules whose names have the 'gittuf-' prefix")
	ErrCannotMeetThreshold                             = errors.New("insufficient keys to meet threshold")
	ErrUnknownGlobalRuleType                           = errors.New("unknown global rule type")
	ErrGlobalRuleBlockForcePushesOnlyAppliesToGitPaths = errors.New("all patterns for block force pushes global rule must be for Git references")
	ErrGlobalRuleNotFound                              = errors.New("global rule not found")
	ErrGlobalRuleAlreadyExists                         = errors.New("global rule already exists")
	ErrPropagationDirectiveNotFound                    = errors.New("specified propagation directive not found")
	ErrNotAControllerRepository                        = errors.New("current repository is not marked as a controller repository")
)

// Principal represents an entity that is granted trust by gittuf metadata. In
// the simplest case, a principal may be a single public key. On the other hand,
// a principal may represent a human (who may control multiple keys), a team
// (consisting of multiple humans) etc.
type Principal interface {
	ID() string
	Keys() []*signerverifier.SSLibKey
	CustomMetadata() map[string]string
}

// RootMetadata represents the root of trust metadata for gittuf.
type RootMetadata interface {
	// SetExpires sets the expiry time for the metadata.
	// TODO: Does expiry make sense for the gittuf context? This is currently
	// unenforced
	SetExpires(expiry string)

	// SchemaVersion returns the metadata schema version.
	SchemaVersion() string

	// GetRepositoryLocation returns the canonical location of the Git
	// repository.
	GetRepositoryLocation() string
	// SetRepositoryLocation sets the specified repository location in the
	// root metadata.
	SetRepositoryLocation(location string)

	// GetPrincipals returns all the principals in the root metadata.
	GetPrincipals() map[string]Principal

	// AddRootPrincipal adds the corresponding principal to the root metadata
	// file and marks it as trusted for subsequent root of trust metadata.
	AddRootPrincipal(principal Principal) error
	// DeleteRootPrincipal removes the corresponding principal from the set of
	// trusted principals for the root of trust.
	DeleteRootPrincipal(principalID string) error
	// UpdateRootThreshold sets the required number of signatures for root of
	// trust metadata.
	UpdateRootThreshold(threshold int) error
	// GetRootPrincipals returns the principals trusted for the root of trust
	// metadata.
	GetRootPrincipals() ([]Principal, error)
	// GetRootThreshold returns the threshold of principals that must sign the
	// root of trust metadata.
	GetRootThreshold() (int, error)

	// AddPrincipalRuleFilePrincipal adds the corresponding principal to the
	// root metadata file and marks it as trusted for the primary rule file.
	AddPrimaryRuleFilePrincipal(principal Principal) error
	// DeletePrimaryRuleFilePrincipal removes the corresponding principal from
	// the set of trusted principals for the primary rule file.
	DeletePrimaryRuleFilePrincipal(principalID string) error
	// UpdatePrimaryRuleFileThreshold sets the required number of signatures for
	// the primary rule file.
	UpdatePrimaryRuleFileThreshold(threshold int) error
	// GetPrimaryRuleFilePrincipals returns the principals trusted for the
	// primary rule file.
	GetPrimaryRuleFilePrincipals() ([]Principal, error)
	// GetPrimaryRuleFileThreshold returns the threshold of principals that must
	// sign the primary rule file.
	GetPrimaryRuleFileThreshold() (int, error)

	// AddGlobalRule adds the corresponding rule to the root metadata.
	AddGlobalRule(globalRule GlobalRule) error
	// GetGlobalRules returns the global rules declared in the root metadata.
	GetGlobalRules() []GlobalRule
	// DeleteGlobalRule removes the global rule from the root metadata.
	DeleteGlobalRule(ruleName string) error

	// AddGitHubAppPrincipal adds the corresponding principal to the root
	// metadata and is trusted for GitHub app attestations.
	AddGitHubAppPrincipal(appName string, principal Principal) error
	// DeleteGitHubAppPrincipal removes the GitHub app attestations role from
	// the root of trust metadata.
	DeleteGitHubAppPrincipal(appName string)
	// EnableGitHubAppApprovals indicates attestations from the GitHub app role
	// must be trusted.
	// TODO: this needs to be generalized across tools
	EnableGitHubAppApprovals()
	// DisableGitHubAppApprovals indicates attestations from the GitHub app role
	// must not be trusted thereafter.
	// TODO: this needs to be generalized across tools
	DisableGitHubAppApprovals()
	// IsGitHubAppApprovalTrusted indicates if the GitHub app is trusted.
	// TODO: this needs to be generalized across tools
	IsGitHubAppApprovalTrusted() bool
	// GetGitHubAppPrincipals returns the principals trusted for the GitHub app
	// attestations.
	// TODO: this needs to be generalized across tools
	GetGitHubAppPrincipals() ([]Principal, error)

	// AddPropagationDirective adds a propagation directive to the root
	// metadata.
	AddPropagationDirective(directive PropagationDirective) error
	// GetPropagationDirectives returns the propagation directives found in the
	// root metadata.
	GetPropagationDirectives() []PropagationDirective
	// DeletePropagationDirective removes a propagation directive from the root
	// metadata.
	DeletePropagationDirective(name string) error

	// IsController indicates if the repository serves as the controller for
	// a multi-repository gittuf network.
	IsController() bool
	// EnableController marks the current repository as a controller
	// repository.
	EnableController() error
	// DisableController marks the current repository as not-a-controller.
	DisableController() error
	// AddControllerRepository adds the specified repository as a controller
	// for the current repository.
	AddControllerRepository(name, location string, initialRootPrincipals []Principal) error
	// AddNetworkRepository adds the specified repository as part of the
	// network for which the current repository is a controller. The current
	// repository must be marked as a controller before this can be used.
	AddNetworkRepository(name, location string, initialRootPrincipals []Principal) error
	// GetControllerRepositories returns the repositories that serve as the
	// controllers for the networks the current repository is a part of.
	GetControllerRepositories() []OtherRepository
	// GetNetworkRepositories returns the repositories that are part of the
	// network for which the current repository is a controller.
	// IsController must return true for this to be set.
	GetNetworkRepositories() []OtherRepository
}

// TargetsMetadata represents gittuf's rule files. Its name is inspired by TUF.
type TargetsMetadata interface {
	// SetExpires sets the expiry time for the metadata.
	// TODO: Does expiry make sense for the gittuf context? This is currently
	// unenforced
	SetExpires(expiry string)

	// SchemaVersion returns the metadata schema version.
	SchemaVersion() string

	// GetPrincipals returns all the principals in the rule file.
	GetPrincipals() map[string]Principal

	// GetRules returns all the rules in the metadata.
	GetRules() []Rule

	// AddRule adds a rule to the metadata file.
	AddRule(ruleName, access string, authorizedPrincipalIDs, rulePatterns []string, threshold int) error
	// UpdateRule updates an existing rule identified by ruleName with the
	// provided parameters.
	UpdateRule(ruleName, access string, authorizedPrincipalIDs, rulePatterns []string, threshold int) error
	// ReorderRules accepts the new order of rules (identified by their
	// ruleNames).
	ReorderRules(newRuleNames []string) error
	// RemoveRule deletes the rule identified by the ruleName.
	RemoveRule(ruleName string) error

	// AddPrincipal adds a principal to the metadata.
	AddPrincipal(principal Principal) error

	// RemovePrincipal removes a principal from the metadata.
	RemovePrincipal(principalID string) error
}

// Rule represents a rule entry in a rule file (`TargetsMetadata`).
type Rule interface {
	// ID returns the identifier of the rule, typically a name.
	ID() string

	// Matches indicates if the rule applies to a specified path.
	Matches(path string) bool

	// GetProtectedNamespaces returns the set of namespaces protected by the
	// rule.
	GetProtectedNamespaces() []string

	// GetPrincipalIDs returns the identifiers of the principals that are listed
	// as trusted by the rule.
	GetPrincipalIDs() *set.Set[string]
	// GetThreshold returns the threshold of principals that must approve to
	// meet the rule.
	GetThreshold() int

	// IsLastTrustedInRuleFile indicates that subsequent rules in the rule file
	// are not to be trusted if the current rule matches the namespace under
	// verification (similar to TUF's terminating behavior). However, the
	// current rule's delegated rules as well as other rules already in the
	// queue are trusted.
	IsLastTrustedInRuleFile() bool
}

// GlobalRule represents a repository-wide constraint set by the owners in the
// root metadata.
type GlobalRule interface {
	// GetName returns the name of the global rule.
	GetName() string
}

// GlobalRuleThreshold indicates the number of required approvals for a change
// to the specified namespaces to be valid.
type GlobalRuleThreshold interface {
	GlobalRule

	// Matches indicates if the rule applies to a specified path.
	Matches(path string) bool

	// GetProtectedNamespaces returns the set of namespaces protected by the
	// rule.
	GetProtectedNamespaces() []string

	// GetThreshold returns the threshold of principals that must approve to
	// meet the rule.
	GetThreshold() int
}

// GlobalRuleBlockForcePushes prevents force pushes or rewriting of history for
// the specified namespaces.
type GlobalRuleBlockForcePushes interface {
	GlobalRule

	// Matches indicates if the rule applies to a specified path.
	Matches(path string) bool

	// GetProtectedNamespaces returns the set of namespaces protected by the
	// rule.
	GetProtectedNamespaces() []string
}

// PropagationDirective represents an instruction to a gittuf client to carry
// out the propagation workflow.
type PropagationDirective interface {
	// GetName returns the name of the directive.
	GetName() string

	// GetUpstreamRepository returns the clone-friendly location of the upstream
	// repository.
	GetUpstreamRepository() string

	// GetUpstreamReference returns the reference that must be propagated from
	// the upstream repository.
	GetUpstreamReference() string

	// GetDownstreamReference returns the reference that the upstream components
	// must be propagated into in the downstream repository (i.e., the
	// repository where this directive is set.)
	GetDownstreamReference() string

	// GetDownstreamPath() returns the path in the Git tree of the downstream
	// reference where the upstream repository's contents must be stored by the
	// propagation workflow.
	GetDownstreamPath() string
}

// MultiRepository is used to configure gittuf to act in multi-repository
// setups. If the repository is a "controller", i.e., it declares policies for
// one or more other repositories, the contents of the controller repository's
// policy must be propagated to each of the other repositories.
type MultiRepository interface {
	// IsController indicates if the current repository acts as a controller
	// for a network of gittuf-enabled repositories.
	IsController() bool

	// GetControllerRepositories returns the repositories configured as a
	// controller for the current repository. In other words, the current
	// repository is a part of the network overseen by each of the
	// configured controller repositories.
	GetControllerRepositories() []OtherRepository

	// GetNetworkRepositories returns the repositories configured as part of
	// the network overseen by the repository. This must return `nil` if
	// IsController is `false`.
	GetNetworkRepositories() []OtherRepository
}

// OtherRepository represents another gittuf-enabled repository in the root
// metadata.
type OtherRepository interface {
	// GetName returns the user-friendly name of the other repository. It
	// must be unique among all the listed OtherRepository entries.
	GetName() string

	// GetLocation returns the clone-friendly location of the other
	// repository.
	GetLocation() string

	// GetInitialRootPrincipals returns the set of principals trusted to
	// sign the other repository's initial gittuf root of trust metadata.
	GetInitialRootPrincipals() []Principal
}
