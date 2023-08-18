// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"os"

	"github.com/getamis/alice/example/dkg"
	"github.com/getamis/alice/example/refresh"
	"github.com/getamis/alice/example/sign"
	"github.com/getamis/alice/example/signSix"
	"github.com/getamis/alice/example/bip32/master"
	"github.com/getamis/alice/example/bip32/child"
	"github.com/getamis/alice/example/signer"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "tss-example",
	Short: "TSS example",
	Long:  `This is a tss-cggmp example`,
}

func init() {
	cmd.AddCommand(dkg.Cmd)
	cmd.AddCommand(refresh.Cmd)
	cmd.AddCommand(sign.Cmd)
	cmd.AddCommand(signSix.Cmd)
	// to support 2 party bip32
	cmd.AddCommand(master.Cmd)
	cmd.AddCommand(child.Cmd)
	// testing GG18 signer for bip32
	cmd.AddCommand(signer.Cmd)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
