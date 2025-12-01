package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ecos-labs/ecos/code/cli/plugins/types"
	"github.com/ecos-labs/ecos/code/cli/plugins/types/mocks"
	"github.com/spf13/cobra"
	"go.uber.org/mock/gomock"
)

func writeTestConfig(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ".ecos.yaml")
	err := os.WriteFile(path, []byte(`
project_name: test
data_source: aws_cur
aws:
  region: us-east-1
  results_bucket: test-bucket
  dbt_workgroup: test-dbt
`), 0o600)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

func TestRunDestroy_ConfigMissing(t *testing.T) {
	tmp := t.TempDir()
	// change working dir to an empty temp dir (no .ecos.yaml)
	t.Chdir(tmp)

	cmd := &cobra.Command{}
	var buf strings.Builder
	// use pointer to strings.Builder which implements io.Writer
	cmd.SetErr(&buf)

	err := runDestroy(cmd, nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ecos configuration file not found") {
		t.Fatalf("expected missing config error, got: %s", out)
	}
}

func TestRunDestroy_LoadFromConfigFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	// Expect LoadFromConfig to be called and return a load error
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(errors.New("bad cfg"))

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetErr(&buf)

	err := runDestroy(cmd, nil)
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	if !strings.Contains(buf.String(), "bad cfg") {
		t.Fatalf("expected printed load error, got: %s", buf.String())
	}
}

func TestRunDestroy_ValidatePrereqFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(nil)
	mockPlugin.EXPECT().ValidatePrerequisites().Return(errors.New("prereq fail"))

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}

	err := runDestroy(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "prerequisite validation failed") {
		t.Fatalf("expected prereq error, got: %v", err)
	}
}

func TestRunDestroy_UserCancels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	cfgPath := writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(nil)
	mockPlugin.EXPECT().ValidatePrerequisites().Return(nil)
	mockPlugin.EXPECT().DescribeDestruction().Return([]types.DestroyResourcePreview{
		{Kind: "S3 Bucket", Name: "bucket", Managed: true},
	})

	// override confirm to simulate user cancelling
	oldConfirm := utilsConfirmPrompt
	utilsConfirmPrompt = func(_ string) bool { return false }
	defer func() { utilsConfirmPrompt = oldConfirm }()

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}
	err := runDestroy(cmd, nil)
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	// check that no backup was created when user cancels (before confirmation)
	if _, statErr := os.Stat(cfgPath + ".backup"); statErr == nil {
		t.Fatalf("expected no backup file when user cancels, but backup exists")
	}
	// check original config file still exists (not renamed to backup)
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		t.Fatalf("expected original config file to still exist after cancellation: %v", statErr)
	}
}

func TestRunDestroy_DestroySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	cfgPath := writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(nil)
	mockPlugin.EXPECT().ValidatePrerequisites().Return(nil)
	mockPlugin.EXPECT().DescribeDestruction().Return([]types.DestroyResourcePreview{
		{Kind: "S3 Bucket", Name: "bucket", Managed: true},
	})
	mockPlugin.EXPECT().DestroyResources().Return([]types.DestroyResourceResult{
		{Kind: "S3 Bucket", Name: "bucket", Status: types.DestroyStatusDeleted},
	}, nil)

	// override confirm to simulate user agreeing
	oldConfirm := utilsConfirmPrompt
	utilsConfirmPrompt = func(_ string) bool { return true }
	defer func() { utilsConfirmPrompt = oldConfirm }()

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}
	err := runDestroy(cmd, nil)
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	// check backup file was created and kept after successful destruction
	if _, statErr := os.Stat(cfgPath + ".backup"); statErr != nil {
		t.Fatalf("expected backup file to exist and be kept after successful destruction: %v", statErr)
	}
	// check original config file no longer exists (it was renamed to backup and kept permanently)
	if _, statErr := os.Stat(cfgPath); statErr == nil {
		t.Fatalf("expected original config file to be renamed to backup")
	}
}

func TestRunDestroy_DestroyFailsConfigRestored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	cfgPath := writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(nil)
	mockPlugin.EXPECT().ValidatePrerequisites().Return(nil)
	mockPlugin.EXPECT().DescribeDestruction().Return([]types.DestroyResourcePreview{
		{Kind: "S3 Bucket", Name: "bucket", Managed: true},
	})
	mockPlugin.EXPECT().DestroyResources().Return(nil, errors.New("destruction failed"))

	// override confirm to simulate user agreeing
	oldConfirm := utilsConfirmPrompt
	utilsConfirmPrompt = func(_ string) bool { return true }
	defer func() { utilsConfirmPrompt = oldConfirm }()

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}
	err := runDestroy(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "resource destruction failed") {
		t.Fatalf("expected destruction error, got: %v", err)
	}

	// check original config file was restored after failure
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		t.Fatalf("expected original config file to be restored: %v", statErr)
	}
	// check backup file no longer exists (it was restored)
	if _, statErr := os.Stat(cfgPath + ".backup"); statErr == nil {
		t.Fatalf("expected backup file to be restored back to original")
	}
}

func TestRunDestroy_DestroyCancelledDuringDestruction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	t.Chdir(tmp)
	cfgPath := writeTestConfig(t, tmp)

	mockPlugin := mocks.NewMockConfigurableDestroyPlugin(ctrl)
	mockPlugin.EXPECT().LoadFromConfig(gomock.Any()).Return(nil)
	mockPlugin.EXPECT().ValidatePrerequisites().Return(nil)
	mockPlugin.EXPECT().DescribeDestruction().Return([]types.DestroyResourcePreview{
		{Kind: "S3 Bucket", Name: "bucket", Managed: true},
	})
	// Simulate user cancelling during destruction (returns nil, nil)
	mockPlugin.EXPECT().DestroyResources().Return(nil, nil)

	// override confirm to simulate user agreeing initially
	oldConfirm := utilsConfirmPrompt
	utilsConfirmPrompt = func(_ string) bool { return true }
	defer func() { utilsConfirmPrompt = oldConfirm }()

	// override loader
	oldLoader := registryLoadDestroy
	registryLoadDestroy = func(_ string) (types.DestroyPlugin, error) {
		return mockPlugin, nil
	}
	defer func() { registryLoadDestroy = oldLoader }()

	cmd := &cobra.Command{}
	err := runDestroy(cmd, nil)
	if err != nil {
		t.Fatalf("expected nil error when user cancels during destruction, got: %v", err)
	}

	// check original config file was restored after cancellation
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		t.Fatalf("expected original config file to be restored after cancellation: %v", statErr)
	}
	// check backup file no longer exists (it was restored)
	if _, statErr := os.Stat(cfgPath + ".backup"); statErr == nil {
		t.Fatalf("expected backup file to be restored back to original after cancellation")
	}
}
