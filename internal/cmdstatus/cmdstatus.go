// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmdstatus

import (
	"context"
	"os"

	"github.com/GoogleContainerTools/kpt/internal/docs/generated/livedocs"
	"github.com/GoogleContainerTools/kpt/internal/util/argutil"
	"github.com/GoogleContainerTools/kpt/pkg/live"
	kptstatus "github.com/GoogleContainerTools/kpt/pkg/status"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/cmd/status"
	"sigs.k8s.io/cli-utils/pkg/apply/poller"
	"sigs.k8s.io/cli-utils/pkg/inventory"
)

func NewRunner(ctx context.Context, factory util.Factory,
	invFactory inventory.ClientFactory, loader status.Loader) *status.Runner {
	r := status.GetRunner(ctx, factory, invFactory, loader)
	r.PollerFactoryFunc = pollerFactoryFunc
	r.Command.Use = "status [PKG_PATH | -]"
	r.Command.Short = livedocs.StatusShort
	r.Command.Long = livedocs.StatusShort + "\n" + livedocs.StatusLong
	r.Command.Example = livedocs.StatusExamples
	return r
}

func NewCommand(ctx context.Context, factory util.Factory,
	invFactory inventory.ClientFactory, loader status.Loader) *cobra.Command {
	return NewRunner(ctx, factory, invFactory, loader).Command
}

func pollerFactoryFunc(f util.Factory) (poller.Poller, error) {
	return kptstatus.NewStatusPoller(f)
}

type RGInventoryLoader struct {
	factory util.Factory
	ctx     context.Context
}

func NewRGInventoryLoader(ctx context.Context, factory util.Factory) *RGInventoryLoader {
	return &RGInventoryLoader{
		factory: factory,
		ctx:     ctx,
	}
}

func (rir *RGInventoryLoader) GetInvInfo(cmd *cobra.Command, args []string) (inventory.Info, error) {
	if len(args) == 0 {
		// default to the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		args = append(args, cwd)
	}

	path := args[0]
	var err error
	if args[0] != "-" {
		path, err = argutil.ResolveSymlink(rir.ctx, path)
		if err != nil {
			return nil, err
		}
	}

	_, inv, err := live.Load(rir.factory, path, cmd.InOrStdin())
	if err != nil {
		return nil, err
	}

	invInfo, err := live.ToInventoryInfo(inv)
	if err != nil {
		return nil, err
	}

	return invInfo, nil
}
