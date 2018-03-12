package cgit

import (
	"testing"
)

func TestGitAdd(t *testing.T) {
	//g := NewGitImpl()
	//repopath := path.Join("testrepo." + fmt.Sprintf("%d", os.Getpid()))
	//
	//// need to create a new repo first
	//repo, err := git.InitRepository(repopath, false)
	//handler.LogAndDieOnFatalError(err)
	//defer os.RemoveAll(path.Clean(path.Join(repo.Path(), "..")))
	//
	//// Need to add some basic config so that tests will pass
	//config, _ := repo.Config()
	//_ = config.SetString("user.name", "Test User")
	//_ = config.SetString("user.email", "test.user@nowhere")
	//
	//Convey("Testing gitAdd", t, func() {
	//	Convey("Test add to non-existent repository", func() {
	//		_, err := g.GitAddCommitFile("/no/such/repo", "testdata/newcreds.json", "message")
	//		So(err, ShouldNotEqual, nil)
	//	})
	//
	//	Convey("Test add non-existent file to a repo", func() {
	//		_, err := g.GitAddCommitFile(repo.Path(), "/no/such/file", "message")
	//		// fmt.Println("gitAdd returned " + fmt.Sprintf("%s", err))
	//		So(err, ShouldNotEqual, nil)
	//	})
	//
	//	Convey("Test add an initial file to the repo", func() {
	//		fp, _ := os.Create(path.Join(repopath, "testfile"))
	//		_, _ = fp.WriteString("A test string")
	//		_ = fp.Close()
	//		commitId, err := g.GitAddCommitFile(repo.Path(), "testfile", "first commit")
	//		So(err, ShouldEqual, nil)
	//		So(commitId, ShouldNotEqual, nil)
	//		So(commitId, ShouldNotBeBlank)
	//	})
	//
	//	Convey("Test add a second file to the repo", func() {
	//		fp, _ := os.Create(path.Join(repopath, "testfile"))
	//		_, _ = fp.WriteString("A second test string")
	//		_ = fp.Close()
	//		commitId, err := g.GitAddCommitFile(repo.Path(), "testfile", "second commit")
	//		So(err, ShouldEqual, nil)
	//		So(commitId, ShouldNotEqual, nil)
	//		So(commitId, ShouldNotBeBlank)
	//	})
	//
	//	Convey("Test checking whether a repo is a repo", func() {
	//		fullpath, _ := filepath.Abs(repopath)
	//		isrepo, err := g.IsGitRepo(fullpath)
	//		So(err, ShouldEqual, nil)
	//		So(isrepo, ShouldEqual, true)
	//	})
	//
	//	Convey("Test checking whether a plain dir is a repo", func() {
	//		isrepo, err := g.IsGitRepo("/tmp")
	//		So(err, ShouldEqual, nil)
	//		So(isrepo, ShouldEqual, false)
	//	})
	//
	//})
}
