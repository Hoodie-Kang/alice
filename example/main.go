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
package main

import (
	"C"
	"fmt"

	// "github.com/getamis/alice/example/dkg"
	// "github.com/getamis/alice/example/refresh"
	// "github.com/getamis/alice/example/sign"
	"github.com/getamis/alice/example/signSix"
	"github.com/getamis/alice/example/bip32/master"
	"github.com/getamis/alice/example/bip32/child"
)

//export Tss
func Tss(fun *C.char, argv *C.char, msg *C.char) {
	protocol := C.GoString(fun)
	path := C.GoString(argv)

	switch protocol {
	case "dkg":
		// dkg.Dkg(path)
	case "refresh":
		// refresh.Refresh(path)
	case "sign":
		// message := C.GoString(msg)
		// sign.Sign(path, message)
	case "signSix":
		message := C.GoString(msg)
		signSix.SignSix(path, message)
	case "bip32master":
		master.Bip32Master(path)
	case "bip32child":
		child.Bip32Child(path)
	default:
		fmt.Println("Retry: protocol match error")
	}
}

// func main() {
// 	fmt.Println("sadf")
// }