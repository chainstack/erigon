package datadir

import (
	"path/filepath"
)

// Dirs is the file system folder the node should use for any data storage
// requirements. The configured data directory will not be directly shared with
// registered services, instead those can use utility methods to create/access
// databases or flat files
type Dirs struct {
	DataDir   string
	Chaindata string
	Tmp       string
	Snap      string
	TxPool    string
	Nodes     string
}

func New(datadir string) Dirs {
	if datadir != "" {
		absdatadir, err := filepath.Abs(datadir)
		if err != nil {
			panic(err)
		}
		datadir = absdatadir
	}

	return Dirs{
		DataDir:   datadir,
		Chaindata: filepath.Join(datadir, "chaindata"),
		Tmp:       filepath.Join(datadir, "etl-temp"),
		Snap:      filepath.Join(datadir, "snapshots"),
		TxPool:    filepath.Join(datadir, "txpool"),
		Nodes:     filepath.Join(datadir, "nodes"),
	}
}