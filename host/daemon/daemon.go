package daemon

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/gertjaap/vertcoin/host/coinparams"
	"github.com/gertjaap/vertcoin/logging"
	"github.com/gertjaap/vertcoin/util"
)

type Daemon struct {
	cmd          *exec.Cmd
	rpc          *rpcclient.Client
	lastGenerate time.Time
	Coin         coinparams.Coin
	Node         coinparams.CoinNode
	Network      coinparams.CoinNetwork
	RpcPort      int
	RpcUser      string
	RpcPassword  string
}

func NewDaemon(coin coinparams.Coin, coinNode coinparams.CoinNode, coinNetwork coinparams.CoinNetwork, rpcPort int) *Daemon {
	rpcUser := util.RandomAlphaNumeric(16)
	rpcPass := util.RandomAlphaNumeric(16)
	return &Daemon{Coin: coin, rpc: nil, lastGenerate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Node: coinNode, Network: coinNetwork, RpcPort: rpcPort, RpcUser: rpcUser, RpcPassword: rpcPass}
}

func (d *Daemon) logPrefix() string {
	return fmt.Sprintf("[%s/%s] ", d.Coin.Id, d.Network.Id)
}

func (d *Daemon) Stop() error {

	if d.cmd == nil {
		// not started (yet)
		return nil
	}
	// Windows doesn't support Interrupt
	if runtime.GOOS == "windows" {
		_ = d.cmd.Process.Signal(os.Kill)
		return nil
	}

	go func() {
		time.Sleep(15 * time.Second)
		_ = d.cmd.Process.Signal(os.Kill)
	}()
	d.cmd.Process.Signal(os.Interrupt)

	return d.cmd.Wait()
}

func (d *Daemon) GenerateIfNecessary() error {
	if d.Network.Generate {
		if time.Since(d.lastGenerate) > (time.Minute * 10) {
			logging.Debugf("%sGenerating block", d.logPrefix())
			cli, err := d.RpcClient()
			if err != nil {
				return err
			}

			_, err = cli.Generate(1)
			if err != nil {
				return err
			}
			d.lastGenerate = time.Now()
		} else {
			logging.Debugf("%sNot generating block - last generation at %s", d.logPrefix(), d.lastGenerate.String())
		}
	} else {
		logging.Debugf("%sNot generating block - generation disabled", d.logPrefix())
	}
	return nil
}

func (d *Daemon) StartIfNecessary() error {
	start := false
	if d.cmd == nil {
		start = true
	} else {
		if d.cmd.Process == nil {
			d.cmd = nil
			start = true
		} else {
			if d.cmd.ProcessState != nil {
				d.cmd = nil
				start = true
			}
		}
	}

	if start {
		return d.Start()
	}

	return nil
}

func (d *Daemon) Start() error {
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

	logging.Debugf("%sStarted Daemon", d.logPrefix())
	return nil
}

func (d *Daemon) unpackDir() string {
	return path.Join(util.DataDirectory(), "nodes", fmt.Sprintf("unpacked-%s-%s", hex.EncodeToString(d.Node.Hash), d.Network.Id))
}

func (d *Daemon) downloadPath() string {
	return path.Join(util.DataDirectory(), "nodes", hex.EncodeToString(d.Node.Hash))
}

func (d *Daemon) unpackPath() string {
	return path.Join(d.unpackDir(), d.Coin.NodeExecutableName)
}

func (d *Daemon) dataDir() string {
	return path.Join(util.DataDirectory(), "nodedata", d.Coin.Id, d.Network.Id)
}

