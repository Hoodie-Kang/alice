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
package utils

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/getamis/alice/crypto/birkhoffinterpolation"
	"github.com/getamis/alice/crypto/ecpointgrouplaw"
	"github.com/getamis/alice/crypto/elliptic"
	"github.com/getamis/alice/crypto/homo/paillier"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	paillierzkproof "github.com/getamis/alice/crypto/zkproof/paillier"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
)

// For sign input
// ped 사용해서 paillierkey 만드는 방식으로 결과를 만드는중 -- 수정 필요!!
type SignInput struct {
	PublicKey     *ecpointgrouplaw.ECPoint
	Share         *big.Int
	Bks           map[string]*birkhoffinterpolation.BkParameter
	PartialPubKey map[string]*ecpointgrouplaw.ECPoint
	Y             map[string]*ecpointgrouplaw.ECPoint
	PedParameter  map[string]*paillierzkproof.PederssenOpenParameter
	PaillierKey   *paillier.Paillier
	YSecret       *big.Int
}

var (
	// ErrConversion for big int conversion error
	ErrConversion = errors.New("conversion error")
)

// GetPeerIDFromPort gets peer ID from port.
func GetPeerIDFromPort(port int64) string {
	// For convenience, we set peer ID as "id-" + port
	return fmt.Sprintf("id-%d", port)
}

// GetCurve returns the curve we used in this example.
func GetCurve() elliptic.Curve {
	// For simplicity, we use S256 curve.
	return elliptic.Secp256k1()
}

// ConvertDKGResult converts DKG result from config.
func ConvertDKGResult(cfgPubkey config.Pubkey, cfgShare string, cfgBKs map[string]config.BK, cfgPPK map[string]config.PartialPubKey) (*dkg.Result, error) {
	// Build public key.
	x, ok := new(big.Int).SetString(cfgPubkey.X, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "x", cfgPubkey.X)
		return nil, ErrConversion
	}
	y, ok := new(big.Int).SetString(cfgPubkey.Y, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "y", cfgPubkey.Y)
		return nil, ErrConversion
	}
	pubkey, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
	if err != nil {
		log.Error("Cannot get public key", "err", err)
		return nil, err
	}

	// Build share.
	share, ok := new(big.Int).SetString(cfgShare, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "share", share)
		return nil, ErrConversion
	}

	dkgResult := &dkg.Result{
		PublicKey: pubkey,
		Share:     share,
		Bks:       make(map[string]*birkhoffinterpolation.BkParameter),
		PartialPubKey: make(map[string]*ecpointgrouplaw.ECPoint),
	}

	// Build bks.
	for peerID, bk := range cfgBKs {
		x, ok := new(big.Int).SetString(bk.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", bk.X)
			return nil, ErrConversion
		}
		dkgResult.Bks[peerID] = birkhoffinterpolation.NewBkParameter(x, bk.Rank)
	}

	// Build PartialPubKey.
	for peerId, ppk := range cfgPPK {
		x, ok := new(big.Int).SetString(ppk.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", ppk.X)
			return nil, ErrConversion
		}
		y, ok := new(big.Int).SetString(ppk.Y, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "y", ppk.Y)
			return nil, ErrConversion
		}
		ppkey, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
		if err != nil {
			log.Error("Cannot get partial public key", "err", err)
			return nil, err
		}
		dkgResult.PartialPubKey[peerId] = ppkey
	}
	return dkgResult, nil
}
// ConvertSignInput converts SingInput(=DKG&Refresh result) from config.
// paillierKey *paillier.Paillier 를 직접 받아오지 못하기 때문에, 일단 PedPara 즉, p q 값을 가져와서 paillierkey 를 만들어서 사용함.
// -> 이 방식은 private key를 드러내는 위험한 방식이므로 테스트 후 key가 드러나지 않게 가져오는 방법으로 반드시 수정해야함.
func ConvertSignInput(cfgShare string, cfgPubkey config.Pubkey, cfgPPK map[string]config.PartialPubKey, cfgAllYs map[string]config.AllY, cfgPriv config.Private, cfgPed map[string]config.Ped, cfgBKs map[string]config.BK, cfgYSec string) (*SignInput, error) {
	// Build public key.
	x, ok := new(big.Int).SetString(cfgPubkey.X, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "x", cfgPubkey.X)
		return nil, ErrConversion
	}
	y, ok := new(big.Int).SetString(cfgPubkey.Y, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "y", cfgPubkey.Y)
		return nil, ErrConversion
	}
	pubkey, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
	if err != nil {
		log.Error("Cannot get public key", "err", err)
		return nil, err
	}

	// Build share.
	share, ok := new(big.Int).SetString(cfgShare, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "share", share)
		return nil, ErrConversion
	}
	// ped
	p, _ := new(big.Int).SetString(cfgPriv.P, 10)
	q, _ := new(big.Int).SetString(cfgPriv.Q, 10)
	// build paillierkey using ped-> p, q
	paillierKey, _ := paillier.NewPaillierWithGivenPrimes(p, q)
	// build ysecret
	ysec, _ := new(big.Int).SetString(cfgYSec, 10)
	
	signInput := &SignInput{
		PublicKey:     pubkey,
		Share:         share,
		Bks:           make(map[string]*birkhoffinterpolation.BkParameter),
		PartialPubKey: make(map[string]*ecpointgrouplaw.ECPoint),
		Y:             make(map[string]*ecpointgrouplaw.ECPoint),
		PedParameter:  make(map[string]*paillierzkproof.PederssenOpenParameter),
		// for testing!! - private key
		PaillierKey:   paillierKey,
		YSecret:       ysec,
	}

	// Build bks.
	for peerID, bk := range cfgBKs {
		x, ok := new(big.Int).SetString(bk.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", bk.X)
			return nil, ErrConversion
		}
		signInput.Bks[peerID] = birkhoffinterpolation.NewBkParameter(x, bk.Rank)
	}

	// Build PartialPubKey.
	for peerId, ppk := range cfgPPK {
		x, ok := new(big.Int).SetString(ppk.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", ppk.X)
			return nil, ErrConversion
		}
		y, ok := new(big.Int).SetString(ppk.Y, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "y", ppk.Y)
			return nil, ErrConversion
		}
		ppkey, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
		if err != nil {
			log.Error("Cannot get partial public key", "err", err)
			return nil, err
		}
		signInput.PartialPubKey[peerId] = ppkey
	}
	// Build All Y.
	for peerId, ally := range cfgAllYs {
		x, ok := new(big.Int).SetString(ally.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", ally.X)
			return nil, ErrConversion
		}
		y, ok := new(big.Int).SetString(ally.Y, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "y", ally.Y)
			return nil, ErrConversion
		}
		allY, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
		if err != nil {
			log.Error("Cannot get partial public key", "err", err)
			return nil, err
		}
		signInput.Y[peerId] = allY
	}

	// Build PedParameter
	for peerId, pedpara := range cfgPed {
		n, ok := new(big.Int).SetString(pedpara.N, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "n", pedpara.N)
			return nil, ErrConversion
		}
		s, ok := new(big.Int).SetString(pedpara.S, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "s", pedpara.S)
			return nil, ErrConversion
		}
		t, ok := new(big.Int).SetString(pedpara.T, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "t", pedpara.T)
			return nil, ErrConversion
		}
		pedparams := paillierzkproof.NewPedersenOpenParameter(n, s, t)
		signInput.PedParameter[peerId] = pedparams
	}
	return signInput, nil
}
