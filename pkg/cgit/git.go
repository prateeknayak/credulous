package cgit

import (
	"time"

	"github.com/realestate-com-au/credulous/pkg/core"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type GitImpl struct{}

func NewGitImpl() *GitImpl {
	return &GitImpl{}
}

// IsGitRepo checks if the path is a git repo or not
func (g *GitImpl) IsGitRepo(checkpath string) (bool, error) {

	_, err := git.PlainOpen(checkpath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g *GitImpl) GitAddCommitFile(repopath, filename, message string, config core.RepoConfig) (commitId string, err error) {

	repo, err := git.PlainOpen(repopath)
	if err != nil {
		return "", err
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	_, err = w.Add(filename)
	if err != nil {
		return "", err

	}
	_, err = w.Status()
	if err != nil {
		return "", err

	}
	author := &object.Signature{When: time.Now()}
	if config.Name != "" {
		author.Name = config.Name
	}
	if config.Email != "" {
		author.Email = config.Email
	}

	commit, err := w.Commit(message, &git.CommitOptions{
		Author: author,
	})

	if err != nil {
		return "", err

	}
	return commit.String(), nil
}