func (d *Daemon) launch() error {
	unpackPath := d.unpackPath()
	logging.Debugf("%sLaunching daemon", d.logPrefix())

	_, err := os.Stat(d.dataDir())
	if os.IsNotExist(err) {
		err = os.MkdirAll(d.dataDir(), 0700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	params := []string{
		"-printtoconsole",
		fmt.Sprintf("-datadir=%s", d.dataDir()),
		fmt.Sprintf("-port=%d", d.Network.DaemonPort),
		fmt.Sprintf("-rpcuser=%s", d.RpcUser),
		fmt.Sprintf("-rpcpassword=%s", d.RpcPassword),
		fmt.Sprintf("-rpcport=%d", d.RpcPort),
	}

	if d.Network.DaemonParameters != "" {
		params = append(params, d.Network.DaemonParameters)
	}

	d.cmd = exec.Command(unpackPath, params...)
	r, w := io.Pipe()
	go func(d *Daemon, rd io.Reader) {
		br := bufio.NewReader(rd)

		for {
			l, _, e := br.ReadLine()
			if e != nil {
				logging.Debugf("%sError on readline from stdout/err: %s", d.logPrefix(), e.Error())
				return

			}
			// Lob off the datestamp from Daemon's log
			if strings.HasPrefix(string(l), time.Now().Format("2006-01-02")) {
				l = l[20:]
			}

			logging.Debugf("%s%s", d.logPrefix(), l)
		}
	}(d, r)
	d.cmd.Stderr = w
	d.cmd.Stdout = w
	return d.cmd.Start()
}

func (d *Daemon) RpcClient() (*rpcclient.Client, error) {
	if d.rpc == nil {
		var err error
		connCfg := &rpcclient.ConnConfig{
			Host:         fmt.Sprintf("localhost:%d", d.RpcPort),
			User:         d.RpcUser,
			Pass:         d.RpcPassword,
			HTTPPostMode: true,
			DisableTLS:   true,
		}
		d.rpc, err = rpcclient.New(connCfg, nil)
		if err != nil {
			return nil, err
		}
	}
	return d.rpc, nil
}

func (d *Daemon) unpackZip(archive, unpackPath string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == d.Coin.NodeExecutableName {
			logging.Debugf("%sUnpacking %s", d.logPrefix(), d.Coin.NodeExecutableName)
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
			logging.Debugf("%sUnpacked succesfully", d.logPrefix())

			return nil
		}
	}

	err = fmt.Errorf("Could not find a binary named '%s' in the archive.", d.Coin.NodeExecutableName)
	logging.Errorf("%s%s", d.logPrefix(), err.Error())
	return err
}

func (d *Daemon) unpackTar(archive, unpackPath string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if strings.HasSuffix(name, d.Coin.NodeExecutableName) {
				outFile, err := os.OpenFile(unpackPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
				if err != nil {
					return err
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, tarReader)
				if err != nil {
					return err
				}
				logging.Debugf("%sUnpacked succesfully", d.logPrefix())

				return nil
			}
		}
	}

	err = fmt.Errorf("Could not find a binary named '%s' in the archive.", d.Coin.NodeExecutableName)
	logging.Errorf("%s%s", d.logPrefix(), err.Error())
	return err
}

func (d *Daemon) unpack() error {
	unpackDir := d.unpackDir()
	unpackPath := d.unpackPath()

	if _, err := os.Stat(unpackDir); !os.IsNotExist(err) {
		logging.Debugf("%sRemoving unpack directory", d.logPrefix())
		err = os.RemoveAll(unpackDir)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(unpackDir); os.IsNotExist(err) {
		logging.Debugf("%s(Re)creating unpack directory", d.logPrefix())
		err = os.Mkdir(unpackDir, 0755)
		if err != nil {
			return err
		}
	}

	archive := d.downloadPath()
	if strings.HasSuffix(d.Node.Url, ".zip") {
		return d.unpackZip(archive, unpackPath)
	} else if strings.HasSuffix(d.Node.Url, ".tar.gz") {
		return d.unpackTar(archive, unpackPath)
	}

	return fmt.Errorf("Unknown archive format, cannot unpack: %s", d.Node.Url)

}

func (d *Daemon) ensureAvailable() error {
	freshDownload := false
	_ = os.Mkdir(path.Join(util.DataDirectory(), "nodes"), 0700)
	nodePath := d.downloadPath()
	_, err := os.Stat(nodePath)
	if os.IsNotExist(err) {
		logging.Debugf("%sDeamon not found, downloading...", d.logPrefix())
		freshDownload = true
		d.download()
	} else if err != nil {
		return err
	} else {
		logging.Debugf("%sDaemon file already exists", d.logPrefix())
	}

	shaSum, err := util.ShaSum(nodePath)
	if err != nil {
		return err
	}
	if !bytes.Equal(shaSum, d.Node.Hash) {
		logging.Warnf("%sHash differs: [%x] vs [%x]", d.logPrefix(), shaSum, d.Node.Hash)
		if !freshDownload {
			err = os.Remove(nodePath)
			if err != nil {
				return err
			}
			return d.ensureAvailable()
		} else {
			logging.Errorf("%sFreshly downloaded node did not have correct SHA256 hash", d.logPrefix())
		}
	}

	logging.Debugf("%sDaemon file is available and correct", d.logPrefix())
	return nil
}

func (d *Daemon) download() error {
	nodePath := path.Join(util.DataDirectory(), "nodes", hex.EncodeToString(d.Node.Hash))

	resp, err := http.Get(d.Node.Url)
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
	logging.Debugf("%sDaemon file downloaded", d.logPrefix())
	return err
}
