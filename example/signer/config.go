// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package signer

import (
	"fmt"
	"io/ioutil"

	signer "github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	// signer "github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type SignerConfig struct {
	Port    int64                `yaml:"port"`
	Share   string               `yaml:"share"`
	Pubkey  config.Pubkey        `yaml:"pubkey"`
	PartialPubKey map[string]config.PartialPubKey `yaml:"partialPubKey"`
	BKs     map[string]config.BK `yaml:"bks"`
	Message string               `yaml:"msg"`
	Peers   []int64              `yaml:"peers"`
	Threshold uint32			 `yaml:"threshold"`
	SSid 	[]byte				 `yaml:"ssid"`
	AllY    map[string]config.AllY `yaml:"ally"`
	Ped     map[string]config.Ped  `yaml:"ped"`
	Private config.Private       `yaml:"private"`
	PaillierKey config.PaillierKey `yaml:"paillierkey"`
}

type SignerResult struct {
	R string `yaml:"r"`
	S string `yaml:"s"`
}

func readSignerConfigFile(filaPath string) (*SignerConfig, error) {
	c := &SignerConfig{}
	yamlFile, err := ioutil.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func writeSignerResult(id string, result *signer.Result) error {
	signerResult := &SignerResult{
		R: result.R.String(),
		S: result.S.String(),
	}
	err := config.WriteYamlFile(signerResult, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("signer/%s-output.yaml", id)
}
