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

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/crypto/birkhoffinterpolation"
	"github.com/getamis/alice/crypto/ecpointgrouplaw"
	"github.com/getamis/alice/crypto/elliptic"
	"github.com/getamis/alice/crypto/homo/paillier"
	paillierzkproof "github.com/getamis/alice/crypto/zkproof/paillier"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
)

// For sign input
type SignInput struct {
	PublicKey     *ecpointgrouplaw.ECPoint
	Share         *big.Int
	Bks           map[string]*birkhoffinterpolation.BkParameter
	PartialPubKey map[string]*ecpointgrouplaw.ECPoint
	PedParameter  map[string]*paillierzkproof.PederssenOpenParameter
	PaillierKey   *paillier.Paillier
}

var (
	// ErrConversion for big int conversion error
	ErrConversion = errors.New("conversion error")
)

// GetPeerIDFromPort gets peer ID from port.
func GetPeerIDFromPort(port int64) string {
	// For convenience, we set peer ID as "id-" + port
	// 지금은 sign 에서만 필요 .. refresh 수정필요
	// if port % 2 == 1 {
	// 	return "Octet"	
	// } else if port % 2 == 0 {
	// 	return "User"
	// }
	return fmt.Sprintf("id-%d", port)
}

// GetCurve returns the curve we used in this example.
func GetCurve() elliptic.Curve {
	// For simplicity, we use S256 curve.
	return elliptic.Secp256k1()
}

// ConvertDKGResult converts DKG result from config.
func ConvertDKGResult(cfgPubkey config.Pubkey, cfgShare string, cfgBKs map[string]config.BK, cfgPPK map[string]config.ECPoint) (*dkg.Result, error) {
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
	for peerID, ppk := range cfgPPK {
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
		dkgResult.PartialPubKey[peerID] = ppkey
	}
	return dkgResult, nil
}

// ConvertSignInput converts SingInput(=DKG&Refresh result) from config.
func ConvertSignInput(cfgShare string, cfgPubkey config.Pubkey, cfgPPK map[string]config.ECPoint, cfgPriv config.PaillierKey, cfgPed map[string]config.Ped, cfgBKs map[string]config.BK) (*SignInput, error) {
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
	
	signInput := &SignInput{
		PublicKey:     pubkey,
		Share:         share,
		Bks:           make(map[string]*birkhoffinterpolation.BkParameter),
		PartialPubKey: make(map[string]*ecpointgrouplaw.ECPoint),
		PedParameter:  make(map[string]*paillierzkproof.PederssenOpenParameter),
		PaillierKey:   paillierKey,
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
	for peerID, ppk := range cfgPPK {
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
		signInput.PartialPubKey[peerID] = ppkey
	}

	// Build PedParameter
	for peerID, pedpara := range cfgPed {
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
		signInput.PedParameter[peerID] = pedparams
	}

	return signInput, nil
}
