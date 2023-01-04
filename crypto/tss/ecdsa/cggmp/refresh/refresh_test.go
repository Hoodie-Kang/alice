// Copyright © 2022 AMIS Technologies
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
package refresh

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/getamis/alice/crypto/birkhoffinterpolation"
	"github.com/getamis/alice/crypto/ecpointgrouplaw"
	pt "github.com/getamis/alice/crypto/ecpointgrouplaw"
	"github.com/getamis/alice/crypto/elliptic"
	// "github.com/getamis/alice/crypto/polynomial"
	"github.com/getamis/alice/crypto/tss"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/types"
	"github.com/getamis/alice/types/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestRefresh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Refresh Suite")
}

var (
	// Assume that threshold <= totalParty
	// secret    = big.NewInt(21313513)
	// publicKey = pt.ScalarBaseMult(elliptic.Secp256k1(), secret)
	// for testing the DKG result -> Refresh
	sh = "4433930311896586173225605911604098714764809986765287553545688226619194396795"
	sh2= "11206683307503035256881204936677301873389877738402403164108038906992854558074"
	X = "112363345206559513967411659444620105992871732728134049791505159850234274012554"
	Y = "90435951029605018774762791226409664020987295549738711287286651149053644515432"
	//bks - x, rank=0
	a = "35708105979892983394462002592936921792160460602972875353936833272812615304178"
	b = "69597477371637251385786663792575032001854147985881576910239092372529440542030"
)
func dkgresult() (*dkg.Result) {
	x, _ := new(big.Int).SetString(X, 10)

	y, _ := new(big.Int).SetString(Y, 10)

	pubkey, _ := ecpointgrouplaw.NewECPoint(elliptic.Secp256k1(), x, y)


	dkgResult := &dkg.Result{
		PublicKey: pubkey,
		Bks:       make(map[string]*birkhoffinterpolation.BkParameter, 2),
	}
	bka, _ := new(big.Int).SetString(a, 10)

	bkb, _ := new(big.Int).SetString(b, 10)

	dkgResult.Bks["id-0"] = birkhoffinterpolation.NewBkParameter(bka, 0)
	dkgResult.Bks["id-1"] = birkhoffinterpolation.NewBkParameter(bkb, 0)
	return dkgResult
}

var _ = Describe("Refresh", func() {
	DescribeTable("NewDKG()", func(threshold uint32, ranks []uint32) {
		totalParty := len(ranks)
		dkgResult := dkgresult()
		publicKey := dkgResult.PublicKey
		shares := make([]*big.Int, 2)
		shares[0], _ = new(big.Int).SetString(sh, 10)
		shares[1], _ = new(big.Int).SetString(sh2, 10)
		bksSlice := make([]*birkhoffinterpolation.BkParameter, 2)
		bksSlice[0] = dkgResult.Bks["id-0"]
		bksSlice[1] = dkgResult.Bks["id-1"]
		curve := publicKey.GetCurve()
		fieldOrder := curve.Params().N
		// shares, bksSlice, err := gernerateShare(curve, threshold, ranks)
		// Expect(err).Should(BeNil())

		// new peer managers and dkgs
		refreshes, bks, listeners := newRefreshes(threshold, totalParty, shares, bksSlice)
		for _, l := range listeners {
			l.On("OnStateChanged", types.StateInit, types.StateDone).Once()
		}
		for _, d := range refreshes {
			d.Start()
		}
		time.Sleep(2 * time.Second)
		for _, l := range listeners {
			l.AssertExpectations(GinkgoT())
		}

		// Set new shares
		afterShares := make([]*big.Int, len(shares))
		afterPartialRefreshPubKeys := make([]*pt.ECPoint, len(shares))

		r0, err := refreshes[tss.GetTestID(0)].GetResult()
		Expect(err).Should(BeNil())
		for i := 0; i < len(afterShares); i++ {
			r, err := refreshes[tss.GetTestID(i)].GetResult()
			Expect(err).Should(BeNil())
			afterShares[i] = r.RefreshShare
			afterPartialRefreshPubKeys[i] = r0.RefreshPartialPubKey[tss.GetTestID(i)]
		}
		// check that all refresh partial public keys, Y, pedParameters are all the same.
		for i := 1; i < len(shares); i++ {
			r, err := refreshes[tss.GetTestID(i)].GetResult()
			Expect(err).Should(BeNil())
			for k, v := range r0.RefreshPartialPubKey {
				Expect(v.Equal(r.RefreshPartialPubKey[k])).Should(BeTrue())
			}
			for k, v := range r0.Y {
				Expect(v.Equal(r.Y[k])).Should(BeTrue())
			}
			for k, v := range r0.PedParameter {
				Expect(v.Getn().Cmp(r.PedParameter[k].Getn()) == 0).Should(BeTrue())
				Expect(v.Gets().Cmp(r.PedParameter[k].Gets()) == 0).Should(BeTrue())
				Expect(v.Gett().Cmp(r.PedParameter[k].Gett()) == 0).Should(BeTrue())
			}
		}

		// check all paillier keys work by comparing the same "N".
		for i := 0; i < len(shares); i++ {
			r, err := refreshes[tss.GetTestID(i)].GetResult()
			Expect(err).Should(BeNil())
			otherIndex := (i + 1) % len(shares)
			rpai, err := refreshes[tss.GetTestID(otherIndex)].GetResult()
			Expect(err).Should(BeNil())
			Expect(r.RefreshPaillierKey.GetN().Cmp(rpai.PedParameter[tss.GetTestID(i)].Getn()) == 0).Should(BeTrue())
		}

		allBks := make(birkhoffinterpolation.BkParameters, len(shares))
		for i := 0; i < len(allBks); i++ {
			allBks[i] = bks[tss.GetTestID(i)]
		}
		bkcoefficient, err := allBks.ComputeBkCoefficient(threshold, fieldOrder)
		Expect(err).Should(BeNil())
		gotSecret := new(big.Int).Mul(afterShares[0], bkcoefficient[0])
		gotSecret.Mod(gotSecret, fieldOrder)
		gotPubKey := afterPartialRefreshPubKeys[0].ScalarMult(bkcoefficient[0])
		for i := 1; i < len(afterShares); i++ {
			gotSecret.Add(gotSecret, new(big.Int).Mul(afterShares[i], bkcoefficient[i]))
			gotSecret.Mod(gotSecret, fieldOrder)
			gotPubKey, err = gotPubKey.Add(afterPartialRefreshPubKeys[i].ScalarMult(bkcoefficient[i]))
			Expect(err).Should(BeNil())
		}
		// Check all partial public keys are correct.
		// Expect(gotSecret.Cmp(secret) == 0).Should(BeTrue())
		Expect(gotPubKey.Equal(publicKey)).Should(BeTrue())
	},
		Entry("Case #0", uint32(2),
			[]uint32{
				0, 0,
			},
		),
		// Entry("Case #1", uint32(2),
		// 	[]uint32{
		// 		0, 1, 1,
		// 	},
		// ),
		// Entry("Case #2", uint32(3),
		// 	[]uint32{
		// 		0, 1, 2,
		// 	},
		// ),
	)
})

