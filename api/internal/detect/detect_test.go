package detect

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJustDetector(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantNames []string
	}{
		{
			"basic recipes",
			"build:\n\tgo build\n\ntest:\n\tgo test\n",
			[]string{"build", "test"},
		},
		{
			"skip assignments",
			"version := \"1.0\"\nbuild:\n\techo ok\n",
			[]string{"build"},
		},
		{
			"silent recipes",
			"@deploy:\n\techo deploy\n",
			[]string{"deploy"},
		},
		{
			"recipes with params",
			"greet(name):\n\techo {{name}}\n",
			[]string{"greet"},
		},
		{
			"skip comments and blanks",
			"# comment\n\nbuild:\n\techo ok\n",
			[]string{"build"},
		},
		{
			"no justfile",
			"",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.content != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "justfile"), []byte(tt.content), 0o644))
			}

			tasks, err := Just{}.Detect(context.Background(), dir)
			require.NoError(t, err)

			if tt.wantNames == nil {
				assert.Nil(t, tasks)
				return
			}

			got := make([]string, len(tasks))
			for i, task := range tasks {
				got[i] = task.Name
				assert.Equal(t, "just", task.Source)
			}
			assert.Equal(t, tt.wantNames, got)
		})
	}
}

func TestJustfileVariants(t *testing.T) {
	for _, name := range []string{"justfile", "Justfile", ".justfile"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("build:\n\techo ok\n"), 0o644))

			tasks, err := Just{}.Detect(context.Background(), dir)
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			assert.Equal(t, "build", tasks[0].Name)
		})
	}
}

func TestDockerfileDetector(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine"), 0o644))

		tasks, err := Dockerfile{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, "build", tasks[0].Name)
		assert.Equal(t, "dockerfile", tasks[0].Source)
		assert.Equal(t, "Dockerfile", tasks[0].RawCommand)
	})

	t.Run("absent", func(t *testing.T) {
		dir := t.TempDir()
		tasks, err := Dockerfile{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		assert.Nil(t, tasks)
	})

	t.Run("is directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(dir, "Dockerfile"), 0o755))

		tasks, err := Dockerfile{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		assert.Nil(t, tasks)
	})
}

func TestGoModDetector(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\ngo 1.21\n"), 0o644))

		tasks, err := GoMod{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		require.Len(t, tasks, 4)

		names := make([]string, len(tasks))
		for i, task := range tasks {
			names[i] = task.Name
			assert.Equal(t, "go", task.Source)
		}
		assert.Equal(t, []string{"build", "fmt", "test", "vet"}, names)
	})

	t.Run("absent", func(t *testing.T) {
		dir := t.TempDir()
		tasks, err := GoMod{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		assert.Nil(t, tasks)
	})
}

func TestCargoDetector(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"foo\"\n"), 0o644))

		tasks, err := Cargo{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		require.Len(t, tasks, 8)

		names := make([]string, len(tasks))
		for i, task := range tasks {
			names[i] = task.Name
			assert.Equal(t, "cargo", task.Source)
		}
		assert.Equal(t, []string{"bench", "build", "check", "clippy", "doc", "fmt", "run", "test"}, names)
	})

	t.Run("absent", func(t *testing.T) {
		dir := t.TempDir()
		tasks, err := Cargo{}.Detect(context.Background(), dir)
		require.NoError(t, err)
		assert.Nil(t, tasks)
	})
}
