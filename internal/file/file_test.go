package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFile(t *testing.T) {

	// Create a test file so we can test against it
	tempFile := filepath.Join(t.TempDir(), "test_file.txt")
	err := os.WriteFile(tempFile, []byte("test file"), 0644)
	assert.NoError(t, err)

	t.Run("should open, read, and close file valid file.", func(t *testing.T) {
		handle, err := NewFile(tempFile)
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		content := make([]byte, 9)
		n, err := handle.Read(content)
		assert.NoError(t, err)
		assert.Equal(t, "test file", string(content[:n]))
		closeErr := handle.Close()
		assert.NoError(t, closeErr)
	})

	t.Run("should not find file", func(t *testing.T) {
		_, err := NewFile("non_exitent_file.alex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	// Remove temp file
	os.Remove(tempFile)
}

func TestCreateOutputFile(t *testing.T) {
	t.Run("should create and write file", func(t *testing.T) {
		// Create a new output file
		outputPath := filepath.Join(t.TempDir(), "output.txt")
		handle, err := CreateOutputFile(outputPath)
		assert.NoError(t, err)
		assert.NotNil(t, handle)

		// Write data to the file
		_, err = handle.Write([]byte("output content"))
		assert.NoError(t, err)

		// Close the file
		err = handle.Close()
		assert.NoError(t, err)

		// Ensure the file exists with correct content
		content, err := os.ReadFile(outputPath)
		assert.NoError(t, err)
		assert.Equal(t, "output content", string(content))
	})

	t.Run("should file not found", func(t *testing.T) {
		// Create a new output file
		outputPath := "/some_root_folder_that_does_not_exist/do_not_create_me.pgtdm"
		_, err := CreateOutputFile(outputPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestCloseHandle(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(t.TempDir(), "test_file.txt")
	handle, err := os.Create(tempFile)
	assert.NoError(t, err)

	// Close the file using CloseHandle
	err = CloseHandle(handle)
	assert.NoError(t, err)

	// Try to close the file again, should return an error
	err = CloseHandle(handle)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file already closed")
}
