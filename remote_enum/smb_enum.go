package remote_enum

import (
	"divine-dragon/transport"
	"divine-dragon/util"
	"fmt"
	iofs "io/fs"
	"net"
	"os"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

type SmbEnumModule struct {
	remoteHost string
	remotePort string
	username   string
	password   string
	domain     string
	hash       string

	conn                 net.Conn
	s                    *smb2.Session
	shares               []string
	interestingFilePaths map[string][]string
	currentShare         string
	logger               util.Logger
}

func NewSmbModuleNewLdapEnumModule(domainOpt string, remoteHostOpt string, remotePortOpt string,
	usernameOpt string, passwordOpt string, hashOpt string, verboseOpt bool, logFileNameOpt string) *SmbEnumModule {
	sem := SmbEnumModule{
		remoteHost: remoteHostOpt,
		remotePort: remotePortOpt,
		username:   usernameOpt,
		password:   passwordOpt,
		hash:       hashOpt,
	}
	if domainOpt != "" {
		sem.domain = domainOpt
	} else {
		sem.domain = "WORKGROUP"
	}
	sem.logger = util.SmbEnumLogger(verboseOpt, logFileNameOpt)
	return &sem
}

func (sem *SmbEnumModule) Run() {
	conn, s, err := transport.SmbConnect(sem.remoteHost, sem.remotePort, sem.domain, sem.username, sem.password, sem.hash)
	if err != nil {
		sem.logger.Log.Error(err)
		return
	} else {
		sem.logger.Log.Noticef("Connected to %s on port %s and opened a session as %s", sem.remoteHost, sem.remotePort, sem.username)
	}
	sem.conn = conn
	sem.s = s
	sem.enumerateSharenames()
	sem.enumerateEachShareContent()
	sem.downloadAllFoundFiles()
	defer transport.SmbClose(conn, s)
}

func (sem *SmbEnumModule) enumerateSharenames() {
	sem.logger.Log.Info("Starting to enumerate sharenames...")
	names, err := sem.s.ListSharenames()
	if err != nil {
		sem.logger.Log.Errorf("can't enumerate shares: %v", err)
	}
	sem.shares = names
	fmt.Println("\tSHARENAMES")
	fmt.Println()
	fmt.Println("\t----------")
	for _, name := range names {
		// sem.accessCheck(name)
		fmt.Printf("\t%s\n", name)
	}
	fmt.Println()
}

func (sem *SmbEnumModule) enumerateEachShareContent() {
	sem.interestingFilePaths = make(map[string][]string)
	for _, share := range sem.shares {
		if share != "IPC$" && share != "NETLOGON" {
			sem.interestingFilePaths[share] = []string{}
			sem.currentShare = share
			sem.enumerateOneShareContent()
		}
	}
	// fmt.Println(sem.interestingFilePaths)
}

func (sem *SmbEnumModule) enumerateOneShareContent() {
	fs, err := sem.s.Mount(sem.currentShare)
	if err != nil {
		sem.logger.Log.Errorf("can't mount a share named %s with error: %v", sem.currentShare, err)
		return
	}
	defer fs.Umount()
	sem.logger.Log.Infof("Enumerating a content of a share - %s:", sem.currentShare)
	// matches, err := iofs.Glob(fs.DirFS("."), "*")
	// if err != nil {
	// 	sem.logger.Log.Errorf("can't enumerate a content of a share with error: %v", err)
	// 	return
	// }
	// for _, match := range matches {
	// 	fmt.Println(match)
	// }
	err = iofs.WalkDir(fs.DirFS("."), ".", sem.walkDirFunc)
	fmt.Println()
	if err != nil {
		sem.logger.Log.Error(err)
		return
	}
	sem.logger.Log.Infof("Done.")
}

func (sem *SmbEnumModule) walkDirFunc(path string, d iofs.DirEntry, err error) error {
	if d != nil {
		fmt.Println(path)
		if !d.IsDir() {
			sem.interestingFilePaths[sem.currentShare] = append(sem.interestingFilePaths[sem.currentShare], path)
		}
	}
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (sem *SmbEnumModule) downloadAllFoundFiles() {
	sem.logger.Log.Info("Downloading all the files from shares...")
	currentDir, err := os.Getwd()
	if err != nil {
		sem.logger.Log.Errorf("can't determine the current working directory: %v", err)
		return
	}
	currentDir += "/SHARES"
	for share, paths := range sem.interestingFilePaths {
		if len(paths) > 0 {
			shareDir := currentDir + "/" + share
			err := os.MkdirAll(shareDir, os.ModePerm)
			if err != nil {
				sem.logger.Log.Errorf("can't create a directory %s: %v", shareDir, err)
				continue
			}
			sem.downloadFileFromShare(share, paths, shareDir)
		}
	}
}

func (sem *SmbEnumModule) downloadFileFromShare(sharename string, paths []string, shareDir string) {
	sem.logger.Log.Infof("Trying to download all files from share %s ...", sharename)
	fs, err := sem.s.Mount(sharename)
	if err != nil {
		sem.logger.Log.Errorf("can't mount a share %s to download: %v", sharename, err)
		return
	}
	defer fs.Umount()
	for _, path := range paths {
		sem.logger.Log.Infof("Dowloading a file %s ...", path)
		fileContent, err := fs.ReadFile(path)
		if err != nil {
			sem.logger.Log.Errorf("can't read a remote file %s: %v", path, err)
			continue
		}
		pathSlice := strings.Split(path, "/")
		filename := pathSlice[len(pathSlice)-1]
		pathWithoutFilename := strings.Replace(path, filename, "", 1)
		fullPath := shareDir + "/" + pathWithoutFilename
		err = os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			sem.logger.Log.Errorf("can't create a directory for a remote path %s: %v", path, err)
		}
		fullFilename := fullPath + "/" + filename
		err = os.WriteFile(fullFilename, fileContent, os.ModePerm)
		if err != nil {
			sem.logger.Log.Errorf("can't create a file %s in path %s", filename, fullPath)
			continue
		}
	}
}

// func (sem *SmbEnumModule) accessCheck(sharename string) string {
// 	fs, err := sem.s.Mount(sharename)
// 	if err != nil {
// 		return "PERMISSION DENIED"
// 	}
// 	matches, err := iofs.Glob(fs.DirFS("."), "*")
// 	for _, match := range matches {
// 		fmt.Println(match)
// 		fi, err := fs.Stat(match)
// 		if err != nil {
// 			fmt.Println(err)
// 			return ""
// 		}
// 		fm := fi.Mode()
// 		if fm.IsDir() {
// 			fmt.Println(fm.String())
// 		}
// 	}
// 	if err != nil {
// 		return "PERMISSION DENIED"
// 	} else {
// 		return "GOT ACCESS"
// 	}
// }