func newRefreshes(threshold uint32, totalParty int, shareSlice []*big.Int, bksPara []*birkhoffinterpolation.BkParameter) (map[string]*Refresh, map[string]*birkhoffinterpolation.BkParameter, map[string]*mocks.StateChangedListener) {
	publicKey := dkgresult().PublicKey
	lens := totalParty
	refreshes := make(map[string]*Refresh, lens)
	refreshesMain := make(map[string]types.MessageMain, lens)
	peerManagers := make([]types.PeerManager, lens)
	listeners := make(map[string]*mocks.StateChangedListener, lens)
	bks := make(map[string]*birkhoffinterpolation.BkParameter)
	share := make(map[string]*big.Int)
	partialPubKey := make(map[string]*pt.ECPoint)
	for i := 0; i < totalParty; i++ {
		bks[tss.GetTestID(i)] = bksPara[i]
		share[tss.GetTestID(i)] = shareSlice[i]
		partialPubKey[tss.GetTestID(i)] = pt.ScalarBaseMult(publicKey.GetCurve(), shareSlice[i])
	}
	fmt.Println("test-partialPubKey",partialPubKey)
	keySize := 2048
	ssidInfo := []byte("A")
	for i := 0; i < lens; i++ {
		id := tss.GetTestID(i)
		pm := tss.NewTestPeerManager(i, lens)
		pm.Set(refreshesMain)
		peerManagers[i] = pm
		listeners[id] = new(mocks.StateChangedListener)
		var err error
		refreshes[id], err = NewRefresh(share[id], publicKey, peerManagers[i], threshold, partialPubKey, bks, keySize, ssidInfo, listeners[id])
		Expect(err).Should(BeNil())
		refreshesMain[id] = refreshes[id]
		r, err := refreshes[id].GetResult()
		Expect(r).Should(BeNil())
		Expect(err).Should(Equal(tss.ErrNotReady))
	}
	return refreshes, bks, listeners
}

// func gernerateShare(curve elliptic.Curve, threshold uint32, ranks []uint32) ([]*big.Int, []*birkhoffinterpolation.BkParameter, error) {
// 	totalParty := len(ranks)
// 	poly, err := polynomial.RandomPolynomial(curve.Params().N, threshold-1)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	poly.SetConstant(secret)
// 	share := make([]*big.Int, totalParty)
// 	bk := make([]*birkhoffinterpolation.BkParameter, totalParty)
// 	for i := 0; i < len(share); i++ {
// 		tempPoly := poly.Differentiate(ranks[i])
// 		share[i] = tempPoly.Evaluate(big.NewInt(int64(i) + 1))
// 		bk[i] = birkhoffinterpolation.NewBkParameter(big.NewInt(int64(i)+1), ranks[i])
// 	}
// 	return share, bk, nil
// }