package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	yaml "gopkg.in/yaml.v3"

	"github.com/stretchr/testify/require"

	"github.com/715d/unusedfunc/pkg/unusedfunc"
)

// LoaderConfig configures package loading.
type LoaderConfig struct {
	// Dir is the directory to load packages from.
	Dir string

	// BuildTags are build tags to apply.
	BuildTags []string

	// EnableCGo enables CGo support.
	EnableCGo bool

	// GOOS overrides the target operating system.
	GOOS string

	// GOARCH overrides the target architecture.
	GOARCH string
}

// LoadPackages loads packages with the given configuration.
func LoadPackages(t *testing.T, loaderCfg *LoaderConfig) []*packages.Package {
	t.Helper()

	// Build environment.
	env := os.Environ()

	cgoEnabled := "0"
	if loaderCfg.EnableCGo {
		cgoEnabled = "1"
	}
	env = updateEnv(env, "CGO_ENABLED", cgoEnabled)

	if loaderCfg.GOOS != "" {
		env = updateEnv(env, "GOOS", loaderCfg.GOOS)
	}

	if loaderCfg.GOARCH != "" {
		env = updateEnv(env, "GOARCH", loaderCfg.GOARCH)
	}

	t.Logf("Loading packages from %q", loaderCfg.Dir)
	pkgs, err := unusedfunc.LoadPackages(t.Context(), unusedfunc.LoaderOptions{
		Packages:  []string{"./..."},
		BuildTags: loaderCfg.BuildTags,
		Dir:       loaderCfg.Dir,
		Env:       env,
	})
	require.NoError(t, err)
	return pkgs
}

// LoadTestCase loads a test case from a directory with a specified testdata root.
func LoadTestCase(t *testing.T, dir, root string) *TestCase {
	t.Helper()
	yamlPath := filepath.Join(dir, "expected.yaml")

	tc := &TestCase{}
	data, err := os.ReadFile(yamlPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(data, tc)
	require.NoError(t, err)

	// Use relative path from testdata root if provided.
	if root != "" {
		relPath, err := filepath.Rel(root, dir)
		if err != nil {
			tc.Dir = filepath.Base(dir)
		} else {
			tc.Dir = relPath
		}
		return tc
	}

	tc.Dir = filepath.Base(dir)
	return tc
}

// LoadRepositoryPackages loads packages from a git repository.
func LoadRepositoryPackages(t *testing.T, repoCfg *RepoConfig, loaderCfg *LoaderConfig) []*packages.Package {
	t.Helper()

	cloneDir := filepath.Join(t.TempDir(), "repo")
	err := cloneRepository(repoCfg.URL, cloneDir, repoCfg.Ref)
	require.NoError(t, err)

	workDir := cloneDir
	if repoCfg.Subdir != "" {
		workDir = filepath.Join(cloneDir, repoCfg.Subdir)
	}

	// Update loader config to use the repository directory.
	repoLoaderConfig := *loaderCfg
	repoLoaderConfig.Dir = workDir

	// Load packages from the repository.
	return LoadPackages(t, &repoLoaderConfig)
}

// cloneRepository clones a git repository to the specified directory and checks out the given ref
func cloneRepository(url, dir, ref string) error {
	// Use shallow clone with depth 1 to get only the latest commit.
	// This drastically reduces download size and time for large repos.
	args := []string{"clone", "--depth", "1"}

	// If a specific ref is provided, fetch only that ref.
	if ref != "" {
		args = append(args, "--branch", ref)
	}

	// Add --single-branch to avoid fetching other branches.
	args = append(args, "--single-branch", url, dir)

	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmd.String(), err)
	}
	return nil
}

// updateEnv updates or adds an environment variable
func updateEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
