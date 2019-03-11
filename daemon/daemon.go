package daemon

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gertjaap/vertcoin/util"
	"github.com/gobuffalo/packr/v2"
)

var vertcoinDaemon *exec.Cmd

type daemon struct {
	Version    string `json:"version"`
	Url        string `json:"url"`
	Hash       []byte
	HashString string `json:"sha256"`
	Os         string `json:"os"`
	Arch       string `json:"arch"`
}

func StopDaemon() error {
	// Windows doesn't support Interrupt
	if runtime.GOOS == "windows" {
		_ = vertcoinDaemon.Process.Signal(os.Kill)
		return nil
	}

	go func() {
		time.Sleep(15 * time.Second)
		_ = vertcoinDaemon.Process.Signal(os.Kill)
	}()
	vertcoinDaemon.Process.Signal(os.Interrupt)

	return vertcoinDaemon.Wait()
}

func StartDaemon() error {
	// Find the available daemons
	box := packr.New("daemon", "./data")

	var daemons []*daemon
	daemonsJson, err := box.Find("daemons.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(daemonsJson, &daemons)
	if err != nil {
		return err
	}

	for _, d := range daemons {
		d.Hash, err = hex.DecodeString(d.HashString)
		if err != nil {
			return err
		}
	}

	var foundDaemon *daemon
	for _, d := range daemons {
		if d.Arch == runtime.GOARCH && d.Os == runtime.GOOS {
			foundDaemon = d
			break
		}
	}

	if foundDaemon == nil {
		return fmt.Errorf("Couldn't find a Vertcoin Core binary for your platform: %s %s", runtime.GOOS, runtime.GOARCH)
	}

	foundDaemon.start()
	return nil
}

func (d *daemon) start() error {
	// Check if the archive is available and it has the right SHA sum. Download if not
	err := d.ensureAvailable()
	if err != nil {
		return err
	}

	// Always re-unpack the archive to ensure no one tampered with the file on disk.
	err = d.unpack()
	if err != nil {
		return err
	}

	// Always do a fresh unpack of the executable to ensure there's been no funny
	// business. EnsureAvailable already checked the SHA hash.
	err = d.launch()
	if err != nil {
		return err
	}

	fmt.Printf("Started Vertcoind\n")
	return nil
}

func (d *daemon) launch() error {
	unpackDir := path.Join(util.DataDirectory(), "nodes", fmt.Sprintf("unpacked-%s", hex.EncodeToString(d.Hash)))
	unpackPath := path.Join(unpackDir, "vertcoind")

	fmt.Printf("Launching %s...\n", unpackPath)
	vertcoinDaemon = exec.Command(unpackPath, "-printtoconsole")
	vertcoinDaemon.Stderr = os.Stderr
	vertcoinDaemon.Stdout = os.Stdout
	fmt.Printf("Starting...\n")
	return vertcoinDaemon.Start()
}

func (d *daemon) unpack() error {

	nodePath := path.Join(util.DataDirectory(), "nodes", hex.EncodeToString(d.Hash))
	unpackDir := path.Join(util.DataDirectory(), "nodes", fmt.Sprintf("unpacked-%s", hex.EncodeToString(d.Hash)))
	r, err := zip.OpenReader(nodePath)
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := os.Stat(unpackDir); !os.IsNotExist(err) {
		fmt.Printf("Removing unpack directory\n")
		err = os.RemoveAll(unpackDir)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(unpackDir); os.IsNotExist(err) {
		fmt.Printf("(Re)creating unpack directory\n")
		err = os.Mkdir(unpackDir, 0755)
		if err != nil {
			return err
		}
	}

	unpackPath := path.Join(unpackDir, "vertcoind")
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "vertcoind") {
			fmt.Printf("Unpacking vertcoind\n")

			outFile, err := os.OpenFile(unpackPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
			fmt.Printf("Unpacked succesfully.\n")

			return nil
		}
	}

	return fmt.Errorf("Could not find a binary named 'vertcoind' in the archive")
}

func (d *daemon) ensureAvailable() error {
	freshDownload := false
	_ = os.Mkdir(path.Join(util.DataDirectory(), "nodes"), 0700)
	nodePath := path.Join(util.DataDirectory(), "nodes", hex.EncodeToString(d.Hash))
	_, err := os.Stat(nodePath)
	if os.IsNotExist(err) {
		fmt.Printf("Deamon not found, downloading...\n")
		freshDownload = true
		d.download()
	} else if err != nil {
		return err
	} else {
		fmt.Printf("Daemon file already exists\n")
	}

	shaSum, err := util.ShaSum(nodePath)
	if err != nil {
		return err
	}
	if !bytes.Equal(shaSum, d.Hash) {
		fmt.Printf("Hash differs: [%x] vs [%x]", shaSum, d.Hash)
		if !freshDownload {
			err = os.Remove(nodePath)
			if err != nil {
				return err
			}
			return d.ensureAvailable()
		} else {
			return fmt.Errorf("Freshly downloaded node did not have correct SHA256 hash")
		}
	}

	fmt.Printf("Daemon file is available and correct\n")
	return nil
}

func (d *daemon) download() error {
	nodePath := path.Join(util.DataDirectory(), "nodes", hex.EncodeToString(d.Hash))

	resp, err := http.Get(d.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(nodePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	fmt.Printf("Daemon file downloaded\n")
	return err
}
