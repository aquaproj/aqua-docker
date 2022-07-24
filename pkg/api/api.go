package api

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dirPermission os.FileMode = 0o755
	binPermission os.FileMode = 0o755
)

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

type Param struct {
	AquaVersion     string
	Dest            string
	Config          string
	Bins            []string
	AquaInstallPath string
}

const helpMessage = `aqua-docker - Install CLIs in Docker images

https://github.com/aquaproj/aqua-docker

Usage:
	$ aqua-docker [--aqua-version latest] [--dest dist] [--config aqua.yaml] command [command, ...]

Options:
	--help          show this help message
	--version       show aqua-docker version
	--aqua-version  aqua version
	--dest          directory file path where commands are copied
	--config        aqua configuration file path`

func Run(ldflags *LDFlags) error {
	ctx := context.Background()

	param := &Param{}
	flag.StringVar(&param.AquaVersion, "aqua-version", "latest", "aqua version")
	flag.StringVar(&param.Dest, "dest", "dist", "directory file path where commands are copied")
	flag.StringVar(&param.Config, "config", "aqua.yaml", "aqua configuration file path")
	help := false
	versionFlag := false
	flag.BoolVar(&help, "help", false, "show this help")
	flag.BoolVar(&versionFlag, "version", false, "show aqua-docker version")
	flag.Parse()
	param.Bins = flag.Args()
	if help {
		fmt.Fprintln(os.Stderr, helpMessage)
		return nil
	}

	if versionFlag {
		fmt.Fprintf(os.Stderr, "%s (%s)", ldflags.Version, ldflags.Commit)
		return nil
	}

	log.Println("[INFO] Installing aqua")
	aquaFile, err := os.CreateTemp("", "aqua")
	if err != nil {
		return fmt.Errorf("create a temporal file to install aqua: %w", err)
	}
	defer aquaFile.Close()
	param.AquaInstallPath = aquaFile.Name()
	if err := os.Chmod(param.AquaInstallPath, binPermission); err != nil {
		return fmt.Errorf("change aqua's file permission: %w", err)
	}

	if err := installAqua(ctx, param); err != nil {
		return fmt.Errorf("install aqua: %w", err)
	}
	if err := command(ctx, param.AquaInstallPath, "-c", param.Config, "i"); err != nil {
		return fmt.Errorf("aqua i: %w", err)
	}
	log.Println("[INFO] Creating a directory")
	if err := os.MkdirAll(param.Dest, dirPermission); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	log.Println("[INFO] Copying files")
	for _, bin := range param.Bins {
		if err := copyFile(ctx, param.AquaInstallPath, filepath.Join(param.Dest, bin), bin); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(ctx context.Context, aquaInstallPath, dest, bin string) error {
	p, err := aquaWhich(ctx, aquaInstallPath, bin)
	if err != nil {
		return fmt.Errorf("aqua which %s: %w", bin, err)
	}
	src, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("open a file %s: %w", p, err)
	}
	defer src.Close()

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", dest, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("copy a file dest=%s, src=%s: %w", dest, p, err)
	}
	if err := os.Chmod(dest, binPermission); err != nil {
		return fmt.Errorf("change a file permission %s: %w", dest, err)
	}
	return nil
}

func installAqua(ctx context.Context, param *Param) error {
	u := fmt.Sprintf("https://github.com/aquaproj/aqua/releases/%s/download/aqua_%s_%s.tar.gz", param.AquaVersion, runtime.GOOS, runtime.GOARCH)
	log.Printf("Downloading %s", u)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("create a HTTP request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	f, err := os.Create(param.AquaInstallPath)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", param.AquaInstallPath, err)
	}
	defer f.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read a response body: %w", err)
		}
		return fmt.Errorf("download aqua but status code >= 400: status_code=%d, response_body=%s", resp.StatusCode, string(b))
	}
	if err := unarchive(f, resp.Body); err != nil {
		return fmt.Errorf("downloand and unarchive aqua: %w", err)
	}
	return nil
}

func command(ctx context.Context, cmdName string, args ...string) error {
	s := cmdName + " " + strings.Join(args, " ")
	fmt.Fprintln(os.Stderr, "+ "+s)
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute a command: %s: %w", s, err)
	}
	return nil
}

func aquaWhich(ctx context.Context, aquaInstallPath, bin string) (string, error) {
	s := "aqua which " + bin
	fmt.Fprintln(os.Stderr, "+ "+s)
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, aquaInstallPath, "which", bin)
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute a command: %s: %w", s, err)
	}
	return strings.TrimSpace(buf.String()), nil
}
