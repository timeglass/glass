package hooks

type Git struct{}

// creates a new hook installer for
// git version control
func NewGit() *Git {
	return &Git{}
}

//@todo implement
func (g *Git) Install() error {

	//parse templates

	//write files

	return nil
}

//@todo implement
func (g *Git) Uninstall() error {

	//remove files

	return nil
}
